package tripconfig

import (
	commonv1 "github.com/godnota/pkg/generated/common/v1"
	"google.golang.org/protobuf/types/known/structpb"
	"math"
	"testing"
)

func legacy() *commonv1.TripConfig {
	return &commonv1.TripConfig{Trip: &commonv1.Trip{Origin: &commonv1.Place{City: "Москва"}, Destinations: []*commonv1.Place{{City: "Казань"}}, DateRange: &commonv1.DateRange{StartDate: "2026-12-12", EndDate: "2026-12-14"}, Travelers: &commonv1.Travelers{Adults: 2, Children: 1, ChildrenAges: []int32{8}}, Accommodation: &commonv1.AccommodationPreferences{Required: true, MinStars: 3, MaxPricePerNight: 8000, PreferredTypes: []string{"hotel"}}, Transport: &commonv1.TransportPreferences{Modes: []string{"train"}, MaxTransfers: 1}, Budget: &commonv1.Budget{Amount: 40000, Scope: "total", IsHardLimit: true}}}
}
func TestLegacyConversionPreservesRequirements(t *testing.T) {
	config := legacy()
	value, err := Canonical(config)
	if err != nil {
		t.Fatal(err)
	}
	input, _, err := Decode(value, nil)
	if err != nil {
		t.Fatal(err)
	}
	if input.Currency != "RUB" || input.Locale != "ru" || input.Budget.Total != 40000 || !input.Budget.IsHard || input.Hotel.PricePerNightMax != 8000 || input.Travelers.ChildrenAges[0] != 8 || *input.Transport.MaxTransfers != 1 {
		t.Fatalf("lost requirements: %+v", input)
	}
	config.Request = value
	config.Trip = nil
	if got, err := Canonical(config); err != nil || got != value {
		t.Fatal(err)
	}
	config = legacy()
	config.Trip.Budget = nil
	config.Trip.Transport = nil
	config.Trip.Accommodation = nil
	if _, err = Canonical(config); err != nil {
		t.Fatal(err)
	}
}
func TestInvalidConfigVariants(t *testing.T) {
	empty, _ := structpb.NewStruct(map[string]any{})
	for _, cfg := range []*commonv1.TripConfig{nil, {}, {Request: empty, Trip: &commonv1.Trip{}}, {Trip: &commonv1.Trip{}}} {
		if _, err := Canonical(cfg); err == nil {
			t.Fatal("invalid config accepted")
		}
	}
	for _, mutate := range []func(*commonv1.TripConfig){func(c *commonv1.TripConfig) { c.Trip.Travelers.Children = 2 }, func(c *commonv1.TripConfig) { c.Trip.DateRange.FlexibleDates = true }, func(c *commonv1.TripConfig) { c.Trip.Transport.MaxWaitHours = 5 }, func(c *commonv1.TripConfig) {
		c.Trip.Preferences = &commonv1.Preferences{AvoidOvernightTransfers: true}
	}, func(c *commonv1.TripConfig) {
		c.Trip.Activities = &commonv1.ActivityPreferences{Interests: []string{"museums"}}
	}} {
		c := legacy()
		mutate(c)
		if _, err := Canonical(c); err == nil {
			t.Fatal("lossy legacy conversion accepted")
		}
	}
	if _, _, err := Decode(nil, nil); err == nil {
		t.Fatal("nil accepted")
	}
	unknown, _ := structpb.NewStruct(map[string]any{"unknown": true})
	if _, _, err := Decode(unknown, nil); err == nil {
		t.Fatal("unknown config fields accepted")
	}
	valid, _ := Canonical(legacy())
	if _, _, err := Decode(valid, unknown); err == nil {
		t.Fatal("unknown resolved fields accepted")
	}
}

func TestLegacyCountriesAndUnspecifiedModes(t *testing.T) {
	cfg := legacy()
	cfg.Trip.Origin.CountryCode = "RU"
	cfg.Trip.Destinations[0].CountryCode = "RU"
	cfg.Trip.Transport.Modes = nil
	value, err := Canonical(cfg)
	if err != nil {
		t.Fatal(err)
	}
	input, _, err := Decode(value, nil)
	if err != nil || input.Origin.CountryCode != "RU" || input.Destination.CountryCode != "RU" || input.Transport == nil || *input.Transport.MaxTransfers != 1 {
		t.Fatalf("constraints dropped: %+v %v", input, err)
	}
	cfg = legacy()
	cfg.Trip.Budget.Amount = -1
	if _, err = Canonical(cfg); err == nil {
		t.Fatal("negative legacy budget silently dropped")
	}
	cfg = legacy()
	cfg.Trip.Budget.Amount = 0
	if _, err = Canonical(cfg); err == nil {
		t.Fatal("hard limit silently dropped")
	}
	cfg = legacy()
	cfg.Trip.Budget.Amount = math.Inf(1)
	if _, err = Canonical(cfg); err == nil {
		t.Fatal("infinite budget accepted")
	}
	invalid := &structpb.Struct{Fields: map[string]*structpb.Value{"x": structpb.NewNumberValue(math.NaN())}}
	if _, _, err = Decode(invalid, nil); err == nil {
		t.Fatal("NaN config accepted")
	}
}
