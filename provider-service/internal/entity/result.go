package entity

import (
	pb "github.com/godnota/pkg/generated/provider/v1"
	"github.com/ostrovok-hackathon-2026/godnota/providerkit"
)

type Result struct {
	Offers []*pb.Offer
	Issues []*pb.ProviderIssue
	Plan   providerkit.TripPlan
}
