package hotelhttp_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/hotel"
	hotelhttp "github.com/ftryyln/hotel-booking-microservices/internal/infrastructure/hotel/http"
	"github.com/ftryyln/hotel-booking-microservices/internal/usecase/hotel"
	"github.com/ftryyln/hotel-booking-microservices/pkg/query"
)

func TestHotelHandlerListHotelsPagination(t *testing.T) {
	repo := &hotelRepoStub{}
	svc := hotel.NewService(repo)
	h := hotelhttp.NewHandler(svc, "secret")
	r := chi.NewRouter()
	r.Mount("/", h.Routes())

	req := httptest.NewRequest(http.MethodGet, "/hotels?limit=1&offset=2", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

// minimal repo stub for handler test
type hotelRepoStub struct{}

func (h *hotelRepoStub) CreateHotel(ctx context.Context, hotel domain.Hotel) error { return nil }
func (h *hotelRepoStub) ListHotels(ctx context.Context, opts query.Options) ([]domain.Hotel, error) {
	return []domain.Hotel{{ID: uuid.New(), Name: "H", Address: "Addr"}}, nil
}
func (h *hotelRepoStub) CreateRoomType(context.Context, domain.RoomType) error { return nil }
func (h *hotelRepoStub) ListRoomTypes(context.Context, uuid.UUID) ([]domain.RoomType, error) {
	return []domain.RoomType{}, nil
}
func (h *hotelRepoStub) ListAllRoomTypes(context.Context, query.Options) ([]domain.RoomType, error) {
	return []domain.RoomType{}, nil
}
func (h *hotelRepoStub) CreateRoom(context.Context, domain.Room) error { return nil }
func (h *hotelRepoStub) GetRoomType(context.Context, uuid.UUID) (domain.RoomType, error) {
	return domain.RoomType{}, nil
}
func (h *hotelRepoStub) ListRooms(context.Context, query.Options) ([]domain.Room, error) {
	return []domain.Room{}, nil
}
func (h *hotelRepoStub) GetHotel(context.Context, uuid.UUID) (domain.Hotel, error) {
	return domain.Hotel{ID: uuid.New(), Name: "H", Address: "A"}, nil
}
