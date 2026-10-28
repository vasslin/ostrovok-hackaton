package user

import (
	"context"

	pb "github.com/godnota/pkg/generated/user/v1"
	"github.com/godnota/user-service/internal/entity"
	usecase "github.com/godnota/user-service/internal/usecase/user"
)

type userService interface {
	GetUser(context.Context, usecase.GetRequest) (entity.User, error)
	UpdatePreferences(context.Context, usecase.PreferencesRequest) error
}

type Controller struct {
	service userService
}

func New(service userService) *Controller { return &Controller{service: service} }

var _ pb.UserServiceServer = (*Controller)(nil)

func (c *Controller) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	if req != nil {
		_, _ = c.service.GetUser(ctx, usecase.GetRequest{UserID: req.GetUserId()})
	}
	return &pb.GetUserResponse{}, nil
}

func (c *Controller) UpdatePreferences(ctx context.Context, req *pb.UpdatePreferencesRequest) (*pb.UpdatePreferencesResponse, error) {
	if req != nil {
		_ = c.service.UpdatePreferences(ctx, usecase.PreferencesRequest{
			UserID:   req.GetUserId(),
			Language: req.GetLanguage(),
			Currency: req.GetCurrency(),
		})
	}
	return &pb.UpdatePreferencesResponse{}, nil
}
