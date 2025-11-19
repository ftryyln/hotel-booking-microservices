package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	paymentbooking "github.com/ftryyln/hotel-booking-microservices/internal/infrastructure/payment/booking"
	paymenthttp "github.com/ftryyln/hotel-booking-microservices/internal/infrastructure/payment/http"
	paymentprovider "github.com/ftryyln/hotel-booking-microservices/internal/infrastructure/payment/provider"
	paymentrepo "github.com/ftryyln/hotel-booking-microservices/internal/infrastructure/payment/repository"
	paymentuc "github.com/ftryyln/hotel-booking-microservices/internal/usecase/payment"
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

	repo := paymentrepo.NewPostgresRepository(db)
	provider := paymentprovider.NewXenditMockProvider(cfg.PaymentProviderKey)
	statusClient := paymentbooking.NewHTTPStatusClient(cfg.BookingServiceURL)
	service := paymentuc.NewService(repo, provider, statusClient)
	handler := paymenthttp.NewHandler(service)

	r := chi.NewRouter()
	r.Mount("/", handler.Routes())

	srv := server.New(cfg.HTTPPort, r, log)
	srv.Start()

	<-ctx.Done()
	_ = srv.Stop(context.Background())
}
