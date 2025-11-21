package assembler

import (
	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/auth"
	"github.com/ftryyln/hotel-booking-microservices/pkg/dto"
)

// ToProfile maps domain user to profile DTO.
func ToProfile(u domain.User) dto.ProfileResponse {
	return dto.ProfileResponse{
		ID:    u.ID.String(),
		Email: u.Email,
		Role:  u.Role,
	}
}
