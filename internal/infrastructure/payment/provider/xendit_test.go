package provider

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/payment"
	"github.com/google/uuid"
)

func TestXenditProvider_Initiate(t *testing.T) {
	var receivedAuth string
	var receivedBody string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/invoices" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		receivedAuth = r.Header.Get("Authorization")
		bodyBytes, _ := io.ReadAll(r.Body)
		receivedBody = string(bodyBytes)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"inv_123","invoice_url":"https://pay.test/inv_123","status":"PENDING","amount":100000,"currency":"IDR"}`))
	}))
	defer ts.Close()

	opt := XenditOptions{BaseURL: ts.URL, SuccessURL: "https://success", FailureURL: "https://fail", InvoiceDuration: 10 * time.Minute, Client: ts.Client()}
	prov := NewXenditProvider("secret-key", "token", opt)
	pay := domain.Payment{ID: uuid.New(), BookingID: uuid.New(), Amount: 100000, Currency: "IDR"}

	got, err := prov.Initiate(context.Background(), pay)
	if err != nil {
		t.Fatalf("initiate err: %v", err)
	}
	if !strings.HasPrefix(receivedAuth, "Basic ") {
		t.Fatalf("expected basic auth header")
	}
	if got.PaymentURL != "https://pay.test/inv_123" {
		t.Fatalf("payment url mismatch: %s", got.PaymentURL)
	}
	if got.Status != domain.StatusPending {
		t.Fatalf("status mismatch: %s", got.Status)
	}
	if !strings.Contains(receivedBody, `"external_id":"`+pay.ID.String()+`"`) {
		t.Fatalf("body missing external_id: %s", receivedBody)
	}
}

func TestXenditProvider_VerifySignature(t *testing.T) {
	prov := NewXenditProvider("key", "token123", XenditOptions{})
	if !prov.VerifySignature(context.Background(), "", "token123") {
		t.Fatalf("expected signature to match")
	}
	if prov.VerifySignature(context.Background(), "", "wrong") {
		t.Fatalf("expected signature mismatch")
	}
}
