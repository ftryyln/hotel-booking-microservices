package booking

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/payment"
)

// HTTPStatusClient notifies booking service of payments.
type HTTPStatusClient struct {
	baseURL string
	client  *http.Client
}

func NewHTTPStatusClient(baseURL string) domain.BookingStatusUpdater {
	return &HTTPStatusClient{baseURL: baseURL, client: &http.Client{Timeout: 5 * time.Second}}
}

func (c *HTTPStatusClient) Update(ctx context.Context, bookingID uuid.UUID, status string) error {
	payload := map[string]string{"status": status}
	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/bookings/%s/status", c.baseURL, bookingID.String())
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("failed to update booking status: %d", resp.StatusCode)
	}
	return nil
}
