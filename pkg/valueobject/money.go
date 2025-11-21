package valueobject

import (
	"strings"

	pkgErrors "github.com/ftryyln/hotel-booking-microservices/pkg/errors"
)

// Money captures amount and currency with validation.
type Money struct {
	Amount   float64
	Currency string
}

// NewMoney validates amount and currency.
func NewMoney(amount float64, currency string) (Money, error) {
	if amount <= 0 {
		return Money{}, pkgErrors.New("bad_request", "amount must be positive")
	}
	if strings.TrimSpace(currency) == "" {
		return Money{}, pkgErrors.New("bad_request", "currency required")
	}
	return Money{Amount: amount, Currency: strings.TrimSpace(currency)}, nil
}
