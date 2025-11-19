package utils

import (
	"errors"
	"time"
)

var ErrInvalidDateRange = errors.New("checkout must be after checkin")

// NightsBetween calculates nights between two times.
func NightsBetween(start, end time.Time) (int, error) {
	if end.Before(start) {
		return 0, ErrInvalidDateRange
	}
	nights := int(end.Sub(start).Hours() / 24)
	if nights <= 0 {
		return 0, ErrInvalidDateRange
	}
	return nights, nil
}
