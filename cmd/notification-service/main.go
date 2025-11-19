package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"

	dispatcher "github.com/ftryyln/hotel-booking-microservices/internal/infrastructure/notification/dispatcher"
	notificationhttp "github.com/ftryyln/hotel-booking-microservices/internal/infrastructure/notification/http"
	notificationuc "github.com/ftryyln/hotel-booking-microservices/internal/usecase/notification"
	"github.com/ftryyln/hotel-booking-microservices/pkg/config"
	"github.com/ftryyln/hotel-booking-microservices/pkg/logger"
	"github.com/ftryyln/hotel-booking-microservices/pkg/server"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()
	log := logger.New()

	dispatch := dispatcher.NewLoggerDispatcher(log)
	service := notificationuc.NewService(dispatch)
	handler := notificationhttp.NewHandler(service)

	r := chi.NewRouter()
	r.Mount("/", handler.Routes())

	srv := server.New(cfg.HTTPPort, r, log)
	srv.Start()

	<-ctx.Done()
	_ = srv.Stop(context.Background())
}
