package grpc

import (
	"context"

	aiv1 "github.com/godnota/pkg/generated/ai/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

type Client struct {
	api aiv1.TravelAIClient
}

func New(api aiv1.TravelAIClient) *Client {
	return &Client{api: api}
}

// RankRouteParts sends the assembled routes. The TravelAI server is a stub and does not reorder them.
func (c *Client) RankRouteParts(ctx context.Context, sessionID string, config *structpb.Struct, routes []*structpb.Struct) ([]*structpb.Struct, error) {
	_, err := c.api.RankRouteParts(ctx, &aiv1.RankRoutePartsRequest{SessionId: sessionID, Config: config, Routes: routes})
	if err != nil {
		return nil, err
	}
	return routes, nil
}
