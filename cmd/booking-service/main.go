package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	bookinghttp "github.com/ftryyln/hotel-booking-microservices/internal/infrastructure/booking/http"
	bookingnotification "github.com/ftryyln/hotel-booking-microservices/internal/infrastructure/booking/notification"
	bookingpayment "github.com/ftryyln/hotel-booking-microservices/internal/infrastructure/booking/payment"
	bookingrepo "github.com/ftryyln/hotel-booking-microservices/internal/infrastructure/booking/repository"
	hotelrepo "github.com/ftryyln/hotel-booking-microservices/internal/infrastructure/hotel/repository"
	bookinguc "github.com/ftryyln/hotel-booking-microservices/internal/usecase/booking"
	"github.com/ftryyln/hotel-booking-microservices/pkg/config"
	"github.com/ftryyln/hotel-booking-microservices/pkg/database"
	"github.com/ftryyln/hotel-booking-microservices/pkg/logger"
	"github.com/ftryyln/hotel-booking-microservices/pkg/server"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()
	log := logger.New()

	db, err := database.NewPostgres(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal("failed to connect to postgres", zap.Error(err))
	}

	repo := bookingrepo.NewPostgresRepository(db)
	hRepo := hotelrepo.NewPostgresRepository(db)
	paymentClient := bookingpayment.NewHTTPGateway(cfg.PaymentServiceURL)
	notifier := bookingnotification.NewHTTPGateway(cfg.NotificationURL)
	service := bookinguc.NewService(repo, hRepo, paymentClient, notifier)
	handler := bookinghttp.NewHandler(service)

	r := chi.NewRouter()
	r.Mount("/", handler.Routes())

	srv := server.New(cfg.HTTPPort, r, log)
	srv.Start()

	<-ctx.Done()
	_ = srv.Stop(context.Background())
}
