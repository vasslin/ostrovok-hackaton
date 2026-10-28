package route

import (
	"context"
	providerv1 "github.com/godnota/pkg/generated/provider/v1"
	pb "github.com/godnota/pkg/generated/route/v1"
	"github.com/godnota/pkg/testkit"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"testing"
)

func TestStage3BoundaryIsExplicit(t *testing.T) {
	conn := testkit.Conn(t, func(s *grpc.Server) { pb.RegisterRouteServiceServer(s, New(nil)) })
	client := pb.NewRouteServiceClient(conn)
	ctx := context.Background()
	if _, err := client.BuildRoutes(ctx, &pb.BuildRoutesRequest{}); status.Code(err) != codes.InvalidArgument {
		t.Fatal(err)
	}
	config, _ := structpb.NewStruct(map[string]any{"schema_version": "1.0"})
	if _, err := client.BuildRoutes(ctx, &pb.BuildRoutesRequest{SessionId: "s", Config: config, Parts: &providerv1.SearchPartsResponse{}}); status.Code(err) != codes.InvalidArgument {
		t.Fatal("invalid trip config accepted", err)
	}
}
