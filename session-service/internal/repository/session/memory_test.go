package session

import (
	"context"
	commonv1 "github.com/godnota/pkg/generated/common/v1"
	"github.com/godnota/session-service/internal/entity"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
)

func TestStorePreservesOwnerAndCopiesConfig(t *testing.T) {
	m := New()
	ctx := context.Background()
	created, e := m.Create(ctx, entity.Session{UserID: "owner"})
	if e != nil || created.ID == "" {
		t.Fatal(e)
	}
	cfg := &commonv1.TripConfig{Version: "1.0"}
	if e = m.Save(ctx, entity.Session{ID: created.ID, UserID: "attacker", Status: "confirmed", Config: cfg}); e != nil {
		t.Fatal(e)
	}
	cfg.Version = "mutated"
	v, e := m.Find(ctx, created.ID)
	if e != nil || v.UserID != "owner" || v.Config.GetVersion() != "1.0" {
		t.Fatalf("owner or config corrupted: %+v %v", v, e)
	}
	v.Config.Version = "mutated"
	v, _ = m.Find(ctx, created.ID)
	if v.Config.GetVersion() != "1.0" {
		t.Fatal("returned shared state")
	}
	if _, e = m.Find(ctx, "missing"); status.Code(e) != codes.NotFound {
		t.Fatal("missing session accepted")
	}
}
