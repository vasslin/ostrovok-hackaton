package adapter

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/godnota/gateway/internal/httpapi"
	orderv1 "github.com/godnota/pkg/generated/order/v1"
	sessionv1 "github.com/godnota/pkg/generated/session/v1"
	travelv1 "github.com/godnota/pkg/generated/travel/v1"
	userv1 "github.com/godnota/pkg/generated/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

const (
	invalidRequest = "invalid request"
	requestTimeout = 5 * time.Second
	maxStreams     = 100
)

type backend struct {
	travel  travelv1.TravelServiceClient
	session sessionv1.SessionServiceClient
	user    userv1.UserServiceClient
	order   orderv1.OrderServiceClient
	conns   []*grpc.ClientConn
	slots   chan struct{}
}

func New(travelAddr, sessionAddr, userAddr, orderAddr string) (*backend, func(), error) {
	travelConn, err := dial(travelAddr)
	if err != nil {
		return nil, nil, err
	}
	sessionConn, err := dial(sessionAddr)
	if err != nil {
		_ = travelConn.Close()
		return nil, nil, err
	}
	userConn, err := dial(userAddr)
	if err != nil {
		_ = travelConn.Close()
		_ = sessionConn.Close()
		return nil, nil, err
	}
	orderConn, err := dial(orderAddr)
	if err != nil {
		_ = travelConn.Close()
		_ = sessionConn.Close()
		_ = userConn.Close()
		return nil, nil, err
	}

	conns := []*grpc.ClientConn{travelConn, sessionConn, userConn, orderConn}
	b := &backend{
		travel:  travelv1.NewTravelServiceClient(travelConn),
		session: sessionv1.NewSessionServiceClient(sessionConn),
		user:    userv1.NewUserServiceClient(userConn),
		order:   orderv1.NewOrderServiceClient(orderConn),
		conns:   conns,
		slots:   make(chan struct{}, maxStreams),
	}
	return b, b.close, nil
}

func dial(target string) (*grpc.ClientConn, error) {
	return grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
}

func (b *backend) close() {
	for _, conn := range b.conns {
		_ = conn.Close()
	}
}

func (b *backend) Call(ctx context.Context, path string, raw json.RawMessage) (json.RawMessage, error) {
	if err := b.acquire(ctx); err != nil {
		return nil, callError(err)
	}
	defer b.release()

	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	switch path {
	case "/v1/session/create":
		return unary(ctx, raw, &sessionv1.CreateSessionRequest{}, b.session.CreateSession)
	case "/v1/session/get":
		return unary(ctx, raw, &sessionv1.GetSessionRequest{}, b.session.GetSession)

	case "/v1/travel/plan":
		return unary(ctx, raw, &travelv1.PlanTripRequest{}, b.travel.PlanTrip)
	case "/v1/travel/answers":
		return unary(ctx, raw, &travelv1.SubmitAnswersRequest{}, b.travel.SubmitAnswers)
	case "/v1/travel/idea":
		return unary(ctx, raw, &travelv1.SelectIdeaRequest{}, b.travel.SelectIdea)
	case "/v1/travel/config/get":
		return unary(ctx, raw, &travelv1.GetConfigRequest{}, b.travel.GetConfig)
	case "/v1/travel/config/update":
		return unary(ctx, raw, &travelv1.UpdateConfigRequest{}, b.travel.UpdateConfig)
	case "/v1/travel/config/confirm":
		return unary(ctx, raw, &travelv1.ConfirmConfigRequest{}, b.travel.ConfirmConfig)
	case "/v1/travel/search":
		return unary(ctx, raw, &travelv1.SearchRoutesRequest{}, b.travel.SearchRoutes)
	case "/v1/travel/routes":
		return unary(ctx, raw, &travelv1.GetRoutesRequest{}, b.travel.GetRoutes)

	case "/v1/order/create":
		return unary(ctx, raw, &orderv1.CreateOrderRequest{}, b.order.CreateOrder)
	case "/v1/order/get":
		return unary(ctx, raw, &orderv1.GetOrderRequest{}, b.order.GetOrder)
	case "/v1/order/pay":
		return unary(ctx, raw, &orderv1.PayOrderRequest{}, b.order.PayOrder)
	case "/v1/order/cancel":
		return unary(ctx, raw, &orderv1.CancelOrderRequest{}, b.order.CancelOrder)

	case "/v1/user/get":
		return unary(ctx, raw, &userv1.GetUserRequest{}, b.user.GetUser)
	case "/v1/user/update":
		return unary(ctx, raw, &userv1.UpdatePreferencesRequest{}, b.user.UpdatePreferences)

	default:
		return nil, &httpapi.Error{Code: int(codes.Unimplemented), Message: "operation is not implemented"}
	}
}

func unary[RES, REQ proto.Message](
	ctx context.Context,
	raw json.RawMessage,
	req REQ,
	grpcCall func(context.Context, REQ, ...grpc.CallOption) (RES, error),
) (json.RawMessage, error) {
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(raw, req); err != nil {
		return nil, callError(status.Error(codes.InvalidArgument, invalidRequest))
	}

	res, err := grpcCall(ctx, req)
	if err != nil {
		return nil, callError(err)
	}

	resJSON, err := protojson.MarshalOptions{EmitUnpopulated: true, UseProtoNames: true}.Marshal(res)
	if err != nil {
		return nil, callError(status.Error(codes.Internal, "failed to marshal response"))
	}
	return resJSON, nil
}

func (b *backend) acquire(ctx context.Context) error {
	select {
	case b.slots <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (b *backend) release() { <-b.slots }

func callError(err error) error {
	if err == nil || errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return err
	}
	st, ok := status.FromError(err)
	if !ok || st.Code() == codes.OK {
		return &httpapi.Error{Code: int(codes.Internal), Message: "internal error"}
	}
	switch st.Code() {
	case codes.DeadlineExceeded:
		return context.DeadlineExceeded
	case codes.Canceled:
		return context.Canceled
	default:
		return &httpapi.Error{Code: int(st.Code()), Message: st.Message()}
	}
}
