package grpc

import (
	"context"

	mapv1 "github.com/godnota/pkg/generated/map/v1"
)

type Client struct {
	api mapv1.MapServiceClient
}

func New(api mapv1.MapServiceClient) *Client {
	return &Client{api: api}
}

func (c *Client) MapToProviderConfigs(ctx context.Context) error {
	_, err := c.api.MapToProviderConfigs(ctx, &mapv1.MapToProviderConfigsRequest{})
	return err
}

func (c *Client) MapRoutesToFrontend(ctx context.Context) error {
	_, err := c.api.MapRoutesToFrontend(ctx, &mapv1.MapRoutesToFrontendRequest{})
	return err
}
