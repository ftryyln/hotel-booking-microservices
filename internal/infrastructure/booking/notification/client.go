package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/booking"
)

// HTTPGateway notifies notification service.
type HTTPGateway struct {
	baseURL string
	client  *http.Client
}

func NewHTTPGateway(baseURL string) domain.NotificationGateway {
	return &HTTPGateway{baseURL: baseURL, client: &http.Client{Timeout: 3 * time.Second}}
}

func (g *HTTPGateway) Notify(ctx context.Context, event string, payload any) error {
	body, _ := json.Marshal(map[string]any{"type": event, "target": "booking", "message": payload})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, g.baseURL+"/notifications", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := g.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("notify failed with status %d", resp.StatusCode)
	}
	return nil
}
