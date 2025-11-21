package hotel

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/ftryyln/hotel-booking-microservices/pkg/query"
)

// Hotel entity.
type Hotel struct {
	ID          uuid.UUID
	Name        string
	Description string
	Address     string
	CreatedAt   time.Time
}

// RoomType entity.
type RoomType struct {
	ID        uuid.UUID
	HotelID   uuid.UUID
	Name      string
	Capacity  int
	BasePrice float64
	Amenities string
}

// Room entity.
type Room struct {
	ID         uuid.UUID
	RoomTypeID uuid.UUID
	Number     string
	Status     string
}

// Repository contract.
type Repository interface {
	CreateHotel(ctx context.Context, h Hotel) error
	ListHotels(ctx context.Context, opts query.Options) ([]Hotel, error)
	CreateRoomType(ctx context.Context, rt RoomType) error
	ListRoomTypes(ctx context.Context, hotelID uuid.UUID) ([]RoomType, error)
	ListAllRoomTypes(ctx context.Context, opts query.Options) ([]RoomType, error)
	CreateRoom(ctx context.Context, room Room) error
	GetRoomType(ctx context.Context, id uuid.UUID) (RoomType, error)
	ListRooms(ctx context.Context, opts query.Options) ([]Room, error)
	GetHotel(ctx context.Context, id uuid.UUID) (Hotel, error)
}
