package route

import (
	"context"
	"testing"

	aiv1 "github.com/godnota/pkg/generated/ai/v1"
	providerv1 "github.com/godnota/pkg/generated/provider/v1"
	pb "github.com/godnota/pkg/generated/route/v1"
	"github.com/godnota/pkg/testkit"
	aigrpc "github.com/godnota/route-service/internal/adapter/ai/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

type travelAIStub struct {
	aiv1.UnimplementedTravelAIServer
	fn func(context.Context, *aiv1.RankRoutePartsRequest) (*aiv1.RankRoutePartsResponse, error)
}

func (s travelAIStub) RankRouteParts(ctx context.Context, req *aiv1.RankRoutePartsRequest) (*aiv1.RankRoutePartsResponse, error) {
	return s.fn(ctx, req)
}

func TestBuildRoutesSendsRoutesToTravelAI(t *testing.T) {
	var seen int
	aiConn := testkit.Conn(t, func(server *grpc.Server) {
		aiv1.RegisterTravelAIServer(server, travelAIStub{fn: func(_ context.Context, req *aiv1.RankRoutePartsRequest) (*aiv1.RankRoutePartsResponse, error) {
			seen = len(req.GetRoutes())
			if req.GetSessionId() != "s" || req.GetConfig().GetFields()["origin"].GetStructValue().GetFields()["city"].GetStringValue() != "Москва" {
				t.Fatalf("rank request: %v", req)
			}
			return &aiv1.RankRoutePartsResponse{}, nil
		}})
	})
	conn := testkit.Conn(t, func(server *grpc.Server) {
		pb.RegisterRouteServiceServer(server, New(nil).WithRanker(aigrpc.New(aiv1.NewTravelAIClient(aiConn)), 4, 2))
	})
	client := pb.NewRouteServiceClient(conn)
	ctx := context.Background()
	if _, err := client.BuildRoutes(ctx, &pb.BuildRoutesRequest{}); status.Code(err) != codes.InvalidArgument {
		t.Fatal(err)
	}
	config, err := structpb.NewStruct(map[string]any{"origin": map[string]any{"city": "Москва"}})
	if err != nil {
		t.Fatal(err)
	}
	empty, err := client.BuildRoutes(ctx, &pb.BuildRoutesRequest{SessionId: "s", Config: config, Parts: &providerv1.SearchPartsResponse{}})
	if err != nil || len(empty.GetCandidates()) != 0 || seen != 0 {
		t.Fatalf("empty parts: %v seen=%d err=%v", empty, seen, err)
	}
	response, err := client.BuildRoutes(ctx, &pb.BuildRoutesRequest{SessionId: "s", Config: config, Parts: sampleParts()})
	if err != nil {
		t.Fatal(err)
	}
	if seen != 2 || len(response.GetCandidates()) != 2 {
		t.Fatalf("ai saw %d, candidates %d", seen, len(response.GetCandidates()))
	}
	if amount(response.GetCandidates()[0]) != 1200 || amount(response.GetCandidates()[1]) != 9200 {
		t.Fatalf("stub changed route order: %v then %v", amount(response.GetCandidates()[0]), amount(response.GetCandidates()[1]))
	}
}

func TestBuildRoutesRequiresRankerWhenRoutesExist(t *testing.T) {
	conn := testkit.Conn(t, func(server *grpc.Server) {
		pb.RegisterRouteServiceServer(server, New(nil))
	})
	client := pb.NewRouteServiceClient(conn)
	config, _ := structpb.NewStruct(map[string]any{"origin": map[string]any{"city": "Москва"}})
	_, err := client.BuildRoutes(context.Background(), &pb.BuildRoutesRequest{SessionId: "s", Config: config, Parts: sampleParts()})
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatal(err)
	}
}

func TestBuildRoutesPropagatesAIError(t *testing.T) {
	aiConn := testkit.Conn(t, func(server *grpc.Server) {
		aiv1.RegisterTravelAIServer(server, travelAIStub{fn: func(context.Context, *aiv1.RankRoutePartsRequest) (*aiv1.RankRoutePartsResponse, error) {
			return nil, status.Error(codes.Unavailable, "model down")
		}})
	})
	conn := testkit.Conn(t, func(server *grpc.Server) {
		pb.RegisterRouteServiceServer(server, New(nil).WithRanker(aigrpc.New(aiv1.NewTravelAIClient(aiConn)), 4, 2))
	})
	config, _ := structpb.NewStruct(map[string]any{"origin": map[string]any{"city": "Москва"}})
	_, err := pb.NewRouteServiceClient(conn).BuildRoutes(context.Background(), &pb.BuildRoutesRequest{SessionId: "s", Config: config, Parts: sampleParts()})
	if status.Code(err) != codes.Unavailable {
		t.Fatal(err)
	}
}

func sampleParts() *providerv1.SearchPartsResponse {
	return &providerv1.SearchPartsResponse{
		TransportParts: []*providerv1.TransportPart{
			{Id: "out", OriginCity: "Москва", DestinationCity: "Казань", DepartureAt: "2026-12-12T08:00:00Z", ArrivalAt: "2026-12-12T10:00:00Z", Price: &providerv1.Money{AmountMinor: 100, Currency: "RUB"}},
			{Id: "back", OriginCity: "Казань", DestinationCity: "Москва", DepartureAt: "2026-12-14T12:00:00Z", ArrivalAt: "2026-12-14T14:00:00Z", Price: &providerv1.Money{AmountMinor: 100, Currency: "RUB"}},
		},
		HotelParts: []*providerv1.HotelPart{
			{Id: "cheap", City: "Казань", CheckIn: "2026-12-12", CheckOut: "2026-12-14", Price: &providerv1.Money{AmountMinor: 1000, Currency: "RUB"}},
			{Id: "costly", City: "Казань", CheckIn: "2026-12-12", CheckOut: "2026-12-14", Price: &providerv1.Money{AmountMinor: 9000, Currency: "RUB"}},
		},
	}
}

func amount(route *structpb.Struct) float64 {
	return route.GetFields()["price"].GetStructValue().GetFields()["amount_minor"].GetNumberValue()
}
