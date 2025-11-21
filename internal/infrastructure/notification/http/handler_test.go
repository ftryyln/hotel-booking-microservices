package notificationhttp_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"

	notificationhttp "github.com/ftryyln/hotel-booking-microservices/internal/infrastructure/notification/http"
	"github.com/ftryyln/hotel-booking-microservices/internal/usecase/notification"
)

type dispatcherStub struct{ err error }

func (d *dispatcherStub) Dispatch(ctx context.Context, target, message string) error { return d.err }

func TestNotificationHandlerSendAndList(t *testing.T) {
	svc := notification.NewService(&dispatcherStub{})
	h := notificationhttp.NewHandler(svc)
	r := chi.NewRouter()
	r.Mount("/", h.Routes())

	req := httptest.NewRequest(http.MethodPost, "/notifications", strings.NewReader(`{"type":"email","target":"x","message":"hi"}`))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusAccepted, rec.Code)

	reqList := httptest.NewRequest(http.MethodGet, "/notifications", nil)
	recList := httptest.NewRecorder()
	r.ServeHTTP(recList, reqList)
	require.Equal(t, http.StatusOK, recList.Code)
}
