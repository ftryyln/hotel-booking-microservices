package dto

// NotificationRequest is sent to notification service.
type NotificationRequest struct {
	Type    string `json:"type"`
	Target  string `json:"target"`
	Message string `json:"message"`
}
