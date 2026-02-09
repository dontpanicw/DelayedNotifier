package port

import (
	"context"
	"github.com/dontpanicw/DelayedNotifier/internal/domain"
)

type MessageQueue interface {
	SendMessage(ctx context.Context, message domain.Message) error
}
