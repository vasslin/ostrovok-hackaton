package entity

// Status values follow the MVP order lifecycle.
const (
	StatusBooked    = "booked"
	StatusPaid      = "paid"
	StatusCancelled = "cancelled"
)

// Order is a saved route selected by the traveller.
type Order struct {
	ID        string
	UserID    string
	SessionID string
	RouteID   string
	Status    string
}
