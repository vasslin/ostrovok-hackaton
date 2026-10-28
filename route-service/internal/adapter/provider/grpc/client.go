package grpc

import (
	"context"

	providerv1 "github.com/godnota/pkg/generated/provider/v1"
)

type Client struct {
	api providerv1.ProviderServiceClient
}

func New(api providerv1.ProviderServiceClient) *Client {
	return &Client{api: api}
}

func (c *Client) SearchTransport(ctx context.Context) error {
	_, err := c.api.SearchTransport(ctx, &providerv1.SearchTransportRequest{})
	return err
}

func (c *Client) SearchHotels(ctx context.Context) error {
	_, err := c.api.SearchHotels(ctx, &providerv1.SearchHotelsRequest{})
	return err
}

func (c *Client) SearchActivities(ctx context.Context) error {
	_, err := c.api.SearchActivities(ctx, &providerv1.SearchActivitiesRequest{})
	return err
}
