package session

import (
	"context"
	pb "github.com/godnota/pkg/generated/session/v1"
	"github.com/godnota/session-service/internal/entity"
	usecase "github.com/godnota/session-service/internal/usecase/session"
)

type sessionService interface {
	CreateSession(context.Context, usecase.CreateRequest) (usecase.CreateResponse, error)
	GetSession(context.Context, usecase.GetRequest) (usecase.GetResponse, error)
	GetAgentContext(context.Context, usecase.AgentContextRequest) (usecase.AgentContext, error)
	SaveContext(context.Context, usecase.SaveRequest) error
}
type Controller struct{ service sessionService }

func New(s sessionService) *Controller { return &Controller{service: s} }

var _ pb.SessionServiceServer = (*Controller)(nil)

func (c *Controller) CreateSession(ctx context.Context, r *pb.CreateSessionRequest) (*pb.CreateSessionResponse, error) {
	s, e := c.service.CreateSession(ctx, usecase.CreateRequest{UserID: r.GetUserId()})
	if e != nil {
		return nil, e
	}
	return &pb.CreateSessionResponse{SessionId: s.SessionID}, nil
}
func (c *Controller) GetSession(ctx context.Context, r *pb.GetSessionRequest) (*pb.GetSessionResponse, error) {
	s, e := c.service.GetSession(ctx, usecase.GetRequest{SessionID: r.GetSessionId()})
	if e != nil {
		return nil, e
	}
	v := s.Session
	return &pb.GetSessionResponse{SessionId: v.ID, UserId: v.UserID, Status: v.Status, Summary: v.Summary, TripConfig: v.Config, Messages: messages(v.Messages), CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}, nil
}
func messages(input []entity.Message) []*pb.DialogMessage {
	out := make([]*pb.DialogMessage, 0, len(input))
	for _, m := range input {
		out = append(out, &pb.DialogMessage{Role: m.Role, Text: m.Text, CreatedAt: m.CreatedAt})
	}
	return out
}
