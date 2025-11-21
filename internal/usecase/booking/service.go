package booking

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/booking"
	hdomain "github.com/ftryyln/hotel-booking-microservices/internal/domain/hotel"
	"github.com/ftryyln/hotel-booking-microservices/internal/usecase/booking/assembler"
	"github.com/ftryyln/hotel-booking-microservices/pkg/dto"
	"github.com/ftryyln/hotel-booking-microservices/pkg/errors"
	"github.com/ftryyln/hotel-booking-microservices/pkg/query"
	"github.com/ftryyln/hotel-booking-microservices/pkg/valueobject"
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
	dateRange, err := valueobject.NewDateRange(req.CheckIn, req.CheckOut)
	if err != nil {
		return dto.BookingResponse{}, err
	}
	nights := dateRange.Nights()
	rtID, err := uuid.Parse(req.RoomTypeID)
	if err != nil {
		return dto.BookingResponse{}, errors.New("bad_request", "invalid room type id")
	}
	rt, err := s.hotels.GetRoomType(ctx, rtID)
	if err != nil {
		return dto.BookingResponse{}, errors.New("not_found", "room type not found")
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return dto.BookingResponse{}, errors.New("bad_request", "invalid user id")
	}

	guests := req.Guests
	if guests <= 0 {
		guests = 1
	}

	booking := domain.Booking{
		ID:          uuid.New(),
		UserID:      userID,
		RoomTypeID:  rtID,
		CheckIn:     req.CheckIn,
		CheckOut:    req.CheckOut,
		Status:      string(valueobject.StatusPendingPayment),
		Guests:      guests,
		TotalPrice:  float64(nights) * rt.BasePrice,
		TotalNights: nights,
	}

	if err := s.repo.Create(ctx, booking); err != nil {
		return dto.BookingResponse{}, err
	}

	paymentResult, err := s.payments.Initiate(ctx, booking.ID, booking.TotalPrice)
	if err != nil {
		return dto.BookingResponse{}, err
	}

	_ = s.notifier.Notify(ctx, "booking_created", booking.ID.String())

	return assembler.ToResponse(booking, paymentResult), nil
}

func (s *Service) CancelBooking(ctx context.Context, id uuid.UUID) error {
	booking, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if booking.Status != string(valueobject.StatusPendingPayment) {
		return errors.New("bad_request", "cannot cancel at this stage")
	}
	return s.repo.UpdateStatus(ctx, id, string(valueobject.StatusCancelled))
}

func (s *Service) ApplyStatus(ctx context.Context, id uuid.UUID, status string) error {
	bk, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("not_found", "booking not found")
		}
		return err
	}
	current, err := valueobject.ValidateBookingStatus(bk.Status)
	if err != nil {
		return err
	}
	target, err := valueobject.ValidateBookingStatus(status)
	if err != nil {
		return err
	}
	if err := current.CanTransition(target); err != nil {
		return err
	}
	return s.repo.UpdateStatus(ctx, id, string(target))
}

func (s *Service) Checkpoint(ctx context.Context, id uuid.UUID, action string) error {
	bk, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("not_found", "booking not found")
		}
		return err
	}
	if bk.Status == domain.StatusCancelled || bk.Status == domain.StatusCompleted {
		return errors.New("bad_request", "booking cannot change state from current status")
	}

	var status string
	switch action {
	case "check_in":
		if bk.Status != domain.StatusConfirmed {
			return errors.New("bad_request", "booking must be confirmed before check-in")
		}
		status = string(valueobject.StatusCheckedIn)
	case "complete":
		if bk.Status != string(valueobject.StatusCheckedIn) {
			return errors.New("bad_request", "booking must be checked-in before completion")
		}
		status = string(valueobject.StatusCompleted)
	default:
		return errors.New("bad_request", "unknown checkpoint action")
	}
	return s.repo.UpdateStatus(ctx, id, status)
}

func (s *Service) GetBooking(ctx context.Context, id uuid.UUID) (dto.BookingResponse, error) {
	b, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return dto.BookingResponse{}, errors.New("not_found", "booking not found")
		}
		return dto.BookingResponse{}, err
	}
	return assembler.ToResponse(b, domain.PaymentResult{}), nil
}

func (s *Service) ListBookings(ctx context.Context, opts query.Options) ([]dto.BookingResponse, error) {
	bks, err := s.repo.List(ctx, opts.Normalize(50))
	if err != nil {
		return nil, err
	}
	resp := make([]dto.BookingResponse, 0, len(bks))
	for _, b := range bks {
		resp = append(resp, assembler.ToResponse(b, domain.PaymentResult{}))
	}
	return resp, nil
}
