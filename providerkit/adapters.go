package providerkit

import (
	"fmt"
	"math"
	"strings"

	"github.com/ostrovok-hackathon-2026/godnota/internal/trip"
)

type AIConfig struct {
	Origin        *string  `json:"origin"`
	Destination   *string  `json:"destination"`
	DepartureDate *string  `json:"departure_date"`
	ReturnDate    *string  `json:"return_date"`
	Adults        *int     `json:"adults"`
	ChildrenAges  []int    `json:"children_ages"`
	Budget        *float64 `json:"budget"`
	Currency      *string  `json:"currency"`
	Preferences   []string `json:"preferences"`
}

type FrontendCriteria struct {
	Origin       string `json:"origin"`
	Destination  string `json:"destination"`
	StartDate    string `json:"startDate"`
	Nights       int    `json:"nights"`
	Adults       int    `json:"adults"`
	ChildrenAges []int  `json:"childrenAges"`
	BudgetMinor  int64  `json:"budgetMinor"`
	Currency     string `json:"currency"`
	Pets         *bool  `json:"pets"`
}

type AdaptedTrip struct {
	Trip        TripRequest
	Unresolved  []string
	Limitations []string
}

func FromAIConfig(input AIConfig) (AdaptedTrip, error) {
	result := AdaptedTrip{Trip: trip.TripRequest{SchemaVersion: "1.0", Locale: "ru"}}
	if input.Origin != nil {
		result.Trip.Origin.City = *input.Origin
	}
	if input.Destination != nil {
		result.Trip.Destination = &trip.Destination{City: *input.Destination}
	}
	if input.DepartureDate != nil {
		result.Trip.Dates.DepartureDate = *input.DepartureDate
	}
	if input.ReturnDate != nil {
		result.Trip.Dates.ReturnDate = *input.ReturnDate
	}
	if input.Adults != nil {
		result.Trip.Travelers.Adults = *input.Adults
	}
	for _, age := range input.ChildrenAges {
		result.Trip.Travelers.ChildrenAges = append(result.Trip.Travelers.ChildrenAges, trip.Age(age))
	}
	if input.Currency != nil {
		result.Trip.Currency = strings.ToUpper(*input.Currency)
	}
	if input.Budget != nil {
		if math.IsNaN(*input.Budget) || math.IsInf(*input.Budget, 0) || *input.Budget <= 0 || math.Trunc(*input.Budget) != *input.Budget || *input.Budget > float64(math.MaxInt32) {
			return AdaptedTrip{}, fmt.Errorf("AI budget must be a positive whole major-unit amount")
		}
		result.Trip.Budget = &trip.Budget{Total: int(*input.Budget)}
		result.Unresolved = append(result.Unresolved, "/budget/scope")
	}
	if len(input.Preferences) > 0 {
		notes := strings.Join(input.Preferences, "; ")
		if len([]rune(notes)) > 500 {
			return AdaptedTrip{}, fmt.Errorf("AI preferences exceed 500 characters")
		}
		result.Trip.AIContext = &trip.AIContext{Notes: notes}
		result.Unresolved = append(result.Unresolved, "/ai_context/preferences")
	}
	result.Limitations = append(result.Limitations, "AI config has no structured hotel, transport, pet or country fields; these require explicit confirmation.")
	return result, nil
}

func FromFrontendCriteria(input FrontendCriteria) (AdaptedTrip, error) {
	result := AdaptedTrip{Trip: trip.TripRequest{
		SchemaVersion: "1.0", Locale: "ru", Currency: strings.ToUpper(input.Currency),
		Origin: trip.Place{City: input.Origin}, Dates: trip.Dates{DepartureDate: input.StartDate, Nights: input.Nights},
		Travelers: trip.Travelers{Adults: input.Adults},
	}}
	if input.Destination != "" {
		result.Trip.Destination = &trip.Destination{City: input.Destination}
	}
	for _, age := range input.ChildrenAges {
		result.Trip.Travelers.ChildrenAges = append(result.Trip.Travelers.ChildrenAges, trip.Age(age))
	}
	if input.BudgetMinor < 0 || input.BudgetMinor > int64(math.MaxInt32)*100 {
		return AdaptedTrip{}, fmt.Errorf("invalid frontend budget")
	}
	if input.BudgetMinor > 0 {
		if input.BudgetMinor%100 != 0 {
			return AdaptedTrip{}, fmt.Errorf("frontend budget cannot be represented without losing minor units")
		}
		result.Trip.Budget = &trip.Budget{Total: int(input.BudgetMinor / 100), Scope: "total"}
	}
	if input.Pets == nil {
		result.Unresolved = append(result.Unresolved, "/travelers/pets")
	} else if *input.Pets {
		result.Unresolved = append(result.Unresolved, "/travelers/pets/kind")
		result.Trip.AIContext = &trip.AIContext{Notes: "Traveler requested a pet; species and transport requirements are not specified."}
	}
	result.Limitations = append(result.Limitations, "Frontend criteria omit country, room allocation and guest citizenship; no values were guessed.")
	return result, nil
}
