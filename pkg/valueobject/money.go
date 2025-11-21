package valueobject

import (
	"fmt"
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
	if amount < 0 {
		return Money{}, pkgErrors.New("bad_request", "amount cannot be negative")
	}
	cur := strings.TrimSpace(currency)
	if cur == "" {
		return Money{}, pkgErrors.New("bad_request", "currency cannot be empty")
	}
	return Money{Amount: amount, Currency: strings.ToUpper(cur)}, nil
}


// Add adds two Money values (must be same currency).
func (m Money) Add(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, pkgErrors.New("bad_request", "cannot add money with different currencies")
	}
	return Money{Amount: m.Amount + other.Amount, Currency: m.Currency}, nil
}

// Subtract subtracts two Money values (must be same currency).
func (m Money) Subtract(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, pkgErrors.New("bad_request", "cannot subtract money with different currencies")
	}
	result := m.Amount - other.Amount
	if result < 0 {
		return Money{}, pkgErrors.New("bad_request", "result cannot be negative")
	}
	return Money{Amount: result, Currency: m.Currency}, nil
}

// Multiply multiplies the money by a factor.
func (m Money) Multiply(factor float64) (Money, error) {
	if factor < 0 {
		return Money{}, pkgErrors.New("bad_request", "factor cannot be negative")
	}
	return Money{Amount: m.Amount * factor, Currency: m.Currency}, nil
}

// IsZero checks if the amount is zero.
func (m Money) IsZero() bool {
	return m.Amount == 0
}

// IsGreaterThan checks if this money is greater than another.
func (m Money) IsGreaterThan(other Money) bool {
	if m.Currency != other.Currency {
		return false
	}
	return m.Amount > other.Amount
}

// String returns a string representation of the money.
func (m Money) String() string {
	return fmt.Sprintf("%.2f %s", m.Amount, m.Currency)
}
