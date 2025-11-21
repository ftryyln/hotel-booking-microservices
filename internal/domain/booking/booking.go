package booking

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/ftryyln/hotel-booking-microservices/pkg/query"
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
	Guests      int
	TotalPrice  float64
	TotalNights int
	CreatedAt   time.Time
}

// Repository handles persistence.
type Repository interface {
	Create(ctx context.Context, b Booking) error
	FindByID(ctx context.Context, id uuid.UUID) (Booking, error)
	List(ctx context.Context, opts query.Options) ([]Booking, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
}

// PaymentGateway used by booking service.
type PaymentGateway interface {
	Initiate(ctx context.Context, bookingID uuid.UUID, amount float64) (PaymentResult, error)
}

// NotificationGateway for events.
type NotificationGateway interface {
	Notify(ctx context.Context, event string, payload any) error
}

// PaymentResult carries minimal payment data after initiation.
type PaymentResult struct {
	ID         uuid.UUID
	Status     string
	Provider   string
	PaymentURL string
}
