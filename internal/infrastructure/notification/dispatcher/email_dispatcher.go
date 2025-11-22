package dispatcher

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
	"time"

	domain "github.com/ftryyln/hotel-booking-microservices/internal/domain/notification"
)

// sendMailFn allows test override.
var sendMailFn = smtp.SendMail

// EmailDispatcher sends notifications via SMTP.
type EmailDispatcher struct {
	auth smtp.Auth
	host string
	from string
}

// NewEmailDispatcher creates an SMTP-backed dispatcher.
func NewEmailDispatcher(host string, port int, username, password, from string) domain.Dispatcher {
	addr := fmt.Sprintf("%s:%d", host, port)
	var auth smtp.Auth
	if username != "" && password != "" {
		auth = smtp.PlainAuth("", username, password, host)
	}
	return &EmailDispatcher{
		auth: auth,
		host: addr,
		from: from,
	}
}

// Dispatch sends email with target as recipient.
func (d *EmailDispatcher) Dispatch(ctx context.Context, target, message string) error {
	subject := "Notification " + time.Now().UTC().Format(time.RFC3339)
	body := fmt.Sprintf("Subject: %s\r\nFrom: %s\r\nTo: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n%s",
		subject, d.from, target, message)
	return sendMailFn(d.host, d.auth, d.from, []string{strings.TrimSpace(target)}, []byte(body))
}
