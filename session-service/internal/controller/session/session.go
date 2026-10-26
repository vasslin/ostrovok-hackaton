package session

import (
	"context"

	pb "github.com/godnota/pkg/generated/session/v1"
	usecase "github.com/godnota/session-service/internal/usecase/session"
)

type sessionService interface {
	CreateSession(context.Context, usecase.CreateRequest) (usecase.CreateResponse, error)
	GetSession(context.Context, usecase.GetRequest) (usecase.GetResponse, error)
	GetContext(context.Context, usecase.GetRequest) (usecase.GetResponse, error)
	SaveContext(context.Context, usecase.SaveRequest) error
}

type Controller struct {
	service sessionService
}

func New(service sessionService) *Controller { return &Controller{service: service} }

var _ pb.SessionServiceServer = (*Controller)(nil)

func (c *Controller) CreateSession(ctx context.Context, req *pb.CreateSessionRequest) (*pb.CreateSessionResponse, error) {
	if req != nil {
		_, _ = c.service.CreateSession(ctx, usecase.CreateRequest{UserID: req.GetUserId()})
	}
	return &pb.CreateSessionResponse{}, nil
}

func (c *Controller) GetSession(ctx context.Context, req *pb.GetSessionRequest) (*pb.GetSessionResponse, error) {
	if req != nil {
		_, _ = c.service.GetSession(ctx, usecase.GetRequest{SessionID: req.GetSessionId()})
	}
	return &pb.GetSessionResponse{}, nil
}

func (c *Controller) GetContext(ctx context.Context, req *pb.GetContextRequest) (*pb.GetContextResponse, error) {
	if req != nil {
		_, _ = c.service.GetContext(ctx, usecase.GetRequest{SessionID: req.GetSessionId()})
	}
	return &pb.GetContextResponse{}, nil
}

func (c *Controller) SaveContext(ctx context.Context, req *pb.SaveContextRequest) (*pb.SaveContextResponse, error) {
	if req != nil {
		_ = c.service.SaveContext(ctx, usecase.SaveRequest{
			SessionID: req.GetSessionId(),
			Summary:   req.GetSummary(),
			Status:    req.GetStatus(),
		})
	}
	return &pb.SaveContextResponse{}, nil
}
