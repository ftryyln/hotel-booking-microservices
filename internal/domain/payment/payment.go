package payment

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const (
	StatusPending = "pending"
	StatusPaid    = "paid"
	StatusFailed  = "failed"
)

// Payment aggregates payment state.
type Payment struct {
	ID         uuid.UUID
	BookingID  uuid.UUID
	Amount     float64
	Currency   string
	Status     string
	Provider   string
	PaymentURL string
	CreatedAt  time.Time
}

// Provider integrates external gateway.
type Provider interface {
	Initiate(ctx context.Context, payment Payment) (Payment, error)
	VerifySignature(ctx context.Context, payload, signature string) bool
	Refund(ctx context.Context, payment Payment, reason string) (string, error)
}

// Repository persists payments.
type Repository interface {
	Create(ctx context.Context, p Payment) error
	FindByID(ctx context.Context, id uuid.UUID) (Payment, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status, paymentURL string) error
}

// BookingStatusUpdater notifies booking service.
type BookingStatusUpdater interface {
	Update(ctx context.Context, bookingID uuid.UUID, status string) error
}
