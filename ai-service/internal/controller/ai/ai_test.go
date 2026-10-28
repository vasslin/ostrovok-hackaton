package ai

import (
	"context"
	"testing"

	pb "github.com/godnota/pkg/generated/ai/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestRankJoinRoutesStubKeepsOrder(t *testing.T) {
	first, err := structpb.NewStruct(map[string]any{"id": "cheap"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := structpb.NewStruct(map[string]any{"id": "costly"})
	if err != nil {
		t.Fatal(err)
	}
	response, err := NewTravelAI().RankJoinRoutes(context.Background(), &pb.RankJoinRoutesRequest{
		SessionId: "s",
		JoinId:    "moscow",
		Routes:    []*structpb.Struct{first, second},
	})
	if err != nil {
		t.Fatal(err)
	}
	got := response.GetRoutes()
	if len(got) != 2 || got[0].GetFields()["id"].GetStringValue() != "cheap" || got[1].GetFields()["id"].GetStringValue() != "costly" {
		t.Fatalf("stub reordered routes: %v", got)
	}
}

func TestRankRoutePartsStubKeepsOrder(t *testing.T) {
	first, err := structpb.NewStruct(map[string]any{"id": "cheap"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := structpb.NewStruct(map[string]any{"id": "costly"})
	if err != nil {
		t.Fatal(err)
	}
	response, err := NewTravelAI().RankRouteParts(context.Background(), &pb.RankRoutePartsRequest{
		SessionId: "s",
		Routes:    []*structpb.Struct{first, second},
	})
	if err != nil {
		t.Fatal(err)
	}
	got := response.GetRoutes()
	if len(got) != 2 || got[0].GetFields()["id"].GetStringValue() != "cheap" || got[1].GetFields()["id"].GetStringValue() != "costly" {
		t.Fatalf("stub reordered routes: %v", got)
	}
}
