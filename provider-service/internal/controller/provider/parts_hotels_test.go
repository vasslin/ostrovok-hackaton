package provider

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"

	pb "github.com/godnota/pkg/generated/provider/v1"
	"github.com/godnota/pkg/testkit"
	"github.com/godnota/provider-service/internal/entity"
	"github.com/ostrovok-hackathon-2026/godnota/providerkit"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
)

type hotelPartsFixture struct {
	called bool
	result entity.Result
}

func (s *hotelPartsFixture) Search(_ context.Context, _ providerkit.TripRequest, _ providerkit.ResolvedTrip, kind string) (entity.Result, error) {
	s.called = kind == "hotel"
	return s.result, nil
}

func TestSearchPartsIncludesHotelOffersAndWarnings(t *testing.T) {
	t.Setenv("HOTELBEDS_MODE", "evaluation")
	t.Setenv("RATEHAWK_MODE", "")
	t.Setenv("TRAVELPAYOUTS_TOKEN", "")
	t.Setenv("YANDEX_RASP_API_KEY", "")
	reg, err := providerkit.LoadRegistry("../../../../configs/providers.json")
	if err != nil {
		t.Fatal(err)
	}
	searcher, err := providerkit.NewSearcher(reg, "live")
	if err != nil {
		t.Fatal(err)
	}
	searcher.Now = func() time.Time { return time.Date(2026, 10, 28, 0, 0, 0, 0, time.UTC) }
	price := int64(13351)
	service := &hotelPartsFixture{result: entity.Result{
		Offers: []*pb.Offer{{Id: "rate", Kind: "hotel", Provider: "hotelbeds", Title: "Test hotel", PriceMinor: &price, Currency: "EUR", PriceScope: "evaluation net rate", SourceStatus: "sandbox_evaluation"}},
		Issues: []*pb.ProviderIssue{{Provider: "hotelbeds", Code: "sandbox_only", Message: "Sandbox data"}, {Provider: "hotelbeds", Code: "filters_unverified", Message: "Filters unverified"}},
	}}
	conn := testkit.Conn(t, func(server *grpc.Server) {
		pb.RegisterProviderServiceServer(server, New(service).WithSearcher(searcher))
	})
	resolved := loadStruct(t, "resolved_trip.kazan_weekend.json")
	resolved.Fields["hotel_checkin"] = structpb.NewStringValue("2026-12-12")
	resolved.Fields["hotel_checkout"] = structpb.NewStringValue("2026-12-14")
	resolved.Fields["hotelbeds_source_market"] = structpb.NewStringValue("RU")
	response, err := pb.NewProviderServiceClient(conn).SearchParts(context.Background(), &pb.SearchPartsRequest{Config: loadStruct(t, "trip_request.kazan_weekend.json"), Resolved: resolved})
	if err != nil || !service.called || len(response.GetHotelParts()) != 1 {
		t.Fatalf("hotel response: %+v, %v", response, err)
	}
	hotel := response.GetHotelParts()[0]
	if hotel.GetSource() != "sandbox_evaluation" || hotel.GetPrice().GetAmountMinor() != price || hotel.GetPrice().GetCurrency() != "EUR" || hotel.GetPrice().GetScope() != "evaluation net rate" || !strings.Contains(strings.Join(hotel.GetWarnings(), " "), "Filters unverified") {
		t.Fatalf("hotel provenance or price lost: %+v", hotel)
	}
	for _, issue := range response.GetErrors() {
		if issue.GetProvider() == "hotel" && issue.GetCode() == "not_configured" {
			t.Fatal("stale hotel configuration error")
		}
	}
}

func TestHotelPartsPreservesMissingAndUnknownPrice(t *testing.T) {
	service := &hotelPartsFixture{result: entity.Result{Plan: providerkit.TripPlan{Missing: []string{"/resolved/hotelbeds_source_market"}}, Offers: []*pb.Offer{{Id: "unknown", Kind: "hotel", SourceStatus: "live"}}}}
	parts := New(service).addHotelParts(context.Background(), providerkit.TripRequest{}, providerkit.ResolvedTrip{}, providerkit.Parts{})
	if len(parts.Errors) != 1 || parts.Errors[0].Message != "/resolved/hotelbeds_source_market" || len(parts.HotelParts) != 1 || parts.HotelParts[0].Price != nil {
		t.Fatalf("missing data was lost: %+v", parts)
	}
}

func TestDemoModeDoesNotCallHotelProvider(t *testing.T) {
	t.Setenv("HOTELBEDS_MODE", "evaluation")
	searcher, _ := providerkit.NewSearcher(nil, "demo")
	searcher.Now = func() time.Time { return time.Date(2026, 10, 28, 0, 0, 0, 0, time.UTC) }
	service := &hotelPartsFixture{}
	response, err := New(service).WithSearcher(searcher).SearchParts(context.Background(), &pb.SearchPartsRequest{Config: loadStruct(t, "trip_request.kazan_weekend.json")})
	if err != nil || service.called || len(response.GetHotelParts()) != 1 || response.GetHotelParts()[0].GetSource() != "synthetic" {
		t.Fatalf("demo mode changed: %+v, %v", response, err)
	}
}

type itineraryHotelFixture struct {
	mu   sync.Mutex
	seen map[string]string
}

func (s *itineraryHotelFixture) Search(_ context.Context, trip providerkit.TripRequest, resolved providerkit.ResolvedTrip, _ string) (entity.Result, error) {
	s.mu.Lock()
	s.seen[trip.Destination.City] = resolved.HotelCheckin + "/" + resolved.HotelCheckout
	s.mu.Unlock()
	return entity.Result{Offers: []*pb.Offer{{Id: trip.Destination.City, Kind: "hotel", Provider: "hotelbeds", Title: trip.Destination.City, SourceStatus: "sandbox_evaluation"}}}, nil
}
func TestItineraryHotelsUseEachConfirmedStay(t *testing.T) {
	var input providerkit.TripRequest
	if err := json.Unmarshal([]byte(`{"schema_version":"1.0","locale":"ru","currency":"RUB","origin":{"city":"Москва"},"destination":{"city":"Самара"},"dates":{"departure_date":"2026-12-12","return_date":"2026-12-14"},"stops":[{"city":"Казань","check_in":"2026-12-12","check_out":"2026-12-13"},{"city":"Самара","check_in":"2026-12-13","check_out":"2026-12-14"}],"travelers":{"adults":2},"hotel":{"needed":true,"rooms":1}}`), &input); err != nil {
		t.Fatal(err)
	}
	service := &itineraryHotelFixture{seen: map[string]string{}}
	result := New(service).addItineraryHotels(context.Background(), input, providerkit.ResolvedTrip{Stops: []providerkit.ResolvedPlace{{}, {}}}, providerkit.Parts{})
	if len(result.HotelParts) != 2 || service.seen["Казань"] != "2026-12-12/2026-12-13" || service.seen["Самара"] != "2026-12-13/2026-12-14" {
		t.Fatalf("wrong stays: %+v, %v", result, service.seen)
	}
	missing := New(service).addItineraryHotels(context.Background(), input, providerkit.ResolvedTrip{}, providerkit.Parts{})
	if len(missing.Errors) != 1 || missing.Errors[0].Code != "missing_field" || len(missing.HotelParts) != 0 {
		t.Fatal("unverified stop used")
	}
}
