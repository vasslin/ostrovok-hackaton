package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	pb "github.com/godnota/pkg/generated/provider/v1"
	"github.com/godnota/provider-service/internal/entity"
	"github.com/ostrovok-hackathon-2026/godnota/providerkit"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

type providerService interface {
	Search(context.Context, providerkit.TripRequest, providerkit.ResolvedTrip, string) (entity.Result, error)
}

type Controller struct {
	pb.UnimplementedProviderServiceServer
	service  providerService
	searcher *providerkit.Searcher
}

func New(service providerService) *Controller { return &Controller{service: service} }

func decode(in *structpb.Struct, dst any) error {
	if in == nil {
		return fmt.Errorf("trip_request is required")
	}
	data, err := json.Marshal(in.AsMap())
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}

func (c *Controller) search(ctx context.Context, input *pb.SearchInput, kind string) (entity.Result, *structpb.Struct, error) {
	if input == nil {
		return entity.Result{}, nil, status.Error(codes.InvalidArgument, "input is required")
	}
	var trip providerkit.TripRequest
	var resolved providerkit.ResolvedTrip
	var adapted providerkit.AdaptedTrip
	switch input.InputFormat {
	case "", "trip_request":
		if err := decode(input.TripRequest, &trip); err != nil {
			return entity.Result{}, nil, status.Errorf(codes.InvalidArgument, "invalid trip_request: %v", err)
		}
	case "ai_config":
		var config providerkit.AIConfig
		if err := decode(input.TripRequest, &config); err != nil {
			return entity.Result{}, nil, status.Errorf(codes.InvalidArgument, "invalid AI config: %v", err)
		}
		var err error
		adapted, err = providerkit.FromAIConfig(config)
		if err != nil {
			return entity.Result{}, nil, status.Errorf(codes.InvalidArgument, "invalid AI config: %v", err)
		}
		trip = adapted.Trip
	case "frontend_criteria":
		var criteria providerkit.FrontendCriteria
		if err := decode(input.TripRequest, &criteria); err != nil {
			return entity.Result{}, nil, status.Errorf(codes.InvalidArgument, "invalid frontend criteria: %v", err)
		}
		var err error
		adapted, err = providerkit.FromFrontendCriteria(criteria)
		if err != nil {
			return entity.Result{}, nil, status.Errorf(codes.InvalidArgument, "invalid frontend criteria: %v", err)
		}
		trip = adapted.Trip
	default:
		return entity.Result{}, nil, status.Error(codes.InvalidArgument, "unknown input_format")
	}
	if input.Resolved != nil {
		if err := decode(input.Resolved, &resolved); err != nil {
			return entity.Result{}, nil, status.Errorf(codes.InvalidArgument, "invalid resolved: %v", err)
		}
	}
	if input.InputFormat != "" && input.InputFormat != "trip_request" {
		missing := append(adapted.Unresolved, trip.MissingRequired()...)
		if trip.Destination == nil || trip.Destination.City == "" {
			missing = append(missing, "/destination/city")
		}
		if len(missing) > 0 {
			result := entity.Result{Plan: providerkit.TripPlan{Missing: missing, Limitations: adapted.Limitations}}
			return result, &structpb.Struct{}, nil
		}
	}
	result, err := c.service.Search(ctx, trip, resolved, kind)
	if err != nil {
		if ctx.Err() != nil {
			return entity.Result{}, nil, status.FromContextError(ctx.Err()).Err()
		}
		return entity.Result{}, nil, status.Errorf(codes.InvalidArgument, "invalid search: %v", err)
	}
	result.Plan.Limitations = append(result.Plan.Limitations, adapted.Limitations...)
	data, err := json.Marshal(result.Plan.Discovery)
	if err != nil {
		return entity.Result{}, nil, status.Error(codes.Internal, "cannot encode discovery")
	}
	var object map[string]any
	if err := json.Unmarshal(data, &object); err != nil {
		return entity.Result{}, nil, status.Error(codes.Internal, "cannot encode discovery")
	}
	discovery, err := structpb.NewStruct(object)
	if err != nil {
		return entity.Result{}, nil, status.Error(codes.Internal, "cannot encode discovery")
	}
	return result, discovery, nil
}

func (c *Controller) SearchTransport(ctx context.Context, req *pb.SearchTransportRequest) (*pb.SearchTransportResponse, error) {
	r, d, err := c.search(ctx, req.GetInput(), "transport")
	if err != nil {
		return nil, err
	}
	return &pb.SearchTransportResponse{Offers: r.Offers, Issues: r.Issues, Missing: r.Plan.Missing, Limitations: r.Plan.Limitations, Discovery: d}, nil
}

func (c *Controller) SearchHotels(ctx context.Context, req *pb.SearchHotelsRequest) (*pb.SearchHotelsResponse, error) {
	r, d, err := c.search(ctx, req.GetInput(), "hotel")
	if err != nil {
		return nil, err
	}
	return &pb.SearchHotelsResponse{Offers: r.Offers, Issues: r.Issues, Missing: r.Plan.Missing, Limitations: r.Plan.Limitations, Discovery: d}, nil
}

func (c *Controller) SearchActivities(ctx context.Context, req *pb.SearchActivitiesRequest) (*pb.SearchActivitiesResponse, error) {
	r, d, err := c.search(ctx, req.GetInput(), "activity")
	if err != nil {
		return nil, err
	}
	return &pb.SearchActivitiesResponse{Offers: r.Offers, Issues: r.Issues, Missing: r.Plan.Missing, Limitations: r.Plan.Limitations, Discovery: d}, nil
}
