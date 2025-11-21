package main

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	hotelhttp "github.com/ftryyln/hotel-booking-microservices/internal/infrastructure/hotel/http"
	hotelrepo "github.com/ftryyln/hotel-booking-microservices/internal/infrastructure/hotel/repository"
	hoteluc "github.com/ftryyln/hotel-booking-microservices/internal/usecase/hotel"
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

	db, err := database.NewGormPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("failed to connect to postgres", zap.Error(err))
	}

	if err := hotelrepo.AutoMigrate(db); err != nil {
		log.Fatal("failed to run migrations", zap.Error(err))
	}

	repo := hotelrepo.NewGormRepository(db)
	service := hoteluc.NewService(repo)
	handler := hotelhttp.NewHandler(service, cfg.JWTSecret)

	r := chi.NewRouter()
	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	r.Mount("/", handler.Routes())

	srv := server.New(cfg.HTTPPort, r, log)
	srv.Start()

	<-ctx.Done()
	_ = srv.Stop(context.Background())
}
