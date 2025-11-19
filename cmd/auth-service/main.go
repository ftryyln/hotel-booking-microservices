package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	authhttp "github.com/ftryyln/hotel-booking-microservices/internal/infrastructure/auth/http"
	authrepo "github.com/ftryyln/hotel-booking-microservices/internal/infrastructure/auth/repository"
	authtoken "github.com/ftryyln/hotel-booking-microservices/internal/infrastructure/auth/token"
	authuc "github.com/ftryyln/hotel-booking-microservices/internal/usecase/auth"
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

	repo := authrepo.NewPostgresRepository(db)
	issuer := authtoken.NewJWTIssuer(cfg.JWTSecret)
	service := authuc.NewService(repo, issuer)
	handler := authhttp.NewHandler(service)

	r := chi.NewRouter()
	r.Mount("/auth", handler.Routes())

	srv := server.New(cfg.HTTPPort, r, log)
	srv.Start()

	<-ctx.Done()
	_ = srv.Stop(context.Background())
}
