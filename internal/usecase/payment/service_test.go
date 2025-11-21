package payment_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/payment"
	"github.com/ftryyln/hotel-booking-microservices/internal/usecase/payment"
	"github.com/ftryyln/hotel-booking-microservices/pkg/dto"
	"github.com/ftryyln/hotel-booking-microservices/pkg/valueobject"
)

func TestHandleWebhook(t *testing.T) {
	paymentID := uuid.New()
	repo := &paymentRepoStub{store: map[uuid.UUID]domain.Payment{
		paymentID: {ID: paymentID, BookingID: uuid.New(), Status: string(valueobject.PaymentPending)},
	}}
	provider := &providerStub{signatureValid: true}
	updater := &bookingUpdaterStub{}
	service := payment.NewService(repo, provider, updater)

	tests := []struct {
		name    string
		valid   bool
		wantErr bool
	}{
		{"valid", true, false},
		{"invalid signature", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider.signatureValid = tt.valid
			updater.statuses = nil
			req := dto.WebhookRequest{PaymentID: paymentID.String(), Status: domain.StatusPaid, Signature: "sig"}
			err := service.HandleWebhook(context.Background(), req, "payload")
			if tt.wantErr {
				require.Error(t, err)
				require.Len(t, updater.statuses, 0)
			} else {
				require.NoError(t, err)
				require.Equal(t, []string{"confirmed"}, updater.statuses)
			}
		})
	}
}

func TestRefund(t *testing.T) {
	paymentID := uuid.New()
	repo := &paymentRepoStub{store: map[uuid.UUID]domain.Payment{
		paymentID: {ID: paymentID},
	}}
	provider := &providerStub{signatureValid: true}
	service := payment.NewService(repo, provider, nil)

	tests := []struct {
		name       string
		shouldFail bool
	}{
		{"success", false},
		{"provider failure", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldFail {
				provider.refundErr = errors.New("fail")
			} else {
				provider.refundErr = nil
			}
			req := dto.RefundRequest{PaymentID: paymentID.String(), Amount: 1000}
			_, err := service.Refund(context.Background(), req)
			if tt.shouldFail {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// stubs

type paymentRepoStub struct {
	store map[uuid.UUID]domain.Payment
}

func (p *paymentRepoStub) Create(ctx context.Context, pay domain.Payment) error {
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
	for _, pay := range p.store {
		if pay.BookingID == bookingID {
			return pay, nil
		}
	}
	return domain.Payment{}, errors.New("not found")
}

func (p *paymentRepoStub) UpdateStatus(ctx context.Context, id uuid.UUID, status, paymentURL string) error {
	if pay, ok := p.store[id]; ok {
		pay.Status = status
		p.store[id] = pay
		return nil
	}
	return errors.New("not found")
}

func (p *paymentRepoStub) Initiate(context.Context, uuid.UUID, float64) (string, error) {
	return "", nil
}

type providerStub struct {
	signatureValid bool
	refundErr      error
}

func (p *providerStub) Initiate(ctx context.Context, payment domain.Payment) (domain.Payment, error) {
	return payment, nil
}

func (p *providerStub) VerifySignature(ctx context.Context, payload, signature string) bool {
	return p.signatureValid
}

func (p *providerStub) Refund(ctx context.Context, payment domain.Payment, reason string) (string, error) {
	return "ref", p.refundErr
}

type bookingUpdaterStub struct {
	statuses []string
}

func (b *bookingUpdaterStub) Update(ctx context.Context, bookingID uuid.UUID, status string) error {
	b.statuses = append(b.statuses, status)
	return nil
}
