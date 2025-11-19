package notification

import (
	"context"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/notification"
	"github.com/ftryyln/hotel-booking-microservices/pkg/dto"
)

// Service wraps dispatcher implementation.
type Service struct {
	dispatcher domain.Dispatcher
}

func NewService(dispatcher domain.Dispatcher) *Service {
	return &Service{dispatcher: dispatcher}
}

func (s *Service) Send(ctx context.Context, req dto.NotificationRequest) error {
	return s.dispatcher.Dispatch(ctx, req.Target, req.Message)
}
