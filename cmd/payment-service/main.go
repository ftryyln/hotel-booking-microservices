package main

import (
	"context"
	"net/http"
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
	"github.com/ftryyln/hotel-booking-microservices/pkg/middleware"
	"github.com/ftryyln/hotel-booking-microservices/pkg/server"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()
	log := logger.New()

	db, err := database.NewGormPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("failed to connect to postgres", zap.Error(err))
	}

	if err := paymentrepo.AutoMigrate(db); err != nil {
		log.Fatal("failed to run migrations", zap.Error(err))
	}

	repo := paymentrepo.NewGormRepository(db)
	provider := paymentprovider.NewXenditMockProvider(cfg.PaymentProviderKey)
	statusClient := paymentbooking.NewHTTPStatusClient(cfg.BookingServiceURL)
	service := paymentuc.NewService(repo, provider, statusClient)
	handler := paymenthttp.NewHandler(service)

	api := chi.NewRouter()
	api.Use(middleware.JWT(cfg.JWTSecret))
	api.Post("/payments", handler.CreatePayment)
	api.Get("/payments/{id}", handler.GetPayment)
	api.Get("/payments/by-booking/{booking_id}", handler.GetByBooking)
	api.Post("/payments/refund", handler.Refund)

	r := chi.NewRouter()
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	// Authenticated payment routes; webhook remains public for provider callbacks.
	r.Mount("/", api)
	r.Post("/payments/webhook", handler.HandleWebhook)

	srv := server.New(cfg.HTTPPort, r, log)
	srv.Start()

	<-ctx.Done()
	_ = srv.Stop(context.Background())
}
