package repository_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	sqlite "github.com/glebarez/sqlite"

	"github.com/ftryyln/hotel-booking-microservices/internal/domain/auth"
	repo "github.com/ftryyln/hotel-booking-microservices/internal/infrastructure/auth/repository"
)

func TestGormRepositoryCreateUser(t *testing.T) {
	db := newTestDB(t)
	require.NoError(t, repo.AutoMigrate(db))
	r := repo.NewGormRepository(db)

	user := auth.User{
		ID:       uuid.New(),
		Email:    "user@example.com",
		Password: "hash",
		Role:     "customer",
	}
	require.NoError(t, r.Create(ctxBackground(), user))

	u, err := r.FindByEmail(ctxBackground(), user.Email)
	require.NoError(t, err)
	require.Equal(t, user.Email, u.Email)
}

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	return db
}

func ctxBackground() context.Context { return context.Background() }
