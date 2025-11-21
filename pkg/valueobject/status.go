package valueobject

import pkgErrors "github.com/ftryyln/hotel-booking-microservices/pkg/errors"

// BookingStatus captures allowed booking states.
type BookingStatus string

const (
	StatusPendingPayment BookingStatus = "pending_payment"
	StatusConfirmed      BookingStatus = "confirmed"
	StatusCancelled      BookingStatus = "cancelled"
	StatusCheckedIn      BookingStatus = "checked_in"
	StatusCompleted      BookingStatus = "completed"
)

// ValidateBookingStatus ensures status is known.
func ValidateBookingStatus(status string) (BookingStatus, error) {
	switch BookingStatus(status) {
	case StatusPendingPayment, StatusConfirmed, StatusCancelled, StatusCheckedIn, StatusCompleted:
		return BookingStatus(status), nil
	default:
		return "", pkgErrors.New("bad_request", "invalid booking status")
	}
}

// CanTransition checks allowed booking transitions.
func (s BookingStatus) CanTransition(target BookingStatus) error {
	switch s {
	case StatusPendingPayment:
		if target == StatusConfirmed || target == StatusCancelled {
			return nil
		}
	case StatusConfirmed:
		if target == StatusCheckedIn || target == StatusCancelled {
			return nil
		}
	case StatusCheckedIn:
		if target == StatusCompleted || target == StatusCancelled {
			return nil
		}
	}
	return pkgErrors.New("bad_request", "booking cannot transition to target status")
}

// PaymentStatus captures allowed payment states.
type PaymentStatus string

const (
	PaymentPending PaymentStatus = "pending"
	PaymentPaid    PaymentStatus = "paid"
	PaymentFailed  PaymentStatus = "failed"
)

// ValidatePaymentStatus ensures status is known.
func ValidatePaymentStatus(status string) (PaymentStatus, error) {
	switch PaymentStatus(status) {
	case PaymentPending, PaymentPaid, PaymentFailed:
		return PaymentStatus(status), nil
	default:
		return "", pkgErrors.New("bad_request", "invalid payment status")
	}
}

// CanTransition checks allowed payment transitions.
func (s PaymentStatus) CanTransition(target PaymentStatus) error {
	switch s {
	case PaymentPending:
		if target == PaymentPaid || target == PaymentFailed {
			return nil
		}
	}
	return pkgErrors.New("bad_request", "payment cannot transition to target status")
}
