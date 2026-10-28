package travel

import (
	"context"
	sessionv1 "github.com/godnota/pkg/generated/session/v1"
	travelv1 "github.com/godnota/pkg/generated/travel/v1"
	"testing"

	commonv1 "github.com/godnota/pkg/generated/common/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestResolvedRetainedByCityAndHiddenFromAI(t *testing.T) {
	request, _ := structpb.NewStruct(map[string]any{"schema_version": "1.0", "locale": "ru", "currency": "RUB", "origin": map[string]any{"city": "Москва"}, "destination": map[string]any{"city": "Париж"}, "dates": map[string]any{"departure_date": "2026-11-04"}, "travelers": map[string]any{"adults": 1}})
	resolved, _ := structpb.NewStruct(map[string]any{"origin": map[string]any{"iata": "MOW"}, "destination": map[string]any{"iata": "PAR"}})
	before := &commonv1.TripConfig{Request: request, Resolved: resolved}
	next := proto.Clone(before).(*commonv1.TripConfig)
	next.Request.Fields["budget"], _ = structpb.NewValue(map[string]any{"total": 100000})
	kept := retainResolved(before, next)
	if !proto.Equal(kept.Resolved, resolved) {
		t.Fatal("budget edit lost verified cities")
	}
	next.Request.Fields["destination"], _ = structpb.NewValue(map[string]any{"city": "Сочи"})
	changed := retainResolved(before, next)
	if changed.Resolved.GetFields()["destination"] != nil || changed.Resolved.GetFields()["origin"] == nil {
		t.Fatal("direction change kept stale codes or lost origin")
	}
	if modelConfig(before).GetResolved() != nil || before.GetResolved() == nil {
		t.Fatal("backend codes leaked or original mutated")
	}
}

func TestSearchPersistsVerifiedResolved(t *testing.T) {
	request, _ := structpb.NewStruct(map[string]any{"schema_version": "1.0", "locale": "ru", "currency": "RUB", "origin": map[string]any{"city": "Москва"}, "destination": map[string]any{"city": "Париж"}, "dates": map[string]any{"departure_date": "2026-11-04"}, "travelers": map[string]any{"adults": 1}})
	resolved, _ := structpb.NewStruct(map[string]any{"origin": map[string]any{"iata": "MOW"}, "destination": map[string]any{"iata": "PAR"}})
	sessions := &sessionsFake{state: &sessionv1.GetAgentContextResponse{UserId: "owner", ChatId: "s", Cfg: &commonv1.TripConfig{Request: request}}, status: "confirmed"}
	providers := &providerFake{resolved: resolved}
	controller := New(nil).WithPlanning(sessions, providers)
	_, err := controller.SearchRoutes(context.Background(), &travelv1.SearchRoutesRequest{UserId: "owner", SessionId: "s"})
	if err != nil {
		t.Fatal(err)
	}
	if !proto.Equal(sessions.state.GetCfg().GetResolved(), resolved) {
		t.Fatal("verified codes not saved to session")
	}
	if _, err = controller.GetRoutes(context.Background(), &travelv1.GetRoutesRequest{UserId: "owner", SessionId: "s"}); err != nil {
		t.Fatalf("saved codes invalidated fresh search: %v", err)
	}
}
