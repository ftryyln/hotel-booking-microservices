package notification_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ftryyln/hotel-booking-microservices/internal/usecase/notification"
	"github.com/ftryyln/hotel-booking-microservices/pkg/dto"
)

func TestSendAndList(t *testing.T) {
	dispatcher := &dispatcherStub{}
	svc := notification.NewService(dispatcher)

	req := dto.NotificationRequest{Type: "email", Target: "user@example.com", Message: "hello"}
	resp, err := svc.Send(context.Background(), req)
	require.NoError(t, err)
	require.NotEmpty(t, resp.ID)

	list := svc.List(context.Background())
	require.Len(t, list, 1)

	found, ok := svc.Get(context.Background(), resp.ID)
	require.True(t, ok)
	require.Equal(t, resp.ID, found.ID)
}

func TestSendDispatchError(t *testing.T) {
	dispatcher := &dispatcherStub{err: errors.New("fail")}
	svc := notification.NewService(dispatcher)
	_, err := svc.Send(context.Background(), dto.NotificationRequest{Type: "email", Target: "x", Message: "y"})
	require.Error(t, err)
}

type dispatcherStub struct {
	err error
}

func (d *dispatcherStub) Dispatch(ctx context.Context, target, message string) error {
	return d.err
}
