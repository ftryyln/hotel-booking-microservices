package repository_test

import (
	"context"
	"testing"

	sqlite "github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/hotel"
	repo "github.com/ftryyln/hotel-booking-microservices/internal/infrastructure/hotel/repository"
	"github.com/ftryyln/hotel-booking-microservices/pkg/query"
)

func TestHotelGormRepository(t *testing.T) {
	db := newTestDB(t)
	require.NoError(t, repo.AutoMigrate(db))
	r := repo.NewGormRepository(db)

	h := domain.Hotel{ID: uuid.New(), Name: "H", Address: "Addr"}
	require.NoError(t, r.CreateHotel(context.Background(), h))

	hotels, err := r.ListHotels(context.Background(), query.Options{Limit: 10})
	require.NoError(t, err)
	require.Len(t, hotels, 1)
}

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	return db
}
