package order

import (
	"context"

	"github.com/godnota/order-service/internal/entity"
	usecase "github.com/godnota/order-service/internal/usecase/order"
	pb "github.com/godnota/pkg/generated/order/v1"
)

type orderService interface {
	CreateOrder(context.Context, usecase.CreateRequest) (entity.Order, error)
	GetOrder(context.Context, usecase.IDRequest) (entity.Order, error)
	PayOrder(context.Context, usecase.IDRequest) (entity.Order, error)
	CancelOrder(context.Context, usecase.IDRequest) (entity.Order, error)
}

type Controller struct {
	service orderService
}

func New(service orderService) *Controller { return &Controller{service: service} }

var _ pb.OrderServiceServer = (*Controller)(nil)

func (c *Controller) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	if req != nil {
		_, _ = c.service.CreateOrder(ctx, usecase.CreateRequest{
			UserID:    req.GetUserId(),
			SessionID: req.GetSessionId(),
			RouteID:   req.GetRouteId(),
		})
	}
	return &pb.CreateOrderResponse{}, nil
}

func (c *Controller) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.GetOrderResponse, error) {
	if req != nil {
		_, _ = c.service.GetOrder(ctx, usecase.IDRequest{OrderID: req.GetOrderId()})
	}
	return &pb.GetOrderResponse{}, nil
}

func (c *Controller) PayOrder(ctx context.Context, req *pb.PayOrderRequest) (*pb.PayOrderResponse, error) {
	if req != nil {
		_, _ = c.service.PayOrder(ctx, usecase.IDRequest{OrderID: req.GetOrderId()})
	}
	return &pb.PayOrderResponse{}, nil
}

func (c *Controller) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*pb.CancelOrderResponse, error) {
	if req != nil {
		_, _ = c.service.CancelOrder(ctx, usecase.IDRequest{OrderID: req.GetOrderId()})
	}
	return &pb.CancelOrderResponse{}, nil
}
