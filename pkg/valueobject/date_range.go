package valueobject

import (
	"time"

	pkgErrors "github.com/ftryyln/hotel-booking-microservices/pkg/errors"
)

// DateRange ensures start < end.
type DateRange struct {
	Start time.Time
	End   time.Time
}

// NewDateRange validates and builds a DateRange.
func NewDateRange(start, end time.Time) (DateRange, error) {
	if start.IsZero() || end.IsZero() {
		return DateRange{}, pkgErrors.New("bad_request", "date required")
	}
	if !start.Before(end) {
		return DateRange{}, pkgErrors.New("bad_request", "check_in must be before check_out")
	}
	return DateRange{Start: start, End: end}, nil
}

// Nights returns the full number of nights between start and end.
func (d DateRange) Nights() int {
	return int(endDateOnly(d.End).Sub(endDateOnly(d.Start)).Hours() / 24)
}

func endDateOnly(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}
