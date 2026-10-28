package adapter

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/godnota/gateway/internal/httpapi"
	mapv1 "github.com/godnota/pkg/generated/map/v1"
	orderv1 "github.com/godnota/pkg/generated/order/v1"
	providerv1 "github.com/godnota/pkg/generated/provider/v1"
	sessionv1 "github.com/godnota/pkg/generated/session/v1"
	travelv1 "github.com/godnota/pkg/generated/travel/v1"
	userv1 "github.com/godnota/pkg/generated/user/v1"
	"github.com/godnota/pkg/testkit"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
	"time"
)

type sessionServer struct {
	sessionv1.UnimplementedSessionServiceServer
}

func (sessionServer) CreateSession(_ context.Context, r *sessionv1.CreateSessionRequest) (*sessionv1.CreateSessionResponse, error) {
	if r.GetUserId() != "owner" {
		return nil, status.Error(codes.InvalidArgument, "user_id required")
	}
	return &sessionv1.CreateSessionResponse{SessionId: "created"}, nil
}

type travelServer struct {
	travelv1.UnimplementedTravelServiceServer
}

func (travelServer) SearchRoutes(_ context.Context, r *travelv1.SearchRoutesRequest) (*travelv1.SearchRoutesResponse, error) {
	return &travelv1.SearchRoutesResponse{SessionId: r.GetSessionId(), Parts: &providerv1.SearchPartsResponse{HotelParts: []*providerv1.HotelPart{{Id: "hotel", Source: "synthetic", Price: &providerv1.Money{AmountMinor: 123456, Currency: "RUB", Scope: "stay_all_rooms"}}}}}, nil
}
func TestGRPCDispatchAndJSONContract(t *testing.T) {
	methods := make(chan string, 30)
	conn := testkit.Conn(t, func(s *grpc.Server) {
		sessionv1.RegisterSessionServiceServer(s, sessionServer{})
		travelv1.RegisterTravelServiceServer(s, travelServer{})
		userv1.RegisterUserServiceServer(s, &userv1.UnimplementedUserServiceServer{})
		orderv1.RegisterOrderServiceServer(s, &orderv1.UnimplementedOrderServiceServer{})
		mapv1.RegisterMapServiceServer(s, &mapv1.UnimplementedMapServiceServer{})
	}, grpc.UnaryInterceptor(func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		methods <- info.FullMethod
		return handler(ctx, req)
	}))
	b := &backend{session: sessionv1.NewSessionServiceClient(conn), travel: travelv1.NewTravelServiceClient(conn), user: userv1.NewUserServiceClient(conn), order: orderv1.NewOrderServiceClient(conn), mapper: mapv1.NewMapServiceClient(conn), slots: make(chan struct{}, 2)}
	raw, err := b.Call(context.Background(), "/v1/session/create", json.RawMessage(`{"user_id":"owner"}`))
	if err != nil || string(raw) != `{"session_id":"created"}` {
		t.Fatalf("session serialization: %s %v", raw, err)
	}
	<-methods
	raw, err = b.Call(context.Background(), "/v1/travel/search", json.RawMessage(`{"session_id":"s"}`))
	if err != nil {
		t.Fatal(err)
	}
	var result struct {
		Session string `json:"session_id"`
		Parts   struct {
			Hotels []struct {
				Price struct {
					Minor string `json:"amount_minor"`
				} `json:"price"`
			} `json:"hotel_parts"`
		} `json:"parts"`
	}
	if json.Unmarshal(raw, &result) != nil || result.Session != "s" || result.Parts.Hotels[0].Price.Minor != "123456" {
		t.Fatalf("money/parts contract: %s", raw)
	}
	<-methods
	for _, test := range []struct{ operation, method string }{{"/v1/map/config", mapv1.MapService_MapToProviderConfigs_FullMethodName}, {"/v1/session/get", sessionv1.SessionService_GetSession_FullMethodName}, {"/v1/travel/plan", travelv1.TravelService_PlanTrip_FullMethodName}, {"/v1/travel/answers", travelv1.TravelService_SubmitAnswers_FullMethodName}, {"/v1/travel/idea", travelv1.TravelService_SelectIdea_FullMethodName}, {"/v1/travel/config/get", travelv1.TravelService_GetConfig_FullMethodName}, {"/v1/travel/config/update", travelv1.TravelService_UpdateConfig_FullMethodName}, {"/v1/travel/config/confirm", travelv1.TravelService_ConfirmConfig_FullMethodName}, {"/v1/travel/routes", travelv1.TravelService_GetRoutes_FullMethodName}, {"/v1/order/create", orderv1.OrderService_CreateOrder_FullMethodName}, {"/v1/order/get", orderv1.OrderService_GetOrder_FullMethodName}, {"/v1/order/pay", orderv1.OrderService_PayOrder_FullMethodName}, {"/v1/order/cancel", orderv1.OrderService_CancelOrder_FullMethodName}, {"/v1/user/get", userv1.UserService_GetUser_FullMethodName}, {"/v1/user/update", userv1.UserService_UpdatePreferences_FullMethodName}} {
		_, _ = b.Call(context.Background(), test.operation, json.RawMessage(`{}`))
		select {
		case got := <-methods:
			if got != test.method {
				t.Fatalf("%s dispatched to %s", test.operation, got)
			}
		case <-time.After(time.Second):
			t.Fatal("RPC not dispatched")
		}
	}
	for _, raw := range []json.RawMessage{json.RawMessage(`{"unknown":true}`), json.RawMessage(`{"user_id":42}`), json.RawMessage(`null garbage`)} {
		if _, err = b.Call(context.Background(), "/v1/session/create", raw); err == nil {
			t.Fatal("invalid JSON accepted")
		}
	}
	if _, err = b.Call(context.Background(), "/unknown", json.RawMessage(`{}`)); err == nil {
		t.Fatal("unknown operation accepted")
	}
	b.slots <- struct{}{}
	b.slots <- struct{}{}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = b.Call(ctx, "/v1/session/get", nil); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
func TestErrorTranslationAndClientLifecycle(t *testing.T) {
	for _, err := range []error{nil, context.Canceled, context.DeadlineExceeded, status.Error(codes.Canceled, "canceled"), status.Error(codes.DeadlineExceeded, "deadline"), status.Error(codes.NotFound, "missing"), errors.New("private internal error")} {
		translated := callError(err)
		if err != nil && translated == nil {
			t.Fatal("error lost")
		}
		if err != nil && err.Error() == "private internal error" {
			public, ok := translated.(*httpapi.Error)
			if !ok || public.Message != "internal error" {
				t.Fatal("internal error leaked")
			}
		}
	}
	b, closeClients, err := New("127.0.0.1:1", "127.0.0.1:1", "127.0.0.1:1", "127.0.0.1:1", "127.0.0.1:1")
	if err != nil || len(b.conns) != 5 {
		t.Fatal(err)
	}
	closeClients()
}
