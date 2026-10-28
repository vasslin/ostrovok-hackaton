package entity

import "google.golang.org/protobuf/types/known/structpb"

const (
	StatusBooked    = "demo_booked"
	StatusPaid      = "demo_paid"
	StatusCancelled = "demo_cancelled"
)

type Order struct {
	ID             string
	UserID         string
	SessionID      string
	RouteID        string
	IdempotencyKey string
	Snapshot       *structpb.Struct
	Status         string
}
