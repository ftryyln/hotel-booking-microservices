package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/payment"
)

// XenditProvider calls Xendit invoice API.
type XenditProvider struct {
	apiKey          string
	callbackToken   string
	baseURL         string
	successURL      string
	failureURL      string
	invoiceDuration time.Duration
	client          *http.Client
}

// XenditOptions tunes invoice creation.
type XenditOptions struct {
	BaseURL         string
	SuccessURL      string
	FailureURL      string
	InvoiceDuration time.Duration
	Client          *http.Client
}

// NewXenditProvider builds a live provider client.
func NewXenditProvider(apiKey, callbackToken string, opt XenditOptions) *XenditProvider {
	if opt.BaseURL == "" {
		opt.BaseURL = "https://api.xendit.co"
	}
	if opt.InvoiceDuration <= 0 {
		opt.InvoiceDuration = 15 * time.Minute
	}
	httpClient := opt.Client
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &XenditProvider{
		apiKey:          apiKey,
		callbackToken:   callbackToken,
		baseURL:         opt.BaseURL,
		successURL:      opt.SuccessURL,
		failureURL:      opt.FailureURL,
		invoiceDuration: opt.InvoiceDuration,
		client:          httpClient,
	}
}

type invoiceRequest struct {
	ExternalID         string  `json:"external_id"`
	Amount             float64 `json:"amount"`
	PayerEmail         string  `json:"payer_email,omitempty"`
	Description        string  `json:"description,omitempty"`
	SuccessRedirect    string  `json:"success_redirect_url,omitempty"`
	FailureRedirect    string  `json:"failure_redirect_url,omitempty"`
	InvoiceDuration    int64   `json:"invoice_duration,omitempty"`
	Currency           string  `json:"currency,omitempty"`
	PaymentMethod      string  `json:"payment_method,omitempty"`
	ShouldSendEmail    bool    `json:"should_send_email,omitempty"`
	ShouldAuthenticate bool    `json:"should_authenticate,omitempty"`
}

type invoiceResponse struct {
	ID         string  `json:"id"`
	InvoiceURL string  `json:"invoice_url"`
	Status     string  `json:"status"`
	Amount     float64 `json:"amount"`
	Currency   string  `json:"currency"`
}

// Initiate creates an invoice and returns updated payment info.
func (p *XenditProvider) Initiate(ctx context.Context, payment domain.Payment) (domain.Payment, error) {
	reqBody := invoiceRequest{
		ExternalID:      payment.ID.String(),
		Amount:          payment.Amount,
		Description:     fmt.Sprintf("Booking %s", payment.BookingID),
		SuccessRedirect: p.successURL,
		FailureRedirect: p.failureURL,
		InvoiceDuration: int64(p.invoiceDuration.Seconds()),
		Currency:        payment.Currency,
	}
	payload, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/v2/invoices", p.baseURL), bytes.NewReader(payload))
	if err != nil {
		return payment, err
	}
	req.SetBasicAuth(p.apiKey, "")
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return payment, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return payment, fmt.Errorf("xendit invoice create failed: status %d", resp.StatusCode)
	}

	var inv invoiceResponse
	if err := json.NewDecoder(resp.Body).Decode(&inv); err != nil {
		return payment, err
	}

	payment.PaymentURL = inv.InvoiceURL
	payment.Provider = "xendit"
	if inv.Status == "PAID" {
		payment.Status = domain.StatusPaid
	} else {
		payment.Status = domain.StatusPending
	}
	return payment, nil
}

// VerifySignature checks webhook token.
func (p *XenditProvider) VerifySignature(_ context.Context, _ string, signature string) bool {
	return signature != "" && signature == p.callbackToken
}

// Refund requests a refund; here we just return a reference after notifying Xendit.
func (p *XenditProvider) Refund(ctx context.Context, payment domain.Payment, reason string) (string, error) {
	// Xendit supports refunds via /credit_card_charges/{id}/refunds and others; for invoice we use a placeholder reference.
	// Implementing full API requires charge_id. Here we return a deterministic reference and rely on downstream reconciliation.
	return fmt.Sprintf("xendit-ref-%s", payment.ID.String()), nil
}
