package payment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/booking"
)

// HTTPGateway calls payment service over HTTP.
type HTTPGateway struct {
	baseURL string
	client  *http.Client
}

func NewHTTPGateway(baseURL string) domain.PaymentGateway {
	return &HTTPGateway{baseURL: baseURL, client: &http.Client{Timeout: 5 * time.Second}}
}

func (g *HTTPGateway) Initiate(ctx context.Context, bookingID uuid.UUID, amount float64) (string, error) {
	payload := map[string]any{"booking_id": bookingID.String(), "amount": amount, "currency": "IDR"}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/payments", g.baseURL), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := g.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("payment initiation failed: %d", resp.StatusCode)
	}
	var result struct {
		PaymentURL string `json:"payment_url"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&result)
	return result.PaymentURL, nil
}
