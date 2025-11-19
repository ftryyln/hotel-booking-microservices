package middleware

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/ftryyln/hotel-booking-microservices/pkg/errors"
)

type contextKey string

// AuthContextKey stores claims inside context.
const AuthContextKey contextKey = "auth_claims"

// Claims extends JWT with extra info.
type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// JWT middleware validates Authorization header.
func JWT(secret string, roles ...string) func(http.Handler) http.Handler {
	allowed := map[string]struct{}{}
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := extractToken(r)
			if tokenString == "" {
				writeError(w, errors.New("unauthorized", "missing token"))
				return
			}

			claims := &Claims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				writeError(w, errors.New("unauthorized", "invalid token"))
				return
			}

			if len(allowed) > 0 {
				if _, ok := allowed[claims.Role]; !ok {
					writeError(w, errors.New("forbidden", "insufficient role"))
					return
				}
			}

			ctx := context.WithValue(r.Context(), AuthContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	parts := strings.Split(auth, " ")
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}

// RateLimiter throttles requests.
func RateLimiter(limitPerMinute int) func(http.Handler) http.Handler {
	if limitPerMinute <= 0 {
		limitPerMinute = 60
	}
	interval := time.Minute / time.Duration(limitPerMinute)
	var mu sync.Mutex
	last := time.Now().Add(-interval)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mu.Lock()
			defer mu.Unlock()
			now := time.Now()
			if since := now.Sub(last); since < interval {
				time.Sleep(interval - since)
			}
			last = time.Now()
			next.ServeHTTP(w, r)
		})
	}
}

func writeError(w http.ResponseWriter, err errors.APIError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errors.StatusCode(err))
	_, _ = w.Write([]byte(`{"code":"` + err.Code + `","message":"` + err.Message + `"}`))
}
