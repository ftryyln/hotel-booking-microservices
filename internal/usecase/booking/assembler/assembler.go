package assembler

import (
	"github.com/google/uuid"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/booking"
	"github.com/ftryyln/hotel-booking-microservices/pkg/dto"
)

// ToResponse maps domain booking plus optional payment info to response DTO.
func ToResponse(b domain.Booking, payment domain.PaymentResult) dto.BookingResponse {
	resp := dto.BookingResponse{
		ID:          b.ID.String(),
		Status:      b.Status,
		Guests:      b.Guests,
		TotalNights: b.TotalNights,
		TotalPrice:  b.TotalPrice,
		CheckIn:     b.CheckIn,
		CheckOut:    b.CheckOut,
	}
	if payment.ID != uuid.Nil {
		resp.Payment = &dto.PaymentResponse{
			ID:         payment.ID.String(),
			Status:     payment.Status,
			Provider:   payment.Provider,
			PaymentURL: payment.PaymentURL,
		}
	}
	return resp
}
