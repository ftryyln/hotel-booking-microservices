package repository_test

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/booking"
	repo "github.com/ftryyln/hotel-booking-microservices/internal/infrastructure/booking/repository"
)

func TestPostgresRepositoryCreate(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repository := repo.NewPostgresRepository(sqlxDB)

	booking := domain.Booking{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		RoomTypeID:  uuid.New(),
		CheckIn:     time.Now(),
		CheckOut:    time.Now().Add(48 * time.Hour),
		Status:      domain.StatusPendingPayment,
		TotalPrice:  1000,
		TotalNights: 2,
	}

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO bookings")).
		WithArgs(booking.ID, booking.UserID, booking.RoomTypeID, booking.CheckIn, booking.CheckOut, booking.Status, booking.TotalPrice, booking.TotalNights).
		WillReturnResult(sqlmock.NewResult(1, 1))

	require.NoError(t, repository.Create(context.Background(), booking))
	require.NoError(t, mock.ExpectationsWereMet())
}
