package auth

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/ftryyln/hotel-booking-microservices/pkg/query"
)

// User represents auth entity.
type User struct {
	ID        uuid.UUID
	Email     string
	Password  string
	Role      string
	CreatedAt time.Time
}

// UserRepository persists users.
type UserRepository interface {
	Create(ctx context.Context, user User) error
	FindByEmail(ctx context.Context, email string) (User, error)
	FindByID(ctx context.Context, id uuid.UUID) (User, error)
	List(ctx context.Context, opts query.Options) ([]User, error)
}

// TokenIssuer issues JWT tokens.
type TokenIssuer interface {
	Generate(ctx context.Context, user User) (access, refresh string, err error)
}
