package grpcapp

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"google.golang.org/grpc"
)

// App is a gRPC server process shared by backend services.
type App struct {
	server *grpc.Server
	port   string
	close  func()
}

// New builds a gRPC server. register attaches service implementations.
// close is called after graceful stop and may be nil.
func New(port string, closeFn func(), register func(*grpc.Server)) *App {
	srv := grpc.NewServer(grpc.UnaryInterceptor(LoggingInterceptor))
	register(srv)
	if closeFn == nil {
		closeFn = func() {}
	}
	return &App{server: srv, port: port, close: closeFn}
}

// Run listens until the process is stopped.
func (a *App) Run() error {
	listener, err := net.Listen("tcp", net.JoinHostPort("", a.port))
	if err != nil {
		return fmt.Errorf("listen grpc: %w", err)
	}
	slog.Info("grpc server started", "port", a.port)
	if err = a.server.Serve(listener); err != nil {
		return fmt.Errorf("serve grpc: %w", err)
	}
	return nil
}

// Shutdown stops the server and closes outbound connections.
func (a *App) Shutdown() {
	a.server.GracefulStop()
	a.close()
}

// LoggingInterceptor writes every incoming RPC before the controller runs.
func LoggingInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	slog.Info("grpc request", "method", info.FullMethod, "payload", fmt.Sprint(req))
	resp, err := handler(ctx, req)
	if err != nil {
		slog.Error("grpc response", "method", info.FullMethod, "error", err)
		return resp, err
	}
	slog.Info("grpc response", "method", info.FullMethod, "status", "OK")
	return resp, nil
}
