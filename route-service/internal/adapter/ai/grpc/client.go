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

// RankJoinRoutes ranks the valid routes of a single join. An empty stub response keeps that order.
func (c *Client) RankJoinRoutes(ctx context.Context, sessionID string, config *structpb.Struct, joinID string, routes []*structpb.Struct) ([]*structpb.Struct, error) {
	response, err := c.api.RankJoinRoutes(ctx, &aiv1.RankJoinRoutesRequest{SessionId: sessionID, Config: config, JoinId: joinID, Routes: routes})
	if err != nil {
		return nil, err
	}
	if response != nil && len(response.GetRoutes()) > 0 {
		return response.GetRoutes(), nil
	}
	return routes, nil
}

// RankRouteParts sends every assembled route. An empty stub response keeps that order.
func (c *Client) RankRouteParts(ctx context.Context, sessionID string, config *structpb.Struct, routes []*structpb.Struct) ([]*structpb.Struct, error) {
	response, err := c.api.RankRouteParts(ctx, &aiv1.RankRoutePartsRequest{SessionId: sessionID, Config: config, Routes: routes})
	if err != nil {
		return nil, err
	}
	if response != nil && len(response.GetRoutes()) > 0 {
		return response.GetRoutes(), nil
	}
	return routes, nil
}
