package grpcapi

import (
	aiv1 "ai-service/api/ai/v1"
	travelv1 "ai-service/api/travel/v1"
	"ai-service/internal/service"
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthv1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type Server struct {
	travelv1.UnimplementedTravelAIServer
	service *service.Service
}

func New(svc *service.Service, key string) *grpc.Server {
	interceptor := func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if key != "" && (strings.HasPrefix(info.FullMethod, "/travel.v1.TravelAI/") || strings.HasPrefix(info.FullMethod, "/godnota.ai.v1.AIService/")) {
			md, _ := metadata.FromIncomingContext(ctx)
			values := md.Get("authorization")
			if len(values) != 1 || subtle.ConstantTimeCompare([]byte(values[0]), []byte("Bearer "+key)) != 1 {
				return nil, status.Error(codes.Unauthenticated, "unauthorized")
			}
		}
		return handler(ctx, req)
	}

	server := grpc.NewServer(grpc.MaxRecvMsgSize(32<<10), grpc.MaxSendMsgSize(1<<20), grpc.UnaryInterceptor(interceptor))
	travelv1.RegisterTravelAIServer(server, &Server{service: svc})
	aiv1.RegisterAIServiceServer(server, &planningServer{service: svc})
	healthServer := health.NewServer()
	healthServer.SetServingStatus("", healthv1.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("travel.v1.TravelAI", healthv1.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("godnota.ai.v1.AIService", healthv1.HealthCheckResponse_SERVING)
	healthv1.RegisterHealthServer(server, healthServer)
	reflection.Register(server)
	return server
}

func input(message proto.Message, target any) error {
	data, err := (protojson.MarshalOptions{UseProtoNames: true}).Marshal(message)
	if err != nil {
		return status.Error(codes.InvalidArgument, "invalid request")
	}

	if len(data) > 32<<10 {
		return status.Error(codes.ResourceExhausted, "JSON request exceeds 32 KiB")
	}

	if err := json.Unmarshal(data, target); err != nil {
		return status.Error(codes.InvalidArgument, "invalid request")
	}
	return nil
}

func output(value any, target proto.Message) error {
	data, err := json.Marshal(value)
	if err != nil {
		return status.Error(codes.Internal, "response encoding failed")
	}

	if err := protojson.Unmarshal(data, target); err != nil {
		return status.Error(codes.Internal, "response encoding failed")
	}
	return nil
}

func convertError(ctx context.Context, err error) error {
	var e *service.Error
	if !errors.As(err, &e) {
		return status.Error(codes.Internal, "internal error")
	}

	code := codes.Internal
	switch e.Status {
	case 400:
		code = codes.InvalidArgument
	case 429:
		code = codes.ResourceExhausted
	case 499:
		code = codes.Canceled
	case 502, 503:
		code = codes.Unavailable
	case 504:
		code = codes.DeadlineExceeded
	}

	if e.RetryAfter != "" {
		_ = grpc.SetTrailer(ctx, metadata.Pairs("retry-after", e.RetryAfter))
	}
	return status.Error(code, e.Message)
}
func (s *Server) ExtractConfig(ctx context.Context, req *travelv1.ExtractConfigRequest) (*travelv1.ExtractConfigResponse, error) {
	var in service.ExtractInput
	if err := input(req, &in); err != nil {
		return nil, err
	}

	result, err := s.service.Extract(ctx, in)
	if err != nil {
		return nil, convertError(ctx, err)
	}

	out := new(travelv1.ExtractConfigResponse)
	if err := output(result, out); err != nil {
		return nil, err
	}

	return out, nil
}
func (s *Server) RankHotels(ctx context.Context, req *travelv1.RankHotelsRequest) (*travelv1.RankHotelsResponse, error) {
	var in service.RankInput
	if err := input(req, &in); err != nil {
		return nil, err
	}

	result, err := s.service.RankHotels(ctx, in)
	if err != nil {
		return nil, convertError(ctx, err)
	}

	out := new(travelv1.RankHotelsResponse)
	if err := output(result, out); err != nil {
		return nil, err
	}

	return out, nil
}
