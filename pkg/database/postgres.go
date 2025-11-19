package database

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// NewPostgres opens sqlx DB with sane pool defaults.
func NewPostgres(ctx context.Context, dsn string) (*sqlx.DB, error) {
	db, err := sqlx.ConnectContext(ctx, "postgres", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	return db, nil
}
