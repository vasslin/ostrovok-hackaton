package session

import (
	"context"
	commonv1 "github.com/godnota/pkg/generated/common/v1"
	pb "github.com/godnota/pkg/generated/session/v1"
	"github.com/godnota/pkg/testkit"
	repo "github.com/godnota/session-service/internal/repository/session"
	usecase "github.com/godnota/session-service/internal/usecase/session"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
)

func TestGRPCSessionLifecycle(t *testing.T) {
	conn := testkit.Conn(t, func(server *grpc.Server) { pb.RegisterSessionServiceServer(server, New(usecase.New(repo.New()))) })
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
	config := &commonv1.TripConfig{Version: "1.0"}
	messages := []*pb.DialogMessage{{Role: "user", Text: "family trip", CreatedAt: "2026-10-28T12:00:00Z"}}
	if _, err = client.SaveContext(ctx, &pb.SaveContextRequest{SessionId: id, Summary: "family", Status: "confirmed", TripConfig: config, Messages: messages}); err != nil {
		t.Fatal(err)
	}
	got, err := client.GetSession(ctx, &pb.GetSessionRequest{SessionId: id})
	if err != nil || got.GetUserId() != "owner" || got.GetSummary() != "family" || got.GetTripConfig().GetVersion() != "1.0" || len(got.GetMessages()) != 1 || got.GetCreatedAt() == "" {
		t.Fatalf("state lost: %+v %v", got, err)
	}
	contextValue, err := client.GetContext(ctx, &pb.GetContextRequest{SessionId: id})
	if err != nil || contextValue.GetStatus() != "confirmed" || contextValue.GetUserId() != "owner" {
		t.Fatal(err)
	}
	if _, err = client.GetSession(ctx, &pb.GetSessionRequest{SessionId: "missing"}); status.Code(err) != codes.NotFound {
		t.Fatal(err)
	}
	if _, err = client.GetContext(ctx, &pb.GetContextRequest{SessionId: "missing"}); status.Code(err) != codes.NotFound {
		t.Fatal(err)
	}
	if _, err = client.SaveContext(ctx, &pb.SaveContextRequest{SessionId: "missing"}); status.Code(err) != codes.NotFound {
		t.Fatal(err)
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err = client.GetContext(canceled, &pb.GetContextRequest{SessionId: id}); status.Code(err) != codes.Canceled {
		t.Fatal(err)
	}
}
