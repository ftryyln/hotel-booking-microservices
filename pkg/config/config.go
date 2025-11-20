package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
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
	AuthServiceURL     string
	AggregateTargetURL string
	RateLimitPerMinute int
	GatewayMode        string
	RoutesFile         string
	HealthInterval     time.Duration
	UpstreamTimeout    time.Duration
	UpstreamRetries    int
	CircuitWindow      time.Duration
	CircuitThreshold   float64
	CircuitCooldown    time.Duration
}

// Load reads env vars with defaults.
func Load() Config {
	limit := intEnv("RATE_LIMIT_PER_MINUTE", 60)

	cfg := Config{
		ServiceName:        getEnv("SERVICE_NAME", "hotel-service"),
		HTTPPort:           getEnv("HTTP_PORT", ":8080"),
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://postgres:postgres@postgres:5432/hotel?sslmode=disable"),
		JWTSecret:          getEnv("JWT_SECRET", "super-secret"),
		PaymentProviderKey: getEnv("PAYMENT_PROVIDER_KEY", "sandbox-key"),
		PaymentServiceURL:  getEnv("PAYMENT_SERVICE_URL", "http://payment-service:8083"),
		BookingServiceURL:  getEnv("BOOKING_SERVICE_URL", "http://booking-service:8082"),
		NotificationURL:    getEnv("NOTIFICATION_SERVICE_URL", "http://notification-service:8085"),
		AuthServiceURL:     getEnv("AUTH_SERVICE_URL", "http://auth-service:8080"),
		AggregateTargetURL: getEnv("AGGREGATE_TARGET_URL", "http://hotel-service:8081"),
		RateLimitPerMinute: limit,
		GatewayMode:        strings.ToLower(getEnv("GATEWAY_MODE", "whitelist")),
		RoutesFile:         getEnv("GATEWAY_ROUTES_FILE", "config/routes.yml"),
		HealthInterval:     durationEnv("HEALTH_INTERVAL", 10*time.Second),
		UpstreamTimeout:    durationEnv("UPSTREAM_TIMEOUT", 5*time.Second),
		UpstreamRetries:    intEnv("UPSTREAM_RETRIES", 2),
		CircuitWindow:      durationEnv("CIRCUIT_BREAKER_WINDOW", 30*time.Second),
		CircuitThreshold:   floatEnv("CIRCUIT_BREAKER_THRESHOLD", 0.5),
		CircuitCooldown:    durationEnv("CIRCUIT_BREAKER_COOLDOWN", 15*time.Second),
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

func intEnv(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			return parsed
		}
	}
	return fallback
}

func durationEnv(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if parsed, err := time.ParseDuration(v); err == nil {
			return parsed
		}
	}
	return fallback
}

func floatEnv(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		if parsed, err := strconv.ParseFloat(v, 64); err == nil {
			return parsed
		}
	}
	return fallback
}
