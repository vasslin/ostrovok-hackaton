package order

import (
	"context"
	"log/slog"

	"github.com/godnota/order-service/internal/entity"
)

type Repository interface {
	Create(ctx context.Context, order entity.Order) (entity.Order, error)
	Find(ctx context.Context, orderID string) (entity.Order, error)
	Save(ctx context.Context, order entity.Order) error
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service { return &Service{repo: repo} }

type CreateRequest struct {
	UserID    string
	SessionID string
	RouteID   string
}

type IDRequest struct {
	OrderID string
}

func (s *Service) CreateOrder(ctx context.Context, req CreateRequest) (entity.Order, error) {
	slog.Info("order-service: create order", "user_id", req.UserID, "session_id", req.SessionID, "route_id", req.RouteID)
	_, _ = s.repo.Create(ctx, entity.Order{UserID: req.UserID, SessionID: req.SessionID, RouteID: req.RouteID})
	return entity.Order{}, nil
}

func (s *Service) GetOrder(ctx context.Context, req IDRequest) (entity.Order, error) {
	slog.Info("order-service: get order", "order_id", req.OrderID)
	_, _ = s.repo.Find(ctx, req.OrderID)
	return entity.Order{}, nil
}

func (s *Service) PayOrder(ctx context.Context, req IDRequest) (entity.Order, error) {
	slog.Info("order-service: pay order", "order_id", req.OrderID)
	_ = s.repo.Save(ctx, entity.Order{ID: req.OrderID, Status: entity.StatusPaid})
	return entity.Order{}, nil
}

func (s *Service) CancelOrder(ctx context.Context, req IDRequest) (entity.Order, error) {
	slog.Info("order-service: cancel order", "order_id", req.OrderID)
	_ = s.repo.Save(ctx, entity.Order{ID: req.OrderID, Status: entity.StatusCancelled})
	return entity.Order{}, nil
}
