package hotel_test

import (
	"context"
	stdErrors "errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/hotel"
	"github.com/ftryyln/hotel-booking-microservices/internal/usecase/hotel"
	"github.com/ftryyln/hotel-booking-microservices/pkg/dto"
	"github.com/ftryyln/hotel-booking-microservices/pkg/query"
)

func TestCreateHotelValidates(t *testing.T) {
	repo := &hotelRepoStub{}
	svc := hotel.NewService(repo)

	_, err := svc.CreateHotel(context.Background(), dto.HotelRequest{Name: "", Address: ""})
	require.Error(t, err)

	id, err := svc.CreateHotel(context.Background(), dto.HotelRequest{Name: "Hilton", Address: "Jakarta"})
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, id)
}

func TestCreateRoomTypeAndRoom(t *testing.T) {
	repo := &hotelRepoStub{}
	svc := hotel.NewService(repo)
	hID, _ := uuid.Parse("11111111-1111-1111-1111-111111111111")

	rtID, err := svc.CreateRoomType(context.Background(), dto.RoomTypeRequest{
		HotelID:   hID.String(),
		Name:      "Deluxe",
		Capacity:  2,
		BasePrice: 1000,
	})
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, rtID)

	_, err = svc.CreateRoom(context.Background(), dto.RoomRequest{
		RoomTypeID: rtID.String(),
		Number:     "101",
		Status:     "available",
	})
	require.NoError(t, err)

	_, err = svc.CreateRoom(context.Background(), dto.RoomRequest{
		RoomTypeID: rtID.String(),
		Number:     "102",
		Status:     "invalid",
	})
	require.Error(t, err)
}

func TestListHotelsRooms(t *testing.T) {
	repo := &hotelRepoStub{}
	hID := uuid.New()
	now := time.Now()
	repo.hotels = append(repo.hotels, domain.Hotel{ID: hID, Name: "H", Address: "Addr", CreatedAt: now})
	repo.roomTypes = append(repo.roomTypes, domain.RoomType{ID: uuid.New(), HotelID: hID, Name: "RT", Capacity: 2, BasePrice: 10})
	repo.rooms = append(repo.rooms, domain.Room{ID: uuid.New(), RoomTypeID: repo.roomTypes[0].ID, Number: "1", Status: "available"})
	svc := hotel.NewService(repo)

	hotels, err := svc.ListHotels(context.Background(), query.Options{Limit: 10})
	require.NoError(t, err)
	require.Len(t, hotels, 1)

	rt, err := svc.ListRoomTypes(context.Background(), query.Options{Limit: 10})
	require.NoError(t, err)
	require.Len(t, rt, 1)

	rooms, err := svc.ListRooms(context.Background(), query.Options{Limit: 10})
	require.NoError(t, err)
	require.Len(t, rooms, 1)
}

// stub repo
type hotelRepoStub struct {
	hotels    []domain.Hotel
	roomTypes []domain.RoomType
	rooms     []domain.Room
}

func (h *hotelRepoStub) CreateHotel(ctx context.Context, v domain.Hotel) error {
	h.hotels = append(h.hotels, v)
	return nil
}
func (h *hotelRepoStub) ListHotels(ctx context.Context, opts query.Options) ([]domain.Hotel, error) {
	return h.hotels, nil
}
func (h *hotelRepoStub) CreateRoomType(ctx context.Context, rt domain.RoomType) error {
	h.roomTypes = append(h.roomTypes, rt)
	return nil
}
func (h *hotelRepoStub) ListRoomTypes(ctx context.Context, hotelID uuid.UUID) ([]domain.RoomType, error) {
	var out []domain.RoomType
	for _, rt := range h.roomTypes {
		if rt.HotelID == hotelID {
			out = append(out, rt)
		}
	}
	return out, nil
}
func (h *hotelRepoStub) ListAllRoomTypes(ctx context.Context, opts query.Options) ([]domain.RoomType, error) {
	return h.roomTypes, nil
}
func (h *hotelRepoStub) CreateRoom(ctx context.Context, r domain.Room) error {
	h.rooms = append(h.rooms, r)
	return nil
}
func (h *hotelRepoStub) GetRoomType(ctx context.Context, id uuid.UUID) (domain.RoomType, error) {
	for _, rt := range h.roomTypes {
		if rt.ID == id {
			return rt, nil
		}
	}
	return domain.RoomType{}, stdErrors.New("not found")
}
func (h *hotelRepoStub) ListRooms(ctx context.Context, opts query.Options) ([]domain.Room, error) {
	return h.rooms, nil
}
func (h *hotelRepoStub) GetHotel(ctx context.Context, id uuid.UUID) (domain.Hotel, error) {
	for _, ht := range h.hotels {
		if ht.ID == id {
			return ht, nil
		}
	}
	return domain.Hotel{}, stdErrors.New("not found")
}
