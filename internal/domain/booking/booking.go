package booking

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const (
	StatusPendingPayment = "pending_payment"
	StatusConfirmed      = "confirmed"
	StatusCancelled      = "cancelled"
	StatusCheckedIn      = "checked_in"
	StatusCompleted      = "completed"
)

// Booking aggregate.
type Booking struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	RoomTypeID  uuid.UUID
	CheckIn     time.Time
	CheckOut    time.Time
	Status      string
	TotalPrice  float64
	TotalNights int
	CreatedAt   time.Time
}

// Repository handles persistence.
type Repository interface {
	Create(ctx context.Context, b Booking) error
	FindByID(ctx context.Context, id uuid.UUID) (Booking, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
}

// PaymentGateway used by booking service.
type PaymentGateway interface {
	Initiate(ctx context.Context, bookingID uuid.UUID, amount float64) (string, error)
}

// NotificationGateway for events.
type NotificationGateway interface {
	Notify(ctx context.Context, event string, payload any) error
}
