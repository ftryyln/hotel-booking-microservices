package dto

// PaymentRequest triggers payment provider.
type PaymentRequest struct {
	BookingID string  `json:"booking_id"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
}

// PaymentResponse describes created payment.
type PaymentResponse struct {
	ID         string `json:"id"`
	Status     string `json:"status"`
	Provider   string `json:"provider"`
	PaymentURL string `json:"payment_url"`
}

// WebhookRequest is provider callback payload.
type WebhookRequest struct {
	PaymentID string `json:"payment_id"`
	Status    string `json:"status"`
	Signature string `json:"signature"`
}

// RefundRequest triggers refunds.
type RefundRequest struct {
	PaymentID string  `json:"payment_id"`
	Amount    float64 `json:"amount"`
	Reason    string  `json:"reason"`
}

// RefundResponse describes refund status.
type RefundResponse struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	Reference string `json:"reference"`
}
