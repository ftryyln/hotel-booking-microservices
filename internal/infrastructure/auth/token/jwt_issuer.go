package token

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/auth"
)

// JWTIssuer issues HS256 tokens.
type JWTIssuer struct {
	secret []byte
}

func NewJWTIssuer(secret string) *JWTIssuer {
	return &JWTIssuer{secret: []byte(secret)}
}

func (i *JWTIssuer) Generate(ctx context.Context, user domain.User) (string, string, error) {
	accessClaims := jwt.MapClaims{
		"sub":   user.ID.String(),
		"role":  user.Role,
		"email": user.Email,
		"exp":   time.Now().Add(30 * time.Minute).Unix(),
	}
	refreshClaims := jwt.MapClaims{
		"sub": user.ID.String(),
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	}

	access, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(i.secret)
	if err != nil {
		return "", "", err
	}
	refresh, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(i.secret)
	return access, refresh, err
}
