package valueobject

import (
	"strings"

	pkgErrors "github.com/ftryyln/hotel-booking-microservices/pkg/errors"
)

// ValidateHotel ensures name/address present.
func ValidateHotel(name, address string) (string, string, error) {
	n := strings.TrimSpace(name)
	a := strings.TrimSpace(address)
	if n == "" {
		return "", "", pkgErrors.New("bad_request", "hotel name required")
	}
	if a == "" {
		return "", "", pkgErrors.New("bad_request", "hotel address required")
	}
	return n, a, nil
}

// RoomTypeSpec validates capacity and base price.
func RoomTypeSpec(capacity int, basePrice float64) error {
	if capacity <= 0 {
		return pkgErrors.New("bad_request", "capacity must be positive")
	}
	if basePrice <= 0 {
		return pkgErrors.New("bad_request", "base price must be positive")
	}
	return nil
}

// RoomStatus represents allowed states.
type RoomStatus string

const (
	RoomAvailable   RoomStatus = "available"
	RoomUnavailable RoomStatus = "unavailable"
	RoomMaintenance RoomStatus = "maintenance"
)

// NormalizeRoomStatus validates or defaults to available.
func NormalizeRoomStatus(raw string) (RoomStatus, error) {
	status := strings.ToLower(strings.TrimSpace(raw))
	if status == "" {
		return RoomAvailable, nil
	}
	switch RoomStatus(status) {
	case RoomAvailable, RoomUnavailable, RoomMaintenance:
		return RoomStatus(status), nil
	default:
		return "", pkgErrors.New("bad_request", "invalid room status")
	}
}
