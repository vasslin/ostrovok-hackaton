package travel

import (
	"context"
	pb "github.com/godnota/pkg/generated/travel/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

func (c *PlanningController) GetCapabilities(ctx context.Context, req *pb.GetConfigRequest) (*structpb.Struct, error) {
	if _, err := c.load(ctx, req.GetUserId(), req.GetSessionId()); err != nil {
		return nil, err
	}
	return c.providers.GetCapabilities(ctx, &structpb.Struct{})
}
