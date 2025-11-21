package bookinghttp_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/booking"
	hdomain "github.com/ftryyln/hotel-booking-microservices/internal/domain/hotel"
	bookinghttp "github.com/ftryyln/hotel-booking-microservices/internal/infrastructure/booking/http"
	"github.com/ftryyln/hotel-booking-microservices/internal/usecase/booking"
	"github.com/ftryyln/hotel-booking-microservices/pkg/query"
)

func TestBookingHandlerListWithPagination(t *testing.T) {
	repo := &bookingRepoStub{
		store: map[uuid.UUID]domain.Booking{
			uuid.New(): {ID: uuid.New(), Status: "pending_payment", CheckIn: time.Now(), CheckOut: time.Now().Add(24 * time.Hour)},
		},
	}
	hRepo := &hotelRepoStub{}
	svc := booking.NewService(repo, hRepo, &paymentGatewayStub{}, &notificationGatewayStub{})
	h := bookinghttp.NewHandler(svc)

	r := chi.NewRouter()
	r.Mount("/", h.Routes())

	req := httptest.NewRequest(http.MethodGet, "/bookings?limit=1&offset=0", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

// stubs for booking handler test
type bookingRepoStub struct {
	store map[uuid.UUID]domain.Booking
}

func (b *bookingRepoStub) Create(ctx context.Context, bk domain.Booking) error { return nil }
func (b *bookingRepoStub) FindByID(ctx context.Context, id uuid.UUID) (domain.Booking, error) {
	return b.store[id], nil
}
func (b *bookingRepoStub) List(ctx context.Context, opts query.Options) ([]domain.Booking, error) {
	var out []domain.Booking
	for _, v := range b.store {
		out = append(out, v)
	}
	return out, nil
}
func (b *bookingRepoStub) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	return nil
}

type hotelRepoStub struct{}

func (h *hotelRepoStub) CreateHotel(context.Context, hdomain.Hotel) error { return nil }
func (h *hotelRepoStub) ListHotels(ctx context.Context, opts query.Options) ([]hdomain.Hotel, error) {
	return nil, nil
}
func (h *hotelRepoStub) CreateRoomType(context.Context, hdomain.RoomType) error { return nil }
func (h *hotelRepoStub) ListRoomTypes(context.Context, uuid.UUID) ([]hdomain.RoomType, error) {
	return nil, nil
}
func (h *hotelRepoStub) ListAllRoomTypes(context.Context, query.Options) ([]hdomain.RoomType, error) {
	return nil, nil
}
func (h *hotelRepoStub) CreateRoom(context.Context, hdomain.Room) error { return nil }
func (h *hotelRepoStub) GetRoomType(context.Context, uuid.UUID) (hdomain.RoomType, error) {
	return hdomain.RoomType{}, nil
}
func (h *hotelRepoStub) ListRooms(context.Context, query.Options) ([]hdomain.Room, error) {
	return nil, nil
}
func (h *hotelRepoStub) GetHotel(context.Context, uuid.UUID) (hdomain.Hotel, error) {
	return hdomain.Hotel{}, nil
}

type paymentGatewayStub struct{}

func (p *paymentGatewayStub) Initiate(context.Context, uuid.UUID, float64) (domain.PaymentResult, error) {
	return domain.PaymentResult{
		ID:         uuid.New(),
		Status:     "pending",
		Provider:   "mock",
		PaymentURL: "http://mock",
	}, nil
}

type notificationGatewayStub struct{}

func (n *notificationGatewayStub) Notify(context.Context, string, any) error { return nil }
