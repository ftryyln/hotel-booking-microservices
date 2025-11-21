package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	sqlite "github.com/glebarez/sqlite"
	"gorm.io/gorm"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/booking"
	repo "github.com/ftryyln/hotel-booking-microservices/internal/infrastructure/booking/repository"
)

func TestGormRepositoryCreate(t *testing.T) {
	db := newTestDB(t)
	require.NoError(t, repo.AutoMigrate(db))
	r := repo.NewGormRepository(db)

	booking := domain.Booking{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		RoomTypeID:  uuid.New(),
		CheckIn:     time.Now(),
		CheckOut:    time.Now().Add(48 * time.Hour),
		Status:      domain.StatusPendingPayment,
		TotalPrice:  1000,
		TotalNights: 2,
		Guests:      1,
	}

	require.NoError(t, r.Create(context.Background(), booking))

	var count int64
	require.NoError(t, db.Model(&repoTestBookingModel{}).Count(&count).Error)
	require.Equal(t, int64(1), count)
}

// repoTestBookingModel mirrors bookingModel table name for counting.
type repoTestBookingModel struct {
	ID uuid.UUID `gorm:"type:uuid;primaryKey"`
}

func (repoTestBookingModel) TableName() string { return "bookings" }

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	return db
}
