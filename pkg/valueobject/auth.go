package valueobject

import (
	"strings"

	pkgErrors "github.com/ftryyln/hotel-booking-microservices/pkg/errors"
)

// NormalizeEmail trims and lowercases basic email form.
func NormalizeEmail(raw string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(raw))
	if email == "" || !strings.Contains(email, "@") {
		return "", pkgErrors.New("bad_request", "invalid email")
	}
	return email, nil
}

// Role represents allowed user roles.
type Role string

const (
	RoleCustomer Role = "customer"
	RoleAdmin    Role = "admin"
)

// ParseRole validates and returns a normalized role.
func ParseRole(raw string) (Role, error) {
	role := strings.ToLower(strings.TrimSpace(raw))
	if role == "" {
		role = string(RoleCustomer)
	}
	switch Role(role) {
	case RoleCustomer, RoleAdmin:
		return Role(role), nil
	default:
		return "", pkgErrors.New("bad_request", "invalid role")
	}
}
