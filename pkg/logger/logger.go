package logger

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

var (
	once sync.Once
	log  *zap.Logger
)

// New returns singleton zap logger.
func New() *zap.Logger {
	once.Do(func() {
		cfg := zap.NewProductionConfig()
		cfg.Encoding = "json"
		var err error
		log, err = cfg.Build()
		if err != nil {
			panic(err)
		}
	})
	return log
}

// WithContext attaches request metadata when available.
func WithContext(ctx context.Context) *zap.Logger {
	l := New()
	if ctx == nil {
		return l
	}
	if reqID, ok := ctx.Value("request_id").(string); ok {
		return l.With(zap.String("request_id", reqID))
	}
	return l
}
