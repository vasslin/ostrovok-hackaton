package travel

import (
	"context"
	commonv1 "github.com/godnota/pkg/generated/common/v1"
	providerv1 "github.com/godnota/pkg/generated/provider/v1"
	routev1 "github.com/godnota/pkg/generated/route/v1"
	sessionv1 "github.com/godnota/pkg/generated/session/v1"
	pb "github.com/godnota/pkg/generated/travel/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"strings"
	"testing"
	"time"
)

type scriptedProvider struct {
	providerv1.ProviderServiceClient
	fn func(context.Context, *providerv1.SearchPartsRequest) (*providerv1.SearchPartsResponse, error)
}

func (f scriptedProvider) SearchParts(ctx context.Context, r *providerv1.SearchPartsRequest, _ ...grpc.CallOption) (*providerv1.SearchPartsResponse, error) {
	return f.fn(ctx, r)
}

type scriptedRoutes struct {
	routev1.RouteServiceClient
	fn func(context.Context, *routev1.BuildRoutesRequest) (*routev1.BuildRoutesResponse, error)
}

func (f scriptedRoutes) BuildRoutes(ctx context.Context, r *routev1.BuildRoutesRequest, _ ...grpc.CallOption) (*routev1.BuildRoutesResponse, error) {
	return f.fn(ctx, r)
}

type failedSessions struct {
	sessionv1.SessionServiceClient
	state             *sessionv1.GetContextResponse
	readErr, writeErr error
}

func (f failedSessions) GetContext(context.Context, *sessionv1.GetContextRequest, ...grpc.CallOption) (*sessionv1.GetContextResponse, error) {
	return f.state, f.readErr
}
func (f failedSessions) SaveContext(context.Context, *sessionv1.SaveContextRequest, ...grpc.CallOption) (*sessionv1.SaveContextResponse, error) {
	return nil, f.writeErr
}
func planningState() *sessionv1.GetContextResponse {
	now := time.Now()
	return &sessionv1.GetContextResponse{SessionId: "s", UserId: "owner", Status: "confirmed", TripConfig: &commonv1.TripConfig{Trip: &commonv1.Trip{Origin: &commonv1.Place{City: "Москва"}, Destinations: []*commonv1.Place{{City: "Казань"}}, DateRange: &commonv1.DateRange{StartDate: now.AddDate(0, 0, 7).Format(time.DateOnly), EndDate: now.AddDate(0, 0, 9).Format(time.DateOnly)}, Travelers: &commonv1.Travelers{Adults: 2}}}}
}
func TestPlanningFailuresDoNotBecomeSuccess(t *testing.T) {
	ctx := context.Background()
	state := planningState()
	s := &sessionsFake{state: state}
	c := New(nil).WithPlanning(s, scriptedProvider{fn: func(context.Context, *providerv1.SearchPartsRequest) (*providerv1.SearchPartsResponse, error) {
		return nil, status.Error(codes.Unavailable, "provider down")
	}})
	if _, err := c.GetConfig(ctx, &pb.GetConfigRequest{}); status.Code(err) != codes.InvalidArgument {
		t.Fatal(err)
	}
	if _, err := c.UpdateConfig(ctx, &pb.UpdateConfigRequest{SessionId: "s"}); status.Code(err) != codes.InvalidArgument {
		t.Fatal(err)
	}
	if _, err := c.SearchRoutes(ctx, &pb.SearchRoutesRequest{SessionId: "s"}); status.Code(err) != codes.Unavailable {
		t.Fatal(err)
	}
	if _, err := c.GetRoutes(ctx, &pb.GetRoutesRequest{SessionId: "s"}); status.Code(err) != codes.NotFound {
		t.Fatal("failed search cached", err)
	}
	c.providers = scriptedProvider{fn: func(context.Context, *providerv1.SearchPartsRequest) (*providerv1.SearchPartsResponse, error) {
		return nil, nil
	}}
	if _, err := c.SearchRoutes(ctx, &pb.SearchRoutesRequest{SessionId: "s"}); status.Code(err) != codes.DataLoss {
		t.Fatal("nil success accepted", err)
	}
	c.sessions = failedSessions{readErr: status.Error(codes.NotFound, "missing")}
	if _, err := c.GetConfig(ctx, &pb.GetConfigRequest{SessionId: "s"}); status.Code(err) != codes.NotFound {
		t.Fatal(err)
	}
	if _, err := c.ConfirmConfig(ctx, &pb.ConfirmConfigRequest{SessionId: "s"}); status.Code(err) != codes.NotFound {
		t.Fatal(err)
	}
	c.sessions = failedSessions{state: state, writeErr: status.Error(codes.Unavailable, "storage")}
	if _, err := c.UpdateConfig(ctx, &pb.UpdateConfigRequest{SessionId: "s", Config: state.TripConfig}); status.Code(err) != codes.Unavailable {
		t.Fatal("save error swallowed", err)
	}
	c.sessions = s
	invalid := &commonv1.TripConfig{}
	if _, err := c.UpdateConfig(ctx, &pb.UpdateConfigRequest{SessionId: "s", Config: invalid}); status.Code(err) != codes.InvalidArgument {
		t.Fatal(err)
	}
	invalid = proto.Clone(state.TripConfig).(*commonv1.TripConfig)
	invalid.Trip.DateRange.EndDate = "2020-01-01"
	if _, err := c.ConfirmConfig(ctx, &pb.ConfirmConfigRequest{SessionId: "s", Config: invalid}); status.Code(err) != codes.InvalidArgument {
		t.Fatal(err)
	}
	s.state = planningState()
	s.state.TripConfig = nil
	if _, err := c.SearchRoutes(ctx, &pb.SearchRoutesRequest{SessionId: "s"}); status.Code(err) != codes.InvalidArgument {
		t.Fatal(err)
	}
}
func TestRouteHandoffPreservesPartialResultsAndCache(t *testing.T) {
	state := planningState()
	s := &sessionsFake{state: state}
	calls := 0
	p := scriptedProvider{fn: func(context.Context, *providerv1.SearchPartsRequest) (*providerv1.SearchPartsResponse, error) {
		return &providerv1.SearchPartsResponse{TransportParts: []*providerv1.TransportPart{{Id: "transport", Source: "synthetic"}}, Errors: []*providerv1.ProviderError{{Provider: "hotel", Code: "timeout"}}}, nil
	}}
	routes := scriptedRoutes{fn: func(_ context.Context, r *routev1.BuildRoutesRequest) (*routev1.BuildRoutesResponse, error) {
		calls++
		if r.GetSessionId() != "s" || r.GetConfig() == nil || len(r.GetParts().GetTransportParts()) != 1 || len(r.GetParts().GetErrors()) != 1 {
			t.Fatal("handoff dropped data")
		}
		return nil, status.Error(codes.Unimplemented, "private-stage3-detail")
	}}
	c := New(nil).WithPlanning(s, p, routes)
	result, err := c.SearchRoutes(context.Background(), &pb.SearchRoutesRequest{SessionId: "s"})
	if err != nil || calls != 1 {
		t.Fatal(err)
	}
	if len(result.GetParts().GetTransportParts()) != 1 || !strings.Contains(result.GetParts().GetLimitations()[0], "Unimplemented") || strings.Contains(result.String(), "private-stage3-detail") {
		t.Fatal("handoff failure lost results or leaked details")
	}
	result.Parts.TransportParts[0].Id = "mutated"
	cached, err := c.GetRoutes(context.Background(), &pb.GetRoutesRequest{SessionId: "s"})
	if err != nil || cached.GetParts().GetTransportParts()[0].GetId() != "transport" {
		t.Fatal("cached shared provider state")
	}
	c.routes = scriptedRoutes{fn: func(context.Context, *routev1.BuildRoutesRequest) (*routev1.BuildRoutesResponse, error) {
		return &routev1.BuildRoutesResponse{}, nil
	}}
	if _, err = c.SearchRoutes(context.Background(), &pb.SearchRoutesRequest{SessionId: "s"}); err != nil {
		t.Fatal(err)
	}
	c.providers = scriptedProvider{fn: func(context.Context, *providerv1.SearchPartsRequest) (*providerv1.SearchPartsResponse, error) {
		return &providerv1.SearchPartsResponse{}, nil
	}}
	c.routes = scriptedRoutes{fn: func(context.Context, *routev1.BuildRoutesRequest) (*routev1.BuildRoutesResponse, error) {
		t.Fatal("empty search handed to stage 3")
		return nil, nil
	}}
	if _, err = c.SearchRoutes(context.Background(), &pb.SearchRoutesRequest{SessionId: "s"}); err != nil {
		t.Fatal(err)
	}
}
