package paymenthttp_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/payment"
	paymenthttp "github.com/ftryyln/hotel-booking-microservices/internal/infrastructure/payment/http"
	paymentuc "github.com/ftryyln/hotel-booking-microservices/internal/usecase/payment"
)

func TestPaymentHandler_GetPayment(t *testing.T) {
	repo := &paymentRepoStub{store: map[uuid.UUID]domain.Payment{}}
	id := uuid.New()
	repo.store[id] = domain.Payment{ID: id, Status: "pending"}
	svc := paymentuc.NewService(repo, &providerStub{}, &bookingUpdaterStub{})
	h := paymenthttp.NewHandler(svc)

	r := chi.NewRouter()
	r.Mount("/", h.Routes())

	req := httptest.NewRequest(http.MethodGet, "/payments/"+id.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

type paymentRepoStub struct {
	store map[uuid.UUID]domain.Payment
}

func (p *paymentRepoStub) Create(ctx context.Context, pay domain.Payment) error {
	if p.store == nil {
		p.store = map[uuid.UUID]domain.Payment{}
	}
	p.store[pay.ID] = pay
	return nil
}
func (p *paymentRepoStub) FindByID(ctx context.Context, id uuid.UUID) (domain.Payment, error) {
	pay, ok := p.store[id]
	if !ok {
		return domain.Payment{}, errors.New("not found")
	}
	return pay, nil
}
func (p *paymentRepoStub) FindByBookingID(ctx context.Context, bookingID uuid.UUID) (domain.Payment, error) {
	for _, v := range p.store {
		if v.BookingID == bookingID {
			return v, nil
		}
	}
	return domain.Payment{}, errors.New("not found")
}
func (p *paymentRepoStub) UpdateStatus(ctx context.Context, id uuid.UUID, status, url string) error {
	pay := p.store[id]
	pay.Status = status
	p.store[id] = pay
	return nil
}

type providerStub struct{}

func (p *providerStub) Initiate(ctx context.Context, pay domain.Payment) (domain.Payment, error) {
	return pay, nil
}
func (p *providerStub) VerifySignature(ctx context.Context, payload, signature string) bool {
	return true
}
func (p *providerStub) Refund(ctx context.Context, pay domain.Payment, reason string) (string, error) {
	return "ref", nil
}

type bookingUpdaterStub struct{}

func (b *bookingUpdaterStub) Update(ctx context.Context, bookingID uuid.UUID, status string) error {
	return nil
}
