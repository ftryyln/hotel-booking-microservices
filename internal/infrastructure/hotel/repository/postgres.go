package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/hotel"
)

// PostgresRepository implements hotel repo.
type PostgresRepository struct {
	db *sqlx.DB
}

func NewPostgresRepository(db *sqlx.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateHotel(ctx context.Context, h domain.Hotel) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO hotels (id, name, description, address) VALUES ($1,$2,$3,$4)`, h.ID, h.Name, h.Description, h.Address)
	return err
}

func (r *PostgresRepository) ListHotels(ctx context.Context) ([]domain.Hotel, error) {
	var hotels []domain.Hotel
	err := r.db.SelectContext(ctx, &hotels, `SELECT id, name, description, address, created_at FROM hotels`)
	return hotels, err
}

func (r *PostgresRepository) CreateRoomType(ctx context.Context, rt domain.RoomType) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO room_types (id, hotel_id, name, capacity, base_price, amenities) VALUES ($1,$2,$3,$4,$5,$6)`, rt.ID, rt.HotelID, rt.Name, rt.Capacity, rt.BasePrice, rt.Amenities)
	return err
}

func (r *PostgresRepository) ListRoomTypes(ctx context.Context, hotelID uuid.UUID) ([]domain.RoomType, error) {
	var rts []domain.RoomType
	err := r.db.SelectContext(ctx, &rts, `SELECT id, hotel_id, name, capacity, base_price, amenities FROM room_types WHERE hotel_id=$1`, hotelID)
	return rts, err
}

func (r *PostgresRepository) CreateRoom(ctx context.Context, room domain.Room) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO rooms (id, room_type_id, number, status) VALUES ($1,$2,$3,$4)`, room.ID, room.RoomTypeID, room.Number, room.Status)
	return err
}

func (r *PostgresRepository) GetRoomType(ctx context.Context, id uuid.UUID) (domain.RoomType, error) {
	var rt domain.RoomType
	err := r.db.GetContext(ctx, &rt, `SELECT id, hotel_id, name, capacity, base_price, amenities FROM room_types WHERE id=$1`, id)
	return rt, err
}
