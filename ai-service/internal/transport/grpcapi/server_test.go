package grpcapi

import (
	aiv1 "ai-service/api/ai/v1"
	commonv1 "ai-service/api/common/v1"
	travelv1 "ai-service/api/travel/v1"
	"ai-service/internal/adapter/agent"
	"ai-service/internal/service"
	"context"
	"encoding/json"
	"google.golang.org/protobuf/encoding/protojson"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	healthv1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestGRPC(t *testing.T) {
	tests := []struct {
		name, auth, modelReply string
		httpStatus             int
		want                   codes.Code
	}{
		{"rank", "Bearer secret", `{"results":[{"id":"h1","score":90,"reason":"Рядом с морем","risks":[]}]}`, 200, codes.OK},
		{"auth", "", `{}`, 200, codes.Unauthenticated},
		{"quota", "Bearer secret", `{}`, 429, codes.ResourceExhausted},
		{"invented hotel", "Bearer secret", `{"results":[{"id":"other","score":90,"reason":"ok"}]}`, 200, codes.Unavailable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			calls := 0
			provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				var body struct {
					Messages []struct {
						Content string `json:"content"`
					} `json:"messages"`
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
				}
				var in service.RankInput
				if len(body.Messages) < 2 {
					t.Error("missing model input")
				} else if err := json.Unmarshal([]byte(body.Messages[1].Content), &in); err != nil || len(in.Hotels) != 1 || in.Hotels[0].Name != "Sea Hotel" || in.Hotels[0].Details["total_price"] != float64(12000) {
					t.Errorf("hotel data was not forwarded: %+v", in)
				}
				w.Header().Set("Retry-After", "3")
				w.WriteHeader(test.httpStatus)
				_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": map[string]string{"content": test.modelReply}, "finish_reason": "stop"}}})
			}))
			defer provider.Close()
			server := New(service.New(agent.New(provider.URL, "test", "model")), "secret")
			conn := connect(t, server)
			details, _ := structpb.NewStruct(map[string]any{"total_price": 12000, "currency": "RUB"})
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", test.auth)
			var trailer metadata.MD
			result, err := travelv1.NewTravelAIClient(conn).RankHotels(ctx, &travelv1.RankHotelsRequest{Hotels: []*travelv1.Hotel{{Id: "h1", Name: "Sea Hotel", Details: details}}}, grpc.Trailer(&trailer))
			if status.Code(err) != test.want {
				t.Fatalf("got %v want %v: %v", status.Code(err), test.want, err)
			}
			if test.want == codes.OK && (len(result.Results) != 1 || result.Results[0].Id != "h1") {
				t.Fatal("wrong result")
			}
			if test.want == codes.Unauthenticated && calls != 0 {
				t.Fatal("unauthorized call consumed quota")
			}
			if test.want == codes.ResourceExhausted && (len(trailer.Get("retry-after")) != 1 || trailer.Get("retry-after")[0] != "3") {
				t.Fatal("retry trailer missing")
			}
			if calls > 1 {
				t.Fatal("hidden retry")
			}
		})
	}
}
func connect(t *testing.T, server *grpc.Server) *grpc.ClientConn {
	t.Helper()
	listener := bufconn.Listen(1 << 20)
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(func() { server.Stop(); listener.Close() })
	conn, err := grpc.NewClient("passthrough:///test", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}
func TestHealthAndMissingKey(t *testing.T) {
	server := New(service.New(agent.New("http://unused", "", "model")), "")
	conn := connect(t, server)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	health, err := healthv1.NewHealthClient(conn).Check(ctx, &healthv1.HealthCheckRequest{})
	if err != nil || health.Status != healthv1.HealthCheckResponse_SERVING {
		t.Fatalf("health: %v", err)
	}
	_, err = travelv1.NewTravelAIClient(conn).ExtractConfig(ctx, &travelv1.ExtractConfigRequest{Message: "В Сочи"})
	if status.Code(err) != codes.Unavailable {
		t.Fatalf("expected unavailable, got %v", err)
	}
}

func TestExtractConfigAndValidation(t *testing.T) {
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": map[string]string{"content": `{"config":{"destination":"Сочи","adults":2},"questions":["Когда едете?"]}`}, "finish_reason": "stop"}}})
	}))
	defer provider.Close()
	conn := connect(t, New(service.New(agent.New(provider.URL, "key", "model")), ""))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	client := travelv1.NewTravelAIClient(conn)
	out, err := client.ExtractConfig(ctx, &travelv1.ExtractConfigRequest{Message: "В Сочи вдвоём"})
	if err != nil {
		t.Fatal(err)
	}
	if out.Config.GetDestination() != "Сочи" || out.Config.Adults == nil || *out.Config.Adults != 2 || out.Config.Origin != nil || len(out.MissingFields) != 3 {
		t.Fatalf("wrong extraction: %v", out)
	}
	_, err = client.RankHotels(ctx, &travelv1.RankHotelsRequest{Hotels: []*travelv1.Hotel{{Id: "h1"}}})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("hotel without name should be invalid: %v", err)
	}
}

func TestPlanningServiceContract(t *testing.T) {
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": map[string]string{"content": `{"status":"generate","config":{"version":"1.0","trip":{"origin":{"city":"Москва"},"destinations":[{"city":"Сочи"}],"date_range":{"start_date":"2026-11-10","end_date":"2026-11-17"},"travelers":{"adults":2}}},"advice":[],"questions":[],"assumptions":[],"cleared_fields":[],"summary":"Поездка в Сочи","warnings":[]}`}, "finish_reason": "stop"}}})
	}))
	defer provider.Close()
	conn := connect(t, New(service.New(agent.New(provider.URL, "key", "model")), "secret"))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	client := aiv1.NewAIServiceClient(conn)
	_, err := client.GenerateOrUpdateConfig(ctx, &aiv1.GenerateOrUpdateConfigRequest{Text: "Поездка"})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("canonical service auth: %v", err)
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer secret")
	out, err := client.GenerateOrUpdateConfig(ctx, &aiv1.GenerateOrUpdateConfigRequest{Text: "Из Москвы в Сочи", Config: &commonv1.TripConfig{Version: "1.0", Trip: &commonv1.Trip{Origin: &commonv1.Place{City: "Москва"}, Destinations: []*commonv1.Place{{City: "Сочи"}}, DateRange: &commonv1.DateRange{StartDate: "2026-11-10", EndDate: "2026-11-17"}, Travelers: &commonv1.Travelers{Adults: 2}}}, CurrentDate: "2026-10-28"})
	if err != nil {
		t.Fatal(err)
	}
	if out.Status != "generate" || out.Config.GetRequest().AsMap()["origin"].(map[string]any)["city"] != "Москва" || out.Config.GetRequest().AsMap()["travelers"].(map[string]any)["adults"] != float64(2) {
		t.Fatalf("bad canonical response: %v", out)
	}
	refined, err := client.RefineConfig(ctx, &aiv1.RefineConfigRequest{Config: out.Config, CurrentDate: "2026-10-28"})
	if err != nil || refined.GetStatus() != "generate" {
		t.Fatalf("refine failed: %v", err)
	}
}

func TestTypedRankingGRPC(t *testing.T) {
	provider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": map[string]string{"content": `{"results":[{"id":"route-1","score":88}]}`}, "finish_reason": "stop"}}})
	}))
	defer provider.Close()
	conn := connect(t, New(service.New(agent.New(provider.URL, "key", "model")), ""))
	data, err := os.ReadFile("../../../../examples/ranking/routes.json")
	if err != nil {
		t.Fatal(err)
	}
	req := new(aiv1.RankRoutesRequest)
	if err = protojson.Unmarshal(data, req); err != nil {
		t.Fatal(err)
	}
	client := aiv1.NewAIServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for _, call := range []func(context.Context, *aiv1.RankRoutesRequest, ...grpc.CallOption) (*aiv1.RankRoutesResponse, error){client.RankRoutes, client.RankJoinOptions, client.RankFinalRoutes} {
		out, err := call(ctx, req)
		if err != nil {
			t.Fatal(err)
		}
		if len(out.Results) != 1 || out.Results[0].GetTotalPrice().GetAmountMinor() != 6500000 || !out.Results[0].RequiresVerification || out.Parts.TransportParts[0].Source != "synthetic" {
			t.Fatalf("bad gRPC ranking: %+v", out)
		}
	}
}
