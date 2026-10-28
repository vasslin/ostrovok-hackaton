package grpc

import (
	"context"

	aiv1 "github.com/godnota/pkg/generated/ai/v1"
)

type Client struct {
	api aiv1.AIServiceClient
}

func New(api aiv1.AIServiceClient) *Client {
	return &Client{api: api}
}

func (c *Client) GenerateOrUpdateConfig(ctx context.Context) error {
	_, err := c.api.GenerateOrUpdateConfig(ctx, &aiv1.GenerateOrUpdateConfigRequest{})
	return err
}

func (c *Client) UpdateConfig(ctx context.Context) error {
	_, err := c.api.UpdateConfig(ctx, &aiv1.UpdateConfigRequest{})
	return err
}

func (c *Client) RefineConfig(ctx context.Context) error {
	_, err := c.api.RefineConfig(ctx, &aiv1.RefineConfigRequest{})
	return err
}

func (c *Client) RankRoutes(ctx context.Context) error {
	_, err := c.api.RankRoutes(ctx, &aiv1.RankRoutesRequest{})
	return err
}
