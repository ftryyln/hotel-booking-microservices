package repository_test

import (
	"context"
	"testing"

	sqlite "github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/ftryyln/hotel-booking-microservices/internal/domain/payment"
	repo "github.com/ftryyln/hotel-booking-microservices/internal/infrastructure/payment/repository"
)

func TestPaymentGormRepository(t *testing.T) {
	db := newTestDB(t)
	require.NoError(t, repo.AutoMigrate(db))
	r := repo.NewGormRepository(db)

	p := payment.Payment{
		ID:        uuid.New(),
		BookingID: uuid.New(),
		Amount:    100,
		Currency:  "IDR",
		Status:    "pending",
		Provider:  "mock",
	}
	require.NoError(t, r.Create(context.Background(), p))

	found, err := r.FindByID(context.Background(), p.ID)
	require.NoError(t, err)
	require.Equal(t, p.ID, found.ID)

	require.NoError(t, r.UpdateStatus(context.Background(), p.ID, "paid", "url"))
	updated, err := r.FindByID(context.Background(), p.ID)
	require.NoError(t, err)
	require.Equal(t, "paid", updated.Status)
}

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	return db
}
