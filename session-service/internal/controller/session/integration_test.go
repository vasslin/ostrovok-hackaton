package session

import (
	"context"
	pb "github.com/godnota/pkg/generated/session/v1"
	"github.com/godnota/pkg/testkit"
	historymem "github.com/godnota/session-service/internal/repository/history/memory"
	repo "github.com/godnota/session-service/internal/repository/session"
	configmem "github.com/godnota/session-service/internal/repository/tripconfig/memory"
	usecase "github.com/godnota/session-service/internal/usecase/session"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
)

func TestGRPCSessionLifecycle(t *testing.T) {
	conn := testkit.Conn(t, func(server *grpc.Server) {
		pb.RegisterSessionServiceServer(server, New(usecase.New(repo.New(), historymem.New(), configmem.New())))
	})
	client := pb.NewSessionServiceClient(conn)
	ctx := context.Background()
	if _, err := client.CreateSession(ctx, &pb.CreateSessionRequest{}); status.Code(err) != codes.InvalidArgument {
		t.Fatal(err)
	}
	created, err := client.CreateSession(ctx, &pb.CreateSessionRequest{UserId: "owner"})
	if err != nil || created.GetSessionId() == "" {
		t.Fatal(err)
	}
	id := created.GetSessionId()
	got, err := client.GetSession(ctx, &pb.GetSessionRequest{SessionId: id})
	if err != nil || got.GetUserId() != "owner" || got.GetStatus() != "draft" || got.GetCreatedAt() == "" {
		t.Fatalf("state lost: %+v %v", got, err)
	}
	if _, err = client.GetSession(ctx, &pb.GetSessionRequest{SessionId: "missing"}); status.Code(err) != codes.NotFound {
		t.Fatal(err)
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err = client.GetSession(canceled, &pb.GetSessionRequest{SessionId: id}); status.Code(err) != codes.Canceled {
		t.Fatal(err)
	}
}
