package provider

import (
	"context"
	"encoding/json"
	pb "github.com/godnota/pkg/generated/provider/v1"
	"github.com/godnota/pkg/testkit"
	"github.com/ostrovok-hackathon-2026/godnota/providerkit"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func loadStruct(t *testing.T, name string) *structpb.Struct {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("../../../../configs/examples", name))
	if err != nil {
		t.Fatal(err)
	}
	var obj map[string]any
	if err = json.Unmarshal(raw, &obj); err != nil {
		t.Fatal(err)
	}
	v, err := structpb.NewStruct(obj)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
func TestGRPCProviderDemoAndErrors(t *testing.T) {
	searcher, _ := providerkit.NewSearcher(nil, "demo")
	searcher.Now = func() time.Time { return time.Date(2026, 10, 28, 0, 0, 0, 0, time.UTC) }
	conn := testkit.Conn(t, func(server *grpc.Server) { pb.RegisterProviderServiceServer(server, New(nil).WithSearcher(searcher)) })
	client := pb.NewProviderServiceClient(conn)
	response, err := client.SearchParts(context.Background(), &pb.SearchPartsRequest{Config: loadStruct(t, "trip_request.kazan_weekend.json")})
	if err != nil || len(response.GetTransportParts()) != 4 || len(response.GetHotelParts()) != 1 {
		t.Fatalf("response lost: %+v %v", response, err)
	}
	if response.GetHotelParts()[0].GetSource() != "synthetic" || response.GetHotelParts()[0].GetPrice().GetAmountMinor() != 1000000 {
		t.Fatal("money/provenance changed over protobuf")
	}
	if _, err = client.SearchParts(context.Background(), &pb.SearchPartsRequest{}); status.Code(err) != codes.InvalidArgument {
		t.Fatal(err)
	}
	invalid := loadStruct(t, "trip_request.kazan_weekend.json")
	invalid.Fields["currency"] = structpb.NewStringValue("JPY")
	if _, err = client.SearchParts(context.Background(), &pb.SearchPartsRequest{Config: invalid}); status.Code(err) != codes.InvalidArgument {
		t.Fatal(err)
	}
	if _, err = New(nil).SearchParts(context.Background(), nil); status.Code(err) != codes.Unavailable {
		t.Fatal(err)
	}
}
func TestGRPCToRealHTTPPartialFailure(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		if strings.Contains(r.URL.Path, "prices_for_dates") {
			if r.Header.Get("X-Access-Token") != "private-key" || r.URL.Query().Get("origin") != "MOW" {
				t.Error("provider request contract incorrect")
			}
			raw, _ := os.ReadFile("../../../../testdata/aviasales/prices_for_dates.json")
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(raw)
			return
		}
		http.Error(w, "private-key must never reach clients", http.StatusTooManyRequests)
	}))
	defer server.Close()
	registry, err := providerkit.LoadRegistry("../../../../configs/providers.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"aviasales", "yandex_rasp"} {
		cfg := registry[name]
		cfg.BaseURL = server.URL
		registry[name] = cfg
	}
	searcher, _ := providerkit.NewSearcher(registry, "live")
	searcher.Client = server.Client()
	searcher.Now = func() time.Time { return time.Date(2026, 10, 28, 0, 0, 0, 0, time.UTC) }
	t.Setenv("TRAVELPAYOUTS_TOKEN", "private-key")
	t.Setenv("YANDEX_RASP_API_KEY", "private-key")
	conn := testkit.Conn(t, func(s *grpc.Server) { pb.RegisterProviderServiceServer(s, New(nil).WithSearcher(searcher)) })
	client := pb.NewProviderServiceClient(conn)
	response, err := client.SearchParts(context.Background(), &pb.SearchPartsRequest{Config: loadStruct(t, "trip_request.kazan_weekend.json"), Resolved: loadStruct(t, "resolved_trip.kazan_weekend.json")})
	if err != nil {
		t.Fatal(err)
	}
	if requests.Load() < 2 || len(response.GetTransportParts()) == 0 || len(response.GetErrors()) == 0 {
		t.Fatalf("partial results lost: %+v", response)
	}
	if strings.Contains(response.String(), "private-key") {
		t.Fatal("credential leaked")
	}
	retryable := false
	for _, e := range response.GetErrors() {
		if e.GetCode() == "http_429" && e.GetRetryable() {
			retryable = true
		}
	}
	if !retryable {
		t.Fatal("429 retryability lost")
	}
}
