package dispatcher

import (
	"context"

	"go.uber.org/zap"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/notification"
)

// LoggerDispatcher logs notifications.
type LoggerDispatcher struct {
	log *zap.Logger
}

func NewLoggerDispatcher(log *zap.Logger) domain.Dispatcher {
	return &LoggerDispatcher{log: log}
}

func (d *LoggerDispatcher) Dispatch(ctx context.Context, target, message string) error {
	d.log.Info("notification", zap.String("target", target), zap.String("message", message))
	return nil
}
