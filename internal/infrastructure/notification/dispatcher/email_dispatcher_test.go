package dispatcher

import (
	"context"
	"errors"
	"net/smtp"
	"testing"
)

func TestEmailDispatcher_Dispatch(t *testing.T) {
	calls := 0
	sendMailFn = func(addr string, _ smtp.Auth, from string, to []string, msg []byte) error {
		calls++
		if addr != "smtp.example.com:587" {
			t.Fatalf("addr mismatch: %s", addr)
		}
		if from != "no-reply@example.com" {
			t.Fatalf("from mismatch: %s", from)
		}
		if len(to) != 1 || to[0] != "user@example.com" {
			t.Fatalf("to mismatch: %+v", to)
		}
		if len(msg) == 0 {
			t.Fatalf("msg empty")
		}
		return nil
	}
	defer func() { sendMailFn = smtp.SendMail }()

	d := NewEmailDispatcher("smtp.example.com", 587, "user", "pass", "no-reply@example.com")
	if err := d.Dispatch(context.Background(), "user@example.com", "hello"); err != nil {
		t.Fatalf("dispatch err: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestEmailDispatcher_DispatchError(t *testing.T) {
	sendMailFn = func(string, smtp.Auth, string, []string, []byte) error {
		return errors.New("fail")
	}
	defer func() { sendMailFn = smtp.SendMail }()
	d := NewEmailDispatcher("smtp.example.com", 587, "", "", "no-reply@example.com")
	if err := d.Dispatch(context.Background(), "user@example.com", "hello"); err == nil {
		t.Fatalf("expected error")
	}
}
