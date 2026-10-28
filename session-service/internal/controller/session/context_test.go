package session

import (
	"context"
	"testing"

	commonv1 "github.com/godnota/pkg/generated/common/v1"
	pb "github.com/godnota/pkg/generated/session/v1"
	"github.com/godnota/pkg/testkit"
	historymem "github.com/godnota/session-service/internal/repository/history/memory"
	repo "github.com/godnota/session-service/internal/repository/session"
	configmem "github.com/godnota/session-service/internal/repository/tripconfig/memory"
	usecase "github.com/godnota/session-service/internal/usecase/session"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAgentContextRPC(t *testing.T) {
	conn := testkit.Conn(t, func(server *grpc.Server) {
		pb.RegisterSessionServiceServer(server, New(usecase.New(repo.New(), historymem.New(), configmem.New())))
	})
	client := pb.NewSessionServiceClient(conn)
	ctx := context.Background()
	summary := func(text string) *string { return &text }
	for _, chatID := range []string{"c1", "c2", "c3", "c4", "c5"} {
		if _, err := client.SaveContext(ctx, &pb.SaveContextRequest{UserId: "owner", ChatId: chatID, Summary: summary(chatID)}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := client.SaveContext(ctx, &pb.SaveContextRequest{
		UserId: "owner", ChatId: "current",
		Messages: []*pb.DialogMessage{{Role: "user", Text: "хочу к морю"}},
		Summary:  summary("море"),
		Cfg:      &commonv1.TripConfig{Version: "1.0"},
	}); err != nil {
		t.Fatal(err)
	}
	got, err := client.GetAgentContext(ctx, &pb.GetAgentContextRequest{UserId: "owner", ChatId: "current"})
	if err != nil || got.GetChatSummary() != "море" || len(got.GetChatHistory()) != 1 || got.GetCfg().GetVersion() != "1.0" || len(got.GetPreviousChatsSummary()) != 4 {
		t.Fatalf("%+v %v", got, err)
	}
	if _, err = client.SaveContext(ctx, &pb.SaveContextRequest{UserId: "owner", ChatId: "current", Summary: summary("только резюме")}); err != nil {
		t.Fatal(err)
	}
	got, err = client.GetAgentContext(ctx, &pb.GetAgentContextRequest{UserId: "owner", ChatId: "current"})
	if err != nil || got.GetChatSummary() != "только резюме" || len(got.GetChatHistory()) != 1 || got.GetCfg().GetVersion() != "1.0" {
		t.Fatalf("untouched fields changed: %+v %v", got, err)
	}
	if _, err = client.SaveContext(ctx, &pb.SaveContextRequest{UserId: "owner", ChatId: "current", Messages: []*pb.DialogMessage{{Role: "user"}}}); status.Code(err) != codes.InvalidArgument {
		t.Fatal(err)
	}
	if _, err = client.SaveContext(ctx, &pb.SaveContextRequest{UserId: "thief", ChatId: "current", Cfg: &commonv1.TripConfig{Version: "9"}}); status.Code(err) != codes.PermissionDenied {
		t.Fatal(err)
	}
}
