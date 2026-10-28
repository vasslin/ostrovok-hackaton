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

func (c *Client) GetAgentContext(ctx context.Context, sessionID string) error {
	session, err := c.api.GetSession(ctx, &sessionv1.GetSessionRequest{SessionId: sessionID})
	if err != nil {
		return err
	}
	_, err = c.api.GetAgentContext(ctx, &sessionv1.GetAgentContextRequest{UserId: session.GetUserId(), ChatId: sessionID})
	return err
}

func (c *Client) SaveContext(ctx context.Context, sessionID string) error {
	session, err := c.api.GetSession(ctx, &sessionv1.GetSessionRequest{SessionId: sessionID})
	if err != nil {
		return err
	}
	_, err = c.api.SaveContext(ctx, &sessionv1.SaveContextRequest{UserId: session.GetUserId(), ChatId: sessionID})
	return err
}
