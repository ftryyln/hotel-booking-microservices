package provider

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/payment"
)

// XenditMockProvider simulates payment provider.
type XenditMockProvider struct {
	secret string
}

func NewXenditMockProvider(secret string) *XenditMockProvider {
	return &XenditMockProvider{secret: secret}
}

func (p *XenditMockProvider) Initiate(ctx context.Context, payment domain.Payment) (domain.Payment, error) {
	payment.PaymentURL = fmt.Sprintf("https://sandbox.xendit.co/pay/%s", payment.ID)
	payment.Status = domain.StatusPending
	return payment, nil
}

func (p *XenditMockProvider) VerifySignature(ctx context.Context, payload, signature string) bool {
	h := hmac.New(sha256.New, []byte(p.secret))
	h.Write([]byte(payload))
	return hmac.Equal([]byte(signature), []byte(hex.EncodeToString(h.Sum(nil))))
}

func (p *XenditMockProvider) Refund(ctx context.Context, payment domain.Payment, reason string) (string, error) {
	return fmt.Sprintf("rf_%s_%d", payment.ID.String(), time.Now().Unix()), nil
}
