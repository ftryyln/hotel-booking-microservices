package server

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// HTTPServer wraps http.Server with helpers.
type HTTPServer struct {
	server *http.Server
	log    *zap.Logger
}

// New constructs server.
func New(addr string, handler http.Handler, log *zap.Logger) *HTTPServer {
	return &HTTPServer{
		server: &http.Server{
			Addr:         addr,
			Handler:      handler,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		log: log,
	}
}

// Start begins serving in goroutine.
func (s *HTTPServer) Start() {
	go func() {
		s.log.Info("http server started", zap.String("addr", s.server.Addr))
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.log.Fatal("http server crashed", zap.Error(err))
		}
	}()
}

// Stop gracefully shuts down server.
func (s *HTTPServer) Stop(ctx context.Context) error {
	s.log.Info("shutting down http server")
	return s.server.Shutdown(ctx)
}
