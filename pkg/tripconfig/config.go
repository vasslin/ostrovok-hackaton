// Package tripconfig bridges protobuf messages to the canonical AI JSON schema.
package tripconfig

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	commonv1 "github.com/godnota/pkg/generated/common/v1"
	"github.com/ostrovok-hackathon-2026/godnota/providerkit"
	"google.golang.org/protobuf/types/known/structpb"
	"io"
	"strings"
)

func Decode(config, resolved *structpb.Struct) (providerkit.TripRequest, providerkit.ResolvedTrip, error) {
	var places providerkit.ResolvedTrip
	if config == nil {
		return providerkit.TripRequest{}, places, errors.New("config required")
	}
	raw, err := config.MarshalJSON()
	if err != nil {
		return providerkit.TripRequest{}, places, err
	}
	input, err := providerkit.DecodeTrip(raw)
	if err != nil {
		return input, places, err
	}
	if resolved != nil {
		raw, err = resolved.MarshalJSON()
		if err != nil {
			return input, places, err
		}
		dec := json.NewDecoder(bytes.NewReader(raw))
		dec.DisallowUnknownFields()
		if err = dec.Decode(&places); err != nil {
			return input, places, err
		}
		if dec.Decode(new(any)) != io.EOF {
			return input, places, errors.New("invalid resolved places")
		}
	}
	return input, places, nil
}
func Canonical(config *commonv1.TripConfig) (*structpb.Struct, error) {
	if config == nil {
		return nil, errors.New("config required")
	}
	if config.GetRequest() != nil {
		if config.GetTrip() != nil {
			return nil, errors.New("use request or legacy trip, not both")
		}
		return config.GetRequest(), nil
	}
	t := config.GetTrip()
	if t == nil {
		return nil, errors.New("config.request required")
	}
	if len(t.GetDestinations()) != 1 {
		return nil, errors.New("stage 2 requires exactly one selected destination")
	}
	if int(t.GetTravelers().GetChildren()) != len(t.GetTravelers().GetChildrenAges()) {
		return nil, errors.New("children count must match children_ages")
	}
	currency := t.GetBudget().GetCurrency()
	if currency == "" {
		currency = "RUB"
	}
	locale := t.GetPreferences().GetLanguage()
	if locale == "" {
		locale = "ru"
	}
	version := config.GetVersion()
	if version == "" {
		version = "1.0"
	}
	children := []any{}
	for _, age := range t.GetTravelers().GetChildrenAges() {
		children = append(children, float64(age))
	}
	obj := map[string]any{"schema_version": version, "locale": locale, "currency": currency, "origin": map[string]any{"city": t.GetOrigin().GetCity()}, "destination": map[string]any{"city": t.GetDestinations()[0].GetCity()}, "dates": map[string]any{"departure_date": t.GetDateRange().GetStartDate(), "return_date": t.GetDateRange().GetEndDate()}, "travelers": map[string]any{"adults": t.GetTravelers().GetAdults(), "children_ages": children}}
	if code := t.GetOrigin().GetCountryCode(); code != "" {
		obj["origin"].(map[string]any)["country_code"] = code
	}
	if code := t.GetDestinations()[0].GetCountryCode(); code != "" {
		obj["destination"].(map[string]any)["country_code"] = code
	}
	if a := t.GetAccommodation(); a != nil {
		obj["hotel"] = map[string]any{"needed": a.GetRequired(), "star_rating_min": a.GetMinStars(), "kinds": a.GetPreferredTypes()}
		if a.GetMaxPricePerNight() > 0 {
			obj["hotel"].(map[string]any)["price_per_night_max"] = a.GetMaxPricePerNight()
		}
	}
	if p := t.GetTransport(); p != nil {
		modes := []any{}
		for _, mode := range p.GetModes() {
			modes = append(modes, mode)
		}
		obj["transport"] = map[string]any{"max_transfers": p.GetMaxTransfers(), "direct_only": p.GetMaxTransfers() == 0}
		if len(modes) > 0 {
			obj["transport"].(map[string]any)["modes"] = modes
		}
	}
	if t.GetDateRange().GetFlexibleDates() || t.GetTransport().GetMaxWaitHours() != 0 || t.GetPreferences().GetAvoidOvernightTransfers() || len(t.GetActivities().GetInterests()) > 0 || t.GetActivities().GetPace() != "" {
		return nil, fmt.Errorf("legacy preferences cannot be mapped losslessly; use config.request (%s)", strings.Join([]string{"dates", "transport", "leisure"}, ", "))
	}
	// Preserve the budget rather than silently dropping a hard user constraint.
	if b := t.GetBudget(); b != nil && (b.GetAmount() < 0 || b.GetAmount() == 0 && b.GetIsHardLimit()) {
		return nil, errors.New("positive amount required for a hard budget")
	}
	if b := t.GetBudget(); b != nil && b.GetAmount() > 0 {
		obj["budget"] = map[string]any{"total": b.GetAmount(), "scope": b.GetScope(), "is_hard": b.GetIsHardLimit()}
	}
	raw, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	var plain map[string]any
	if err = json.Unmarshal(raw, &plain); err != nil {
		return nil, err
	}
	return structpb.NewStruct(plain)
}
