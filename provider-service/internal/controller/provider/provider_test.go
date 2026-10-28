package provider

import (
	"context"
	"net"
	"testing"
	"time"

	pb "github.com/godnota/pkg/generated/provider/v1"
	"github.com/godnota/provider-service/internal/entity"
	"github.com/ostrovok-hackathon-2026/godnota/providerkit"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/structpb"
)

type testService struct {
	called bool
	trip   providerkit.TripRequest
}

func (s *testService) Search(_ context.Context, trip providerkit.TripRequest, _ providerkit.ResolvedTrip, _ string) (entity.Result, error) {
	s.called, s.trip = true, trip
	return entity.Result{Offers: []*pb.Offer{{Id: "one", Provider: "example", PriceScope: "unknown"}}, Plan: providerkit.TripPlan{Discovery: providerkit.RequestSet{}}}, nil
}

func object(t *testing.T, v map[string]any) *structpb.Struct {
	t.Helper()
	x, err := structpb.NewStruct(v)
	if err != nil {
		t.Fatal(err)
	}
	return x
}

func frontendInput(t *testing.T, pets bool) *pb.SearchInput {
	t.Helper()
	return &pb.SearchInput{InputFormat: "frontend_criteria", TripRequest: object(t, map[string]any{
		"origin": "Москва", "destination": "Казань", "startDate": time.Now().UTC().AddDate(0, 0, 10).Format(time.DateOnly),
		"nights": 2, "adults": 2, "childrenAges": []any{}, "budgetMinor": 4000000, "currency": "RUB", "pets": pets,
	})}
}

func TestFrontendUnresolvedPetStopsSearch(t *testing.T) {
	service := &testService{}
	response, err := New(service).SearchTransport(context.Background(), &pb.SearchTransportRequest{Input: frontendInput(t, true)})
	if err != nil {
		t.Fatal(err)
	}
	if service.called || len(response.Missing) == 0 || response.Missing[0] != "/travelers/pets/kind" {
		t.Fatalf("unexpected response: %+v", response)
	}
}

func TestGRPCFrontendContract(t *testing.T) {
	listener := bufconn.Listen(1 << 20)
	server := grpc.NewServer()
	service := &testService{}
	pb.RegisterProviderServiceServer(server, New(service))
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(server.Stop)
	conn, err := grpc.NewClient("passthrough:///bufnet", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	response, err := pb.NewProviderServiceClient(conn).SearchTransport(ctx, &pb.SearchTransportRequest{Input: frontendInput(t, false)})
	if err != nil {
		t.Fatal(err)
	}
	if !service.called || service.trip.Budget == nil || service.trip.Budget.Total != 40000 || len(response.Offers) != 1 || response.Offers[0].PriceMinor != nil {
		t.Fatalf("incorrect gRPC mapping: %+v", response)
	}
}

func TestUnknownInputFieldRejected(t *testing.T) {
	service := &testService{}
	input := frontendInput(t, false)
	input.TripRequest.Fields["extra"] = structpb.NewStringValue("unexpected")
	_, err := New(service).SearchTransport(context.Background(), &pb.SearchTransportRequest{Input: input})
	if status.Code(err) != codes.InvalidArgument || service.called {
		t.Fatalf("expected invalid argument, got %v", err)
	}
}
