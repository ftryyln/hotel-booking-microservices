package config

import (
	"log"
	"os"
	"strconv"
)

// Config stores environment driven configuration.
type Config struct {
	ServiceName        string
	HTTPPort           string
	DatabaseURL        string
	JWTSecret          string
	PaymentProviderKey string
	PaymentServiceURL  string
	BookingServiceURL  string
	NotificationURL    string
	AggregateTargetURL string
	RateLimitPerMinute int
}

// Load reads env vars with defaults.
func Load() Config {
	limit := 60
	if v := os.Getenv("RATE_LIMIT_PER_MINUTE"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			limit = parsed
		}
	}

	cfg := Config{
		ServiceName:        getEnv("SERVICE_NAME", "hotel-service"),
		HTTPPort:           getEnv("HTTP_PORT", ":8080"),
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://postgres:postgres@postgres:5432/hotel?sslmode=disable"),
		JWTSecret:          getEnv("JWT_SECRET", "super-secret"),
		PaymentProviderKey: getEnv("PAYMENT_PROVIDER_KEY", "sandbox-key"),
		PaymentServiceURL:  getEnv("PAYMENT_SERVICE_URL", "http://payment-service:8083"),
		BookingServiceURL:  getEnv("BOOKING_SERVICE_URL", "http://booking-service:8082"),
		NotificationURL:    getEnv("NOTIFICATION_SERVICE_URL", "http://notification-service:8085"),
		AggregateTargetURL: getEnv("AGGREGATE_TARGET_URL", "http://hotel-service:8081"),
		RateLimitPerMinute: limit,
	}

	if cfg.ServiceName == "" {
		log.Println("SERVICE_NAME not provided; using hotel-service")
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
