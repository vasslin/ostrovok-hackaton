package session

import (
	"context"
	"errors"
	"time"

	pb "github.com/godnota/pkg/generated/session/v1"
	"github.com/godnota/session-service/internal/entity"
	usecase "github.com/godnota/session-service/internal/usecase/session"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (c *Controller) GetAgentContext(ctx context.Context, req *pb.GetAgentContextRequest) (*pb.GetAgentContextResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	got, err := c.service.GetAgentContext(ctx, usecase.AgentContextRequest{
		UserID: req.GetUserId(),
		ChatID: req.GetChatId(),
		Limit:  int(req.GetPreviousChatsLimit()),
	})
	if err != nil {
		return nil, rpcError(err)
	}
	previous := make([]*pb.PreviousChatSummary, 0, len(got.Previous))
	for _, item := range got.Previous {
		previous = append(previous, &pb.PreviousChatSummary{
			ChatId:    item.ChatID,
			Summary:   item.Summary,
			UpdatedAt: item.UpdatedAt.UTC().Format(time.RFC3339Nano),
		})
	}
	return &pb.GetAgentContextResponse{
		UserId:               got.UserID,
		ChatId:               got.ChatID,
		ChatHistory:          messages(got.History),
		ChatSummary:          got.Summary,
		PreviousChatsSummary: previous,
		Cfg:                  got.Config,
	}, nil
}

func (c *Controller) SaveContext(ctx context.Context, req *pb.SaveContextRequest) (*pb.SaveContextResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	stored := make([]entity.Message, 0, len(req.GetMessages()))
	for _, message := range req.GetMessages() {
		stored = append(stored, entity.Message{Role: message.GetRole(), Text: message.GetText(), CreatedAt: message.GetCreatedAt()})
	}
	err := c.service.SaveContext(ctx, usecase.SaveRequest{
		UserID:      req.GetUserId(),
		ChatID:      req.GetChatId(),
		Messages:    stored,
		SetMessages: len(stored) > 0,
		Summary:     req.GetSummary(),
		SetSummary:  req.Summary != nil,
		Config:      req.GetCfg(),
		SetConfig:   req.Cfg != nil,
	})
	if err != nil {
		return nil, rpcError(err)
	}
	return &pb.SaveContextResponse{UserId: req.GetUserId(), ChatId: req.GetChatId()}, nil
}

func rpcError(err error) error {
	switch {
	case errors.Is(err, usecase.ErrInvalid):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, usecase.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, usecase.ErrForbidden):
		return status.Error(codes.PermissionDenied, err.Error())
	default:
		return err
	}
}
