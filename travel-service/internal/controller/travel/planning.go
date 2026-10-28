package travel

import (
	"context"
	commonv1 "github.com/godnota/pkg/generated/common/v1"
	providerv1 "github.com/godnota/pkg/generated/provider/v1"
	routev1 "github.com/godnota/pkg/generated/route/v1"
	sessionv1 "github.com/godnota/pkg/generated/session/v1"
	pb "github.com/godnota/pkg/generated/travel/v1"
	"github.com/godnota/pkg/tripconfig"
	"github.com/ostrovok-hackathon-2026/godnota/providerkit"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"sync"
	"time"
)

// PlanningController implements the stage-2 path; stage-1 handlers remain on Controller.
type PlanningController struct {
	*Controller
	sessions  sessionv1.SessionServiceClient
	providers providerv1.ProviderServiceClient
	routes    routev1.RouteServiceClient
	mu        sync.Mutex
	locks     map[string]*sync.Mutex
	cache     map[string]cachedParts
}
type cachedParts struct {
	config     *commonv1.TripConfig
	parts      *providerv1.SearchPartsResponse
	candidates []*structpb.Struct
}

func cloneCandidates(items []*structpb.Struct) []*structpb.Struct {
	result := make([]*structpb.Struct, 0, len(items))
	for _, item := range items {
		if item != nil {
			result = append(result, proto.Clone(item).(*structpb.Struct))
		}
	}
	return result
}

func (c *Controller) WithPlanning(s sessionv1.SessionServiceClient, p providerv1.ProviderServiceClient, routeClients ...routev1.RouteServiceClient) *PlanningController {
	controller := &PlanningController{Controller: c, sessions: s, providers: p, locks: map[string]*sync.Mutex{}, cache: map[string]cachedParts{}}
	if len(routeClients) > 0 {
		controller.routes = routeClients[0]
	}
	return controller
}
func (c *PlanningController) lock(id string) func() {
	c.mu.Lock()
	m := c.locks[id]
	if m == nil {
		m = &sync.Mutex{}
		c.locks[id] = m
	}
	c.mu.Unlock()
	m.Lock()
	return m.Unlock
}
func (c *PlanningController) context(ctx context.Context, id string) (*sessionv1.GetContextResponse, error) {
	if id == "" {
		return nil, status.Error(codes.InvalidArgument, "session_id required")
	}
	return c.sessions.GetContext(ctx, &sessionv1.GetContextRequest{SessionId: id})
}
func (c *PlanningController) GetConfig(ctx context.Context, r *pb.GetConfigRequest) (*pb.GetConfigResponse, error) {
	s, err := c.context(ctx, r.GetSessionId())
	if err != nil {
		return nil, err
	}
	return &pb.GetConfigResponse{SessionId: r.GetSessionId(), Config: s.GetTripConfig()}, nil
}
func (c *PlanningController) save(ctx context.Context, id string, config *commonv1.TripConfig, confirmed bool) error {
	current, err := c.context(ctx, id)
	if err != nil {
		return err
	}
	if config == nil {
		config = current.GetTripConfig()
	}
	canonical, err := tripconfig.Canonical(config)
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	input, _, err := tripconfig.Decode(canonical, config.GetResolved())
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	if confirmed {
		err = providerkit.ValidateSearch(input, time.Now())
	} else {
		err = input.ValidateAt(time.Now())
	}
	if err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	state := "draft"
	if confirmed {
		state = "confirmed"
	}
	_, err = c.sessions.SaveContext(ctx, &sessionv1.SaveContextRequest{SessionId: id, TripConfig: config, Summary: current.GetSummary(), Messages: current.GetMessages(), Status: state})
	if err == nil {
		c.mu.Lock()
		delete(c.cache, id)
		c.mu.Unlock()
	}
	return err
}
func (c *PlanningController) UpdateConfig(ctx context.Context, r *pb.UpdateConfigRequest) (*pb.UpdateConfigResponse, error) {
	unlock := c.lock(r.GetSessionId())
	defer unlock()
	if r.GetConfig() == nil {
		return nil, status.Error(codes.InvalidArgument, "config required")
	}
	if err := c.save(ctx, r.GetSessionId(), r.GetConfig(), false); err != nil {
		return nil, err
	}
	return &pb.UpdateConfigResponse{SessionId: r.GetSessionId(), Config: r.GetConfig()}, nil
}
func (c *PlanningController) ConfirmConfig(ctx context.Context, r *pb.ConfirmConfigRequest) (*pb.ConfirmConfigResponse, error) {
	unlock := c.lock(r.GetSessionId())
	defer unlock()
	if err := c.save(ctx, r.GetSessionId(), r.GetConfig(), true); err != nil {
		return nil, err
	}
	return &pb.ConfirmConfigResponse{SessionId: r.GetSessionId(), Confirmed: true}, nil
}
func (c *PlanningController) SearchRoutes(ctx context.Context, r *pb.SearchRoutesRequest) (*pb.SearchRoutesResponse, error) {
	unlock := c.lock(r.GetSessionId())
	defer unlock()
	s, err := c.context(ctx, r.GetSessionId())
	if err != nil {
		return nil, err
	}
	if s.GetStatus() != "confirmed" {
		return nil, status.Error(codes.FailedPrecondition, "confirm trip config before search")
	}
	canonical, err := tripconfig.Canonical(s.GetTripConfig())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	parts, err := c.providers.SearchParts(ctx, &providerv1.SearchPartsRequest{Config: canonical, Resolved: s.GetTripConfig().GetResolved()})
	if err != nil {
		return nil, err
	}
	if parts == nil {
		return nil, status.Error(codes.DataLoss, "provider returned no response")
	}
	var candidates []*structpb.Struct
	if len(parts.GetTransportParts()) > 0 || len(parts.GetHotelParts()) > 0 {
		if c.routes == nil {
			parts.Limitations = append(parts.Limitations, "route-service: handoff not configured")
		} else {
			handoff, cancel := context.WithTimeout(ctx, 750*time.Millisecond)
			assembled, handoffErr := c.routes.BuildRoutes(handoff, &routev1.BuildRoutesRequest{SessionId: r.GetSessionId(), Config: canonical, Parts: proto.Clone(parts).(*providerv1.SearchPartsResponse)})
			cancel()
			if handoffErr == nil && assembled == nil {
				handoffErr = status.Error(codes.DataLoss, "empty route response")
			}
			if handoffErr != nil {
				if ctx.Err() != nil {
					return nil, status.FromContextError(ctx.Err()).Err()
				}
				parts.Limitations = append(parts.Limitations, "route-service: "+status.Code(handoffErr).String()+"; stage-2 parts preserved")
			} else {
				candidates = cloneCandidates(assembled.GetCandidates())
				if len(candidates) == 0 {
					parts.Limitations = append(parts.Limitations, "route-service: no compatible complete routes; stage-2 parts preserved")
				}
			}
		}
	}
	c.mu.Lock()
	c.cache[r.GetSessionId()] = cachedParts{config: proto.Clone(s.GetTripConfig()).(*commonv1.TripConfig), parts: proto.Clone(parts).(*providerv1.SearchPartsResponse), candidates: cloneCandidates(candidates)}
	c.mu.Unlock()
	return &pb.SearchRoutesResponse{SessionId: r.GetSessionId(), Parts: parts, Candidates: candidates}, nil
}
func (c *PlanningController) GetRoutes(ctx context.Context, r *pb.GetRoutesRequest) (*pb.GetRoutesResponse, error) {
	unlock := c.lock(r.GetSessionId())
	defer unlock()
	s, err := c.context(ctx, r.GetSessionId())
	if err != nil {
		return nil, err
	}
	c.mu.Lock()
	cached, ok := c.cache[r.GetSessionId()]
	c.mu.Unlock()
	if !ok || s.GetStatus() != "confirmed" || !proto.Equal(cached.config, s.GetTripConfig()) {
		return nil, status.Error(codes.NotFound, "no search results for current config")
	}
	return &pb.GetRoutesResponse{SessionId: r.GetSessionId(), Parts: proto.Clone(cached.parts).(*providerv1.SearchPartsResponse), Candidates: cloneCandidates(cached.candidates)}, nil
}
