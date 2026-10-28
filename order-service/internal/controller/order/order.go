package order

import (
	"context"
	"errors"

	"github.com/godnota/order-service/internal/entity"
	repo "github.com/godnota/order-service/internal/repository/order"
	usecase "github.com/godnota/order-service/internal/usecase/order"
	pb "github.com/godnota/pkg/generated/order/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type orderService interface {
	CreateOrder(context.Context, usecase.CreateRequest) (entity.Order, error)
	GetOrder(context.Context, usecase.IDRequest) (entity.Order, error)
	PayOrder(context.Context, usecase.IDRequest) (entity.Order, error)
	CancelOrder(context.Context, usecase.IDRequest) (entity.Order, error)
}
type Controller struct{ service orderService }

func New(service orderService) *Controller { return &Controller{service: service} }

var _ pb.OrderServiceServer = (*Controller)(nil)

func convert(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, usecase.ErrInput):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, repo.ErrNotFound):
		return status.Error(codes.NotFound, "order not found")
	case errors.Is(err, repo.ErrConflict):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, repo.ErrState):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded):
		return status.FromContextError(err).Err()
	default:
		return status.Error(codes.Internal, "order operation failed")
	}
}
func (c *Controller) CreateOrder(ctx context.Context, r *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	v, e := c.service.CreateOrder(ctx, usecase.CreateRequest{UserID: r.GetUserId(), SessionID: r.GetSessionId(), RouteID: r.GetRouteId(), IdempotencyKey: r.GetIdempotencyKey(), Snapshot: r.GetRouteSnapshot()})
	if e != nil {
		return nil, convert(e)
	}
	return &pb.CreateOrderResponse{OrderId: v.ID, Status: v.Status, Mode: "demo"}, nil
}
func (c *Controller) GetOrder(ctx context.Context, r *pb.GetOrderRequest) (*pb.GetOrderResponse, error) {
	v, e := c.service.GetOrder(ctx, usecase.IDRequest{OrderID: r.GetOrderId()})
	if e != nil {
		return nil, convert(e)
	}
	return &pb.GetOrderResponse{OrderId: v.ID, UserId: v.UserID, SessionId: v.SessionID, RouteId: v.RouteID, Status: v.Status, Mode: "demo", RouteSnapshot: v.Snapshot}, nil
}
func (c *Controller) PayOrder(ctx context.Context, r *pb.PayOrderRequest) (*pb.PayOrderResponse, error) {
	v, e := c.service.PayOrder(ctx, usecase.IDRequest{OrderID: r.GetOrderId()})
	if e != nil {
		return nil, convert(e)
	}
	return &pb.PayOrderResponse{OrderId: v.ID, Status: v.Status, Mode: "demo"}, nil
}
func (c *Controller) CancelOrder(ctx context.Context, r *pb.CancelOrderRequest) (*pb.CancelOrderResponse, error) {
	v, e := c.service.CancelOrder(ctx, usecase.IDRequest{OrderID: r.GetOrderId()})
	if e != nil {
		return nil, convert(e)
	}
	return &pb.CancelOrderResponse{OrderId: v.ID, Status: v.Status, Mode: "demo"}, nil
}
