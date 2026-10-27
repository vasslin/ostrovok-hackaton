// Package testkit supplies real in-memory gRPC transports for integration tests.
package testkit

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"net"
	"testing"
)

func Conn(t *testing.T, register func(*grpc.Server), options ...grpc.ServerOption) *grpc.ClientConn {
	t.Helper()
	listener := bufconn.Listen(1 << 20)
	server := grpc.NewServer(options...)
	register(server)
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(func() { server.Stop(); _ = listener.Close() })
	conn, err := grpc.NewClient("passthrough:///in-memory", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return listener.DialContext(ctx) }))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}
