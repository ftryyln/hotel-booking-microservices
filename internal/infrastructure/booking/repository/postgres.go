package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/booking"
)

// PostgresRepository persists bookings.
type PostgresRepository struct {
	db *sqlx.DB
}

func NewPostgresRepository(db *sqlx.DB) *PostgresRepository { return &PostgresRepository{db: db} }

func (r *PostgresRepository) Create(ctx context.Context, b domain.Booking) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO bookings (id, user_id, room_type_id, check_in, check_out, status, total_price, total_nights) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, b.ID, b.UserID, b.RoomTypeID, b.CheckIn, b.CheckOut, b.Status, b.TotalPrice, b.TotalNights)
	return err
}

func (r *PostgresRepository) FindByID(ctx context.Context, id uuid.UUID) (domain.Booking, error) {
	var b domain.Booking
	err := r.db.GetContext(ctx, &b, `SELECT id, user_id, room_type_id, check_in, check_out, status, total_price, total_nights FROM bookings WHERE id=$1`, id)
	return b, err
}

func (r *PostgresRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE bookings SET status=$2 WHERE id=$1`, id, status)
	return err
}
