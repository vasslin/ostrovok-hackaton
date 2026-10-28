package order

import (
	"context"
	"sync"
	"testing"

	repo "github.com/godnota/order-service/internal/repository/order"
	uc "github.com/godnota/order-service/internal/usecase/order"
	pb "github.com/godnota/pkg/generated/order/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestDemoOrderLifecycleAndIdempotency(t *testing.T) {
	c := New(uc.New(repo.New()))
	snapshot, _ := structpb.NewStruct(map[string]any{"id": "route-a", "bookable": false})
	req := &pb.CreateOrderRequest{UserId: "user", SessionId: "session", RouteId: "route-a", IdempotencyKey: "key", RouteSnapshot: snapshot}
	ids := make(chan string, 20)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, e := c.CreateOrder(context.Background(), req)
			if e != nil {
				t.Error(e)
				return
			}
			ids <- v.GetOrderId()
		}()
	}
	wg.Wait()
	close(ids)
	var id string
	for x := range ids {
		if id != "" && x != id {
			t.Fatal("duplicate demo order")
		}
		id = x
	}
	if id == "" {
		t.Fatal("order missing")
	}
	read, e := c.GetOrder(context.Background(), &pb.GetOrderRequest{OrderId: id})
	if e != nil || read.GetStatus() != "demo_booked" || read.GetMode() != "demo" || read.GetRouteSnapshot().GetFields()["id"].GetStringValue() != "route-a" {
		t.Fatalf("bad order %+v %v", read, e)
	}
	snapshot.Fields["id"] = structpb.NewStringValue("mutated")
	read, e = c.GetOrder(context.Background(), &pb.GetOrderRequest{OrderId: id})
	if e != nil || read.GetRouteSnapshot().GetFields()["id"].GetStringValue() != "route-a" {
		t.Fatal("snapshot mutated")
	}
	req.RouteId = "other"
	req.RouteSnapshot, _ = structpb.NewStruct(map[string]any{"id": "other"})
	if _, e = c.CreateOrder(context.Background(), req); status.Code(e) != codes.AlreadyExists {
		t.Fatal("key reuse accepted", e)
	}
	for i := 0; i < 2; i++ {
		v, e := c.PayOrder(context.Background(), &pb.PayOrderRequest{OrderId: id})
		if e != nil || v.GetStatus() != "demo_paid" {
			t.Fatal(e)
		}
	}
	for i := 0; i < 2; i++ {
		v, e := c.CancelOrder(context.Background(), &pb.CancelOrderRequest{OrderId: id})
		if e != nil || v.GetStatus() != "demo_cancelled" {
			t.Fatal(e)
		}
	}
	if _, e = c.PayOrder(context.Background(), &pb.PayOrderRequest{OrderId: id}); status.Code(e) != codes.FailedPrecondition {
		t.Fatal("paid cancelled order", e)
	}
	if _, e = c.GetOrder(context.Background(), &pb.GetOrderRequest{OrderId: "missing"}); status.Code(e) != codes.NotFound {
		t.Fatal(e)
	}
}
