package route

import (
	"context"

	pb "github.com/godnota/pkg/generated/route/v1"
	"github.com/godnota/route-service/internal/entity"
	"github.com/godnota/route-service/internal/join"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

type routeService interface {
	SearchRoutes(context.Context) ([]entity.Variant, error)
}

// Ranker asks TravelAI to order routes inside one join, then across all joins.
type Ranker interface {
	RankJoinRoutes(context.Context, string, *structpb.Struct, string, []*structpb.Struct) ([]*structpb.Struct, error)
	RankRouteParts(context.Context, string, *structpb.Struct, []*structpb.Struct) ([]*structpb.Struct, error)
}

type Controller struct {
	service       routeService
	rank          Ranker
	maxJoins      int
	routesPerJoin int
	threshold     int
}

func New(service routeService) *Controller {
	return &Controller{service: service, maxJoins: 4, routesPerJoin: 3, threshold: 5}
}

func (c *Controller) WithRanker(rank Ranker, maxJoins, routesPerJoin, threshold int) *Controller {
	c.rank = rank
	if maxJoins > 0 {
		c.maxJoins = maxJoins
	}
	if routesPerJoin > 0 {
		c.routesPerJoin = routesPerJoin
	}
	if threshold > 0 {
		c.threshold = threshold
	}
	return c
}

var _ pb.RouteServiceServer = (*Controller)(nil)

func (c *Controller) BuildRoutes(ctx context.Context, req *pb.BuildRoutesRequest) (*pb.BuildRoutesResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, status.FromContextError(err).Err()
	}
	if req.GetSessionId() == "" || req.GetConfig() == nil || req.GetParts() == nil {
		return nil, status.Error(codes.InvalidArgument, "session_id, config and parts required")
	}
	joins, err := join.Assemble(req.GetParts(), originCity(req.GetConfig()), c.maxJoins)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	payload, err := c.rankJoins(ctx, req.GetSessionId(), req.GetConfig(), joins)
	if err != nil {
		return nil, err
	}
	if len(payload) == 0 {
		return &pb.BuildRoutesResponse{}, nil
	}
	ranked, err := c.rank.RankRouteParts(ctx, req.GetSessionId(), req.GetConfig(), payload)
	if err != nil {
		return nil, err
	}
	if len(ranked) > c.threshold {
		ranked = ranked[:c.threshold]
	}
	return &pb.BuildRoutesResponse{Candidates: ranked}, nil
}

func (c *Controller) rankJoins(ctx context.Context, sessionID string, config *structpb.Struct, joins []join.Join) ([]*structpb.Struct, error) {
	var payload []*structpb.Struct
	for _, item := range joins {
		if err := ctx.Err(); err != nil {
			return nil, status.FromContextError(err).Err()
		}
		documents, err := routeDocuments(join.Expand(item))
		if err != nil {
			return nil, err
		}
		if len(documents) == 0 {
			continue
		}
		if c.rank == nil {
			return nil, status.Error(codes.FailedPrecondition, "ai-service is not configured")
		}
		ranked, err := c.rank.RankJoinRoutes(ctx, sessionID, config, item.ID, documents)
		if err != nil {
			return nil, err
		}
		if len(ranked) > c.routesPerJoin {
			ranked = ranked[:c.routesPerJoin]
		}
		payload = append(payload, ranked...)
	}
	return payload, nil
}

func routeDocuments(routes []join.Route) ([]*structpb.Struct, error) {
	payload := make([]*structpb.Struct, 0, len(routes))
	for _, route := range routes {
		document, err := route.Proto()
		if err != nil {
			return nil, status.Error(codes.Internal, "route document is invalid")
		}
		payload = append(payload, document)
	}
	return payload, nil
}

func originCity(config *structpb.Struct) string {
	origin := config.GetFields()["origin"].GetStructValue()
	if origin == nil {
		return ""
	}
	return origin.GetFields()["city"].GetStringValue()
}

func (c *Controller) SearchRoutes(ctx context.Context, _ *pb.SearchRoutesRequest) (*pb.SearchRoutesResponse, error) {
	_, _ = c.service.SearchRoutes(ctx)
	return &pb.SearchRoutesResponse{}, nil
}
