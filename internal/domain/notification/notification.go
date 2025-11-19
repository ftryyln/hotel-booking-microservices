package notification

import "context"

// Dispatcher sends notifications.
type Dispatcher interface {
	Dispatch(ctx context.Context, target, message string) error
}
