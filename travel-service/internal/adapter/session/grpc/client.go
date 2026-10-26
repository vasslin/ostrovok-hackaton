package grpc

import (
	"context"

	sessionv1 "github.com/godnota/pkg/generated/session/v1"
)

type Client struct {
	api sessionv1.SessionServiceClient
}

func New(api sessionv1.SessionServiceClient) *Client {
	return &Client{api: api}
}

func (c *Client) GetContext(ctx context.Context, sessionID string) error {
	_, err := c.api.GetContext(ctx, &sessionv1.GetContextRequest{SessionId: sessionID})
	return err
}

func (c *Client) SaveContext(ctx context.Context, sessionID string) error {
	_, err := c.api.SaveContext(ctx, &sessionv1.SaveContextRequest{SessionId: sessionID})
	return err
}
