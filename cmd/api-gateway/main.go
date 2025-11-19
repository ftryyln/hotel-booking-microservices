package main

// @title Hotel Booking Microservices API
// @version 1.0
// @description Aggregated API surface for hotel booking platform.
// @BasePath /
// @schemes http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"

	"github.com/ftryyln/hotel-booking-microservices/internal/infrastructure/gateway"
	"github.com/ftryyln/hotel-booking-microservices/pkg/config"
	"github.com/ftryyln/hotel-booking-microservices/pkg/logger"
	"github.com/ftryyln/hotel-booking-microservices/pkg/middleware"
	"github.com/ftryyln/hotel-booking-microservices/pkg/server"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()
	log := logger.New()

	handler := gateway.NewHandler(cfg.BookingServiceURL, cfg.PaymentServiceURL, cfg.AggregateTargetURL, cfg.RateLimitPerMinute)

	r := chi.NewRouter()
	r.Use(middleware.JWT(cfg.JWTSecret))
	r.Mount("/gateway", handler.Routes())

	srv := server.New(cfg.HTTPPort, r, log)
	srv.Start()

	<-ctx.Done()
	_ = srv.Stop(context.Background())
}
