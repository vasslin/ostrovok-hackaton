package tests

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	providerv1 "github.com/godnota/pkg/generated/provider/v1"
	routev1 "github.com/godnota/pkg/generated/route/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestStage2SearchThenStage3Rank(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	provider := providerv1.NewProviderServiceClient(dial(t, env("PROVIDER_GRPC_ADDR", "127.0.0.1:50058")))
	routes := routev1.NewRouteServiceClient(dial(t, env("ROUTE_GRPC_ADDR", "127.0.0.1:50057")))
	config := loadStruct(t, "trip_request.kazan_weekend.json")
	resolved := loadStruct(t, "resolved_trip.kazan_weekend.json")

	var parts *providerv1.SearchPartsResponse
	var err error
	wait(t, ctx, func() bool {
		parts, err = provider.SearchParts(ctx, &providerv1.SearchPartsRequest{Config: config, Resolved: resolved})
		return err == nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(parts.GetTransportParts()) != 4 || len(parts.GetHotelParts()) != 1 {
		t.Fatalf("provider parts: transport=%d hotels=%d", len(parts.GetTransportParts()), len(parts.GetHotelParts()))
	}
	if parts.GetHotelParts()[0].GetSource() != "synthetic" || !hasOrigin(parts, "Москва") || !hasOrigin(parts, "Казань") {
		t.Fatalf("config was not applied to provider data: %+v", parts)
	}

	var built *routev1.BuildRoutesResponse
	wait(t, ctx, func() bool {
		built, err = routes.BuildRoutes(ctx, &routev1.BuildRoutesRequest{SessionId: "stage23", Config: config, Parts: parts})
		return err == nil
	})
	if err != nil {
		t.Fatal(err)
	}
	candidates := built.GetCandidates()
	if len(candidates) != 1 || candidates[0].GetFields()["rank_source"].GetStringValue() != "mock" {
		t.Fatalf("ranked candidates: %v", candidates)
	}
	if !hasCity(candidates[0], "Москва") || !hasCity(candidates[0], "Казань") || !hasBlockKind(candidates[0], "transport") || !hasBlockKind(candidates[0], "hotel") {
		t.Fatalf("ranked route: %v", candidates[0])
	}
	if candidates[0].GetFields()["price"].GetStructValue().GetFields()["amount_minor"].GetNumberValue() <= 0 {
		t.Fatalf("ranked price: %v", candidates[0].GetFields()["price"])
	}
}

func dial(t *testing.T, addr string) *grpc.ClientConn {
	t.Helper()
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

func wait(t *testing.T, ctx context.Context, ready func() bool) {
	t.Helper()
	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("missing deadline")
	}
	for {
		if ready() {
			return
		}
		if time.Now().After(deadline) {
			return
		}
		time.Sleep(300 * time.Millisecond)
	}
}

func loadStruct(t *testing.T, name string) *structpb.Struct {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "configs", "examples", name))
	if err != nil {
		t.Fatal(err)
	}
	var obj map[string]any
	if err = json.Unmarshal(raw, &obj); err != nil {
		t.Fatal(err)
	}
	value, err := structpb.NewStruct(obj)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func hasOrigin(parts *providerv1.SearchPartsResponse, city string) bool {
	for _, part := range parts.GetTransportParts() {
		if part.GetOriginCity() == city {
			return true
		}
	}
	return false
}

func hasCity(route *structpb.Struct, city string) bool {
	for _, item := range route.GetFields()["cities"].GetListValue().GetValues() {
		if item.GetStringValue() == city {
			return true
		}
	}
	return false
}

func hasBlockKind(route *structpb.Struct, kind string) bool {
	for _, item := range route.GetFields()["blocks"].GetListValue().GetValues() {
		if item.GetStructValue().GetFields()["kind"].GetStringValue() == kind {
			return true
		}
	}
	return false
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
