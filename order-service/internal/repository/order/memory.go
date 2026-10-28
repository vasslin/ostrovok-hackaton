package order

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"

	"github.com/godnota/order-service/internal/entity"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

var ErrNotFound = errors.New("order not found")
var ErrConflict = errors.New("idempotency key is already used for another order")
var ErrState = errors.New("order state does not allow this action")

type Memory struct {
	mu     sync.Mutex
	orders map[string]entity.Order
	keys   map[string]string
}

func New() *Memory { return &Memory{orders: map[string]entity.Order{}, keys: map[string]string{}} }
func copyOrder(v entity.Order) entity.Order {
	if v.Snapshot != nil {
		v.Snapshot = proto.Clone(v.Snapshot).(*structpb.Struct)
	}
	return v
}
func (m *Memory) Create(ctx context.Context, v entity.Order) (entity.Order, error) {
	if err := ctx.Err(); err != nil {
		return entity.Order{}, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	key := v.UserID + "\x00" + v.IdempotencyKey
	if id, ok := m.keys[key]; ok {
		old := m.orders[id]
		if old.SessionID != v.SessionID || old.RouteID != v.RouteID || !proto.Equal(old.Snapshot, v.Snapshot) {
			return entity.Order{}, ErrConflict
		}
		return copyOrder(old), nil
	}
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return entity.Order{}, err
	}
	v.ID = hex.EncodeToString(b[:])
	v.Status = entity.StatusBooked
	m.orders[v.ID] = copyOrder(v)
	m.keys[key] = v.ID
	return copyOrder(v), nil
}
func (m *Memory) Find(ctx context.Context, id string) (entity.Order, error) {
	if err := ctx.Err(); err != nil {
		return entity.Order{}, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.orders[id]
	if !ok {
		return entity.Order{}, ErrNotFound
	}
	return copyOrder(v), nil
}
func (m *Memory) Transition(ctx context.Context, id, target string) (entity.Order, error) {
	if err := ctx.Err(); err != nil {
		return entity.Order{}, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.orders[id]
	if !ok {
		return entity.Order{}, ErrNotFound
	}
	if v.Status == target {
		return copyOrder(v), nil
	}
	if target == entity.StatusPaid && v.Status != entity.StatusBooked {
		return entity.Order{}, ErrState
	}
	if target == entity.StatusCancelled && v.Status != entity.StatusBooked && v.Status != entity.StatusPaid {
		return entity.Order{}, ErrState
	}
	v.Status = target
	m.orders[id] = v
	return copyOrder(v), nil
}
