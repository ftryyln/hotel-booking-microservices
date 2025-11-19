package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/auth"
)

// PostgresRepository implements UserRepository via sqlx.
type PostgresRepository struct {
	db *sqlx.DB
}

func NewPostgresRepository(db *sqlx.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, user domain.User) error {
	query := `INSERT INTO users (id, email, password, role) VALUES ($1,$2,$3,$4)`
	_, err := r.db.ExecContext(ctx, query, user.ID, user.Email, user.Password, user.Role)
	return err
}

func (r *PostgresRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	var u domain.User
	err := r.db.GetContext(ctx, &u, `SELECT id, email, password, role, created_at FROM users WHERE email=$1`, email)
	return u, err
}

func (r *PostgresRepository) FindByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	var u domain.User
	err := r.db.GetContext(ctx, &u, `SELECT id, email, password, role, created_at FROM users WHERE id=$1`, id)
	return u, err
}
