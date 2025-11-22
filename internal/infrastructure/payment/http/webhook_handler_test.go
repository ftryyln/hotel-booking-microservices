package paymenthttp_test

import (
	"bytes"
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

func TestWebhookHandler_HeaderSignatureFallback(t *testing.T) {
	pid := uuid.New()
	repo := &paymentRepoStub2{store: map[uuid.UUID]domain.Payment{
		pid: {ID: pid, BookingID: uuid.New(), Status: "pending"},
	}}
	prov := &providerStub2{}
	svc := paymentuc.NewService(repo, prov, &bookingUpdaterStub2{})
	h := paymenthttp.NewHandler(svc)

	r := chi.NewRouter()
	r.Mount("/", h.Routes())

	body := []byte(`{"payment_id":"` + pid.String() + `","status":"paid"}`)
	req := httptest.NewRequest(http.MethodPost, "/payments/webhook", bytes.NewReader(body))
	req.Header.Set("X-CALLBACK-TOKEN", "header-token")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "header-token", repo.store[pid].WebhookSignature)
}

type paymentRepoStub2 struct {
	store map[uuid.UUID]domain.Payment
}

func (p *paymentRepoStub2) Create(_ context.Context, pay domain.Payment) error {
	if p.store == nil {
		p.store = map[uuid.UUID]domain.Payment{}
	}
	p.store[pay.ID] = pay
	return nil
}
func (p *paymentRepoStub2) FindByID(_ context.Context, id uuid.UUID) (domain.Payment, error) {
	if val, ok := p.store[id]; ok {
		return val, nil
	}
	return domain.Payment{}, errors.New("not found")
}
func (p *paymentRepoStub2) FindByBookingID(_ context.Context, bookingID uuid.UUID) (domain.Payment, error) {
	for _, v := range p.store {
		if v.BookingID == bookingID {
			return v, nil
		}
	}
	return domain.Payment{}, errors.New("not found")
}
func (p *paymentRepoStub2) UpdateStatus(_ context.Context, id uuid.UUID, status, url, payload, signature string) error {
	pay := p.store[id]
	pay.Status = status
	pay.PaymentURL = url
	pay.WebhookPayload = payload
	pay.WebhookSignature = signature
	p.store[id] = pay
	return nil
}

type providerStub2 struct{}

func (p *providerStub2) Initiate(_ context.Context, pay domain.Payment) (domain.Payment, error) {
	return pay, nil
}
func (p *providerStub2) VerifySignature(_ context.Context, payload, signature string) bool {
	return signature == "header-token"
}
func (p *providerStub2) Refund(_ context.Context, pay domain.Payment, reason string) (string, error) {
	return "ref", nil
}

type bookingUpdaterStub2 struct{}

func (b *bookingUpdaterStub2) Update(_ context.Context, bookingID uuid.UUID, status string) error {
	return nil
}
