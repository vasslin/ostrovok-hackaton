package grpc

import (
	"context"

	routev1 "github.com/godnota/pkg/generated/route/v1"
)

type Client struct {
	api routev1.RouteServiceClient
}

func New(api routev1.RouteServiceClient) *Client {
	return &Client{api: api}
}

func (c *Client) SearchRoutes(ctx context.Context) error {
	_, err := c.api.SearchRoutes(ctx, &routev1.SearchRoutesRequest{})
	return err
}
