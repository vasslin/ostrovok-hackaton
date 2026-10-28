package providerkit

import (
	"time"

	"github.com/ostrovok-hackathon-2026/godnota/internal/providers"
	"github.com/ostrovok-hackathon-2026/godnota/internal/trip"
)

type TripRequest = trip.TripRequest
type Request = providers.Request
type RequestSet = providers.RequestSet
type Registry = providers.Registry
type ResolvedPlace = providers.ResolvedPlace
type ResolvedTrip = providers.ResolvedTrip
type TripPlan = providers.TripPlan

func BuildTripPlan(input TripRequest, resolved ResolvedTrip, now time.Time) (TripPlan, error) {
	return providers.BuildTripPlan(input, resolved, now)
}

func LoadRegistry(path string) (Registry, error) {
	return providers.LoadRegistry(path)
}
