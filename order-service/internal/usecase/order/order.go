package order

import (
	"context"
	"errors"
	"strings"

	"github.com/godnota/order-service/internal/entity"
	"google.golang.org/protobuf/types/known/structpb"
)

type Repository interface {
	Create(context.Context, entity.Order) (entity.Order, error)
	Find(context.Context, string) (entity.Order, error)
	Transition(context.Context, string, string) (entity.Order, error)
}
type Service struct{ repo Repository }

func New(repo Repository) *Service { return &Service{repo: repo} }

type CreateRequest struct {
	UserID, SessionID, RouteID, IdempotencyKey string
	Snapshot                                   *structpb.Struct
}
type IDRequest struct{ OrderID string }

var ErrInput = errors.New("user, session, route, idempotency key and matching route snapshot are required")

func (s *Service) CreateOrder(ctx context.Context, r CreateRequest) (entity.Order, error) {
	if strings.TrimSpace(r.UserID) == "" || strings.TrimSpace(r.SessionID) == "" || strings.TrimSpace(r.RouteID) == "" || strings.TrimSpace(r.IdempotencyKey) == "" || len(r.IdempotencyKey) > 128 || r.Snapshot == nil || r.Snapshot.GetFields()["id"].GetStringValue() != r.RouteID {
		return entity.Order{}, ErrInput
	}
	return s.repo.Create(ctx, entity.Order{UserID: r.UserID, SessionID: r.SessionID, RouteID: r.RouteID, IdempotencyKey: r.IdempotencyKey, Snapshot: r.Snapshot})
}
func (s *Service) GetOrder(ctx context.Context, r IDRequest) (entity.Order, error) {
	if strings.TrimSpace(r.OrderID) == "" {
		return entity.Order{}, ErrInput
	}
	return s.repo.Find(ctx, r.OrderID)
}
func (s *Service) PayOrder(ctx context.Context, r IDRequest) (entity.Order, error) {
	if strings.TrimSpace(r.OrderID) == "" {
		return entity.Order{}, ErrInput
	}
	return s.repo.Transition(ctx, r.OrderID, entity.StatusPaid)
}
func (s *Service) CancelOrder(ctx context.Context, r IDRequest) (entity.Order, error) {
	if strings.TrimSpace(r.OrderID) == "" {
		return entity.Order{}, ErrInput
	}
	return s.repo.Transition(ctx, r.OrderID, entity.StatusCancelled)
}
