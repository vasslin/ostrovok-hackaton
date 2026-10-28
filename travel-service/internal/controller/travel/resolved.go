package travel

import (
	"reflect"

	commonv1 "github.com/godnota/pkg/generated/common/v1"
	"github.com/godnota/pkg/tripconfig"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func modelConfig(config *commonv1.TripConfig) *commonv1.TripConfig {
	if config == nil {
		return nil
	}
	out := proto.Clone(config).(*commonv1.TripConfig)
	out.Resolved = nil
	return out
}

func retainResolved(previous, next *commonv1.TripConfig) *commonv1.TripConfig {
	if next == nil {
		return nil
	}
	out := proto.Clone(next).(*commonv1.TripConfig)
	out.Resolved = nil
	before, e1 := tripconfig.Canonical(previous)
	after, e2 := tripconfig.Canonical(next)
	if e1 != nil || e2 != nil || previous.GetResolved() == nil {
		return out
	}
	oldTrip, _, e1 := tripconfig.Decode(before, nil)
	newTrip, _, e2 := tripconfig.Decode(after, nil)
	if e1 != nil || e2 != nil {
		return out
	}
	values := previous.GetResolved().AsMap()
	if !reflect.DeepEqual(oldTrip.Origin, newTrip.Origin) {
		delete(values, "origin")
	}
	if !reflect.DeepEqual(oldTrip.Destination, newTrip.Destination) {
		delete(values, "destination")
	}
	oldCities, newCities := []string{}, []string{}
	for _, stop := range oldTrip.Stops {
		oldCities = append(oldCities, stop.City)
	}
	for _, stop := range newTrip.Stops {
		newCities = append(newCities, stop.City)
	}
	if !reflect.DeepEqual(oldCities, newCities) {
		delete(values, "stops")
	}
	if oldTrip.Dates != newTrip.Dates || !reflect.DeepEqual(oldCities, newCities) || !reflect.DeepEqual(oldTrip.Destination, newTrip.Destination) {
		delete(values, "hotel_checkin")
		delete(values, "hotel_checkout")
	}
	out.Resolved, _ = structpb.NewStruct(values)
	return out
}
