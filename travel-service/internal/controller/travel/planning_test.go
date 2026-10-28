package travel

import (
	"context"
	"encoding/json"
	commonv1 "github.com/godnota/pkg/generated/common/v1"
	providerv1 "github.com/godnota/pkg/generated/provider/v1"
	sessionv1 "github.com/godnota/pkg/generated/session/v1"
	pb "github.com/godnota/pkg/generated/travel/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"os"
	"testing"
	"time"
)

type sessionsFake struct {
	sessionv1.SessionServiceClient
	state  *sessionv1.GetAgentContextResponse
	status string
}

func (f *sessionsFake) GetSession(context.Context, *sessionv1.GetSessionRequest, ...grpc.CallOption) (*sessionv1.GetSessionResponse, error) {
	return &sessionv1.GetSessionResponse{SessionId: f.state.GetChatId(), UserId: f.state.GetUserId(), Status: f.status}, nil
}
func (f *sessionsFake) GetAgentContext(context.Context, *sessionv1.GetAgentContextRequest, ...grpc.CallOption) (*sessionv1.GetAgentContextResponse, error) {
	return f.state, nil
}
func (f *sessionsFake) SaveContext(_ context.Context, r *sessionv1.SaveContextRequest, _ ...grpc.CallOption) (*sessionv1.SaveContextResponse, error) {
	if r.GetCfg() != nil {
		f.state.Cfg = r.GetCfg()
	}
	return &sessionv1.SaveContextResponse{}, nil
}

type providerFake struct {
	providerv1.ProviderServiceClient
	calls int
}

func (f *providerFake) SearchParts(context.Context, *providerv1.SearchPartsRequest, ...grpc.CallOption) (*providerv1.SearchPartsResponse, error) {
	f.calls++
	return &providerv1.SearchPartsResponse{HotelParts: []*providerv1.HotelPart{{Id: "h", Source: "synthetic"}}}, nil
}
func TestConfirmationAndInvalidation(t *testing.T) {
	raw, e := os.ReadFile("../../../../configs/examples/trip_request.kazan_weekend.json")
	if e != nil {
		t.Fatal(e)
	}
	var plain map[string]any
	_ = json.Unmarshal(raw, &plain)
	// Keep this test valid independently of the calendar date.
	now := time.Now()
	plain["dates"] = map[string]any{"departure_date": now.AddDate(0, 0, 7).Format("2006-01-02"), "return_date": now.AddDate(0, 0, 9).Format("2006-01-02"), "nights": 2}
	doc, _ := structpb.NewStruct(plain)
	cfg := &commonv1.TripConfig{Request: doc}
	s := &sessionsFake{state: &sessionv1.GetAgentContextResponse{UserId: "owner", ChatId: "s", Cfg: cfg}, status: "draft"}
	p := &providerFake{}
	c := New(nil).WithPlanning(s, p)
	ctx := context.Background()
	if _, e = c.SearchRoutes(ctx, &pb.SearchRoutesRequest{SessionId: "s", UserId: "owner"}); status.Code(e) != codes.FailedPrecondition {
		t.Fatal("unconfirmed search accepted")
	}
	if _, e = c.ConfirmConfig(ctx, &pb.ConfirmConfigRequest{SessionId: "s", UserId: "owner"}); e != nil {
		t.Fatal(e)
	}
	result, e := c.SearchRoutes(ctx, &pb.SearchRoutesRequest{SessionId: "s", UserId: "owner"})
	if e != nil || len(result.GetParts().GetHotelParts()) != 1 || p.calls != 1 {
		t.Fatalf("parts lost: %v", e)
	}
	if _, e = c.GetRoutes(ctx, &pb.GetRoutesRequest{SessionId: "s", UserId: "owner"}); e != nil {
		t.Fatal(e)
	}
	if _, e = c.UpdateConfig(ctx, &pb.UpdateConfigRequest{SessionId: "s", UserId: "owner", Config: cfg}); e != nil {
		t.Fatal(e)
	}
	if _, e = c.GetRoutes(ctx, &pb.GetRoutesRequest{SessionId: "s", UserId: "owner"}); status.Code(e) != codes.NotFound {
		t.Fatal("stale search survived edit")
	}
	if _, e = c.SearchRoutes(ctx, &pb.SearchRoutesRequest{SessionId: "s", UserId: "owner"}); status.Code(e) != codes.FailedPrecondition {
		t.Fatal("edit did not reset confirmation")
	}
}
