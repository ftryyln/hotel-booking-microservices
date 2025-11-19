package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/payment"
)

// PostgresRepository persists payments.
type PostgresRepository struct {
	db *sqlx.DB
}

func NewPostgresRepository(db *sqlx.DB) *PostgresRepository { return &PostgresRepository{db: db} }

func (r *PostgresRepository) Create(ctx context.Context, p domain.Payment) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO payments (id, booking_id, amount, currency, status, provider, payment_url) VALUES ($1,$2,$3,$4,$5,$6,$7)`, p.ID, p.BookingID, p.Amount, p.Currency, p.Status, p.Provider, p.PaymentURL)
	return err
}

func (r *PostgresRepository) FindByID(ctx context.Context, id uuid.UUID) (domain.Payment, error) {
	var p domain.Payment
	err := r.db.GetContext(ctx, &p, `SELECT id, booking_id, amount, currency, status, provider, payment_url FROM payments WHERE id=$1`, id)
	return p, err
}

func (r *PostgresRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status, paymentURL string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE payments SET status=$2, payment_url=COALESCE($3, payment_url) WHERE id=$1`, id, status, paymentURL)
	return err
}
