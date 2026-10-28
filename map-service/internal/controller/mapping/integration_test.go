package mapping

import (
	"context"
	"encoding/json"
	pb "github.com/godnota/pkg/generated/map/v1"
	"github.com/godnota/pkg/testkit"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"os"
	"testing"
	"time"
)

func TestGRPCMapConfig(t *testing.T) {
	conn := testkit.Conn(t, func(server *grpc.Server) { pb.RegisterMapServiceServer(server, New(nil)) })
	client := pb.NewMapServiceClient(conn)
	raw, err := os.ReadFile("../../../../configs/examples/trip_request.kazan_weekend.json")
	if err != nil {
		t.Fatal(err)
	}
	var obj map[string]any
	if err = json.Unmarshal(raw, &obj); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	obj["dates"] = map[string]any{"departure_date": now.AddDate(0, 0, 7).Format(time.DateOnly), "return_date": now.AddDate(0, 0, 9).Format(time.DateOnly), "nights": 2}
	input, _ := structpb.NewStruct(obj)
	raw, err = os.ReadFile("../../../../configs/examples/resolved_trip.kazan_weekend.json")
	if err != nil {
		t.Fatal(err)
	}
	var codesObj map[string]any
	if err = json.Unmarshal(raw, &codesObj); err != nil {
		t.Fatal(err)
	}
	resolved, _ := structpb.NewStruct(codesObj)
	response, err := client.MapToProviderConfigs(context.Background(), &pb.MapToProviderConfigsRequest{Config: input, Resolved: resolved})
	if err != nil {
		t.Fatal(err)
	}
	providers := map[string]bool{}
	for _, cfg := range response.GetConfigs() {
		providers[cfg.GetProvider()] = true
		if cfg.GetProvider() == "hotel" {
			q := cfg.GetParams().AsMap()
			if q["meal"] != "breakfast" || q["currency"] != "RUB" || q["adults"] != float64(2) {
				t.Fatalf("hotel requirements lost: %+v", q)
			}
		}
	}
	if !providers["hotel"] || !providers["aviasales"] || !providers["yandex_rasp"] || len(response.GetMissing()) != 0 {
		t.Fatalf("incomplete mapping: %+v", response)
	}
	withoutCodes, err := client.MapToProviderConfigs(context.Background(), &pb.MapToProviderConfigsRequest{Config: input})
	if err != nil || len(withoutCodes.GetMissing()) == 0 {
		t.Fatal("missing codes not reported", err)
	}
	unknown, _ := structpb.NewStruct(map[string]any{"unexpected": true})
	for _, invalid := range []*structpb.Struct{nil, unknown} {
		if _, err = client.MapToProviderConfigs(context.Background(), &pb.MapToProviderConfigsRequest{Config: invalid}); status.Code(err) != codes.InvalidArgument {
			t.Fatal(err)
		}
	}
}
