package booking

import (
	"context"

	"github.com/google/uuid"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/booking"
	hdomain "github.com/ftryyln/hotel-booking-microservices/internal/domain/hotel"
	"github.com/ftryyln/hotel-booking-microservices/pkg/dto"
	"github.com/ftryyln/hotel-booking-microservices/pkg/errors"
	"github.com/ftryyln/hotel-booking-microservices/pkg/utils"
)

// Service handles booking lifecycle.
type Service struct {
	repo     domain.Repository
	hotels   hdomain.Repository
	payments domain.PaymentGateway
	notifier domain.NotificationGateway
}

func NewService(repo domain.Repository, hotels hdomain.Repository, payments domain.PaymentGateway, notifier domain.NotificationGateway) *Service {
	return &Service{repo: repo, hotels: hotels, payments: payments, notifier: notifier}
}

func (s *Service) CreateBooking(ctx context.Context, req dto.BookingRequest) (dto.BookingResponse, error) {
	nights, err := utils.NightsBetween(req.CheckIn, req.CheckOut)
	if err != nil {
		return dto.BookingResponse{}, errors.New("bad_request", "invalid dates")
	}

	roomTypeID, err := uuid.Parse(req.RoomTypeID)
	if err != nil {
		return dto.BookingResponse{}, errors.New("bad_request", "invalid room type id")
	}

	rt, err := s.hotels.GetRoomType(ctx, roomTypeID)
	if err != nil {
		return dto.BookingResponse{}, errors.New("not_found", "room type not found")
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return dto.BookingResponse{}, errors.New("bad_request", "invalid user id")
	}

	booking := domain.Booking{
		ID:          uuid.New(),
		UserID:      userID,
		RoomTypeID:  rt.ID,
		CheckIn:     req.CheckIn,
		CheckOut:    req.CheckOut,
		Status:      domain.StatusPendingPayment,
		TotalPrice:  float64(nights) * rt.BasePrice,
		TotalNights: nights,
	}

	if err := s.repo.Create(ctx, booking); err != nil {
		return dto.BookingResponse{}, err
	}

	if _, err := s.payments.Initiate(ctx, booking.ID, booking.TotalPrice); err != nil {
		return dto.BookingResponse{}, err
	}

	_ = s.notifier.Notify(ctx, "booking_created", booking.ID.String())

	return toDTO(booking), nil
}

func (s *Service) CancelBooking(ctx context.Context, id uuid.UUID) error {
	booking, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if booking.Status != domain.StatusPendingPayment {
		return errors.New("bad_request", "cannot cancel at this stage")
	}
	return s.repo.UpdateStatus(ctx, id, domain.StatusCancelled)
}

func (s *Service) ApplyStatus(ctx context.Context, id uuid.UUID, status string) error {
	return s.repo.UpdateStatus(ctx, id, status)
}

func (s *Service) Checkpoint(ctx context.Context, id uuid.UUID, action string) error {
	var status string
	switch action {
	case "check_in":
		status = domain.StatusCheckedIn
	case "complete":
		status = domain.StatusCompleted
	default:
		return errors.New("bad_request", "unknown checkpoint action")
	}
	return s.repo.UpdateStatus(ctx, id, status)
}

func (s *Service) GetBooking(ctx context.Context, id uuid.UUID) (dto.BookingResponse, error) {
	b, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return dto.BookingResponse{}, err
	}
	return toDTO(b), nil
}

func toDTO(b domain.Booking) dto.BookingResponse {
	return dto.BookingResponse{
		ID:          b.ID.String(),
		Status:      b.Status,
		TotalNights: b.TotalNights,
		TotalPrice:  b.TotalPrice,
		CheckIn:     b.CheckIn,
		CheckOut:    b.CheckOut,
	}
}
