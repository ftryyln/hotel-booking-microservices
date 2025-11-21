package assembler

import (
	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/payment"
	"github.com/ftryyln/hotel-booking-microservices/pkg/dto"
)

// ToResponse maps domain Payment to DTO.
func ToResponse(p domain.Payment) dto.PaymentResponse {
	return dto.PaymentResponse{
		ID:         p.ID.String(),
		Status:     p.Status,
		Provider:   p.Provider,
		PaymentURL: p.PaymentURL,
	}
}

// ToRefundResponse maps refund.
func ToRefundResponse(id string, status, ref string) dto.RefundResponse {
	return dto.RefundResponse{
		ID:        id,
		Status:    status,
		Reference: ref,
	}
}
