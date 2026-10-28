package main

import (
	"context"
	"log/slog"
	"net"
	"os"

	aiv1 "github.com/godnota/pkg/generated/ai/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func main() {
	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50055"
	}
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		slog.Error("listen", "error", err)
		os.Exit(1)
	}
	server := grpc.NewServer()
	aiv1.RegisterTravelAIServer(server, mock{})
	slog.Info("ai mock listening", "port", port)
	if err = server.Serve(listener); err != nil {
		slog.Error("serve", "error", err)
		os.Exit(1)
	}
}

type mock struct{}

func (mock) ExtractConfig(context.Context, *aiv1.ExtractConfigRequest) (*aiv1.ExtractConfigResponse, error) {
	return &aiv1.ExtractConfigResponse{}, nil
}

func (mock) RankJoinRoutes(_ context.Context, req *aiv1.RankJoinRoutesRequest) (*aiv1.RankJoinRoutesResponse, error) {
	slog.Info("mock rank join", "join_id", req.GetJoinId(), "routes", len(req.GetRoutes()))
	return &aiv1.RankJoinRoutesResponse{Routes: req.GetRoutes()}, nil
}

func (mock) RankRouteParts(_ context.Context, req *aiv1.RankRoutePartsRequest) (*aiv1.RankRoutePartsResponse, error) {
	routes := req.GetRoutes()
	slog.Info("mock rank routes", "session_id", req.GetSessionId(), "routes", len(routes))
	if len(routes) == 0 {
		return &aiv1.RankRoutePartsResponse{}, nil
	}
	last := proto.Clone(routes[len(routes)-1]).(*structpb.Struct)
	if last.Fields == nil {
		last.Fields = map[string]*structpb.Value{}
	}
	last.Fields["rank_source"] = structpb.NewStringValue("mock")
	return &aiv1.RankRoutePartsResponse{Routes: []*structpb.Struct{last}}, nil
}
