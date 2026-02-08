package port

import (
	"context"
	"github.com/dontpanicw/DelayedNotifier/internal/domain"
)

type Usecases interface {
	CreateAndSendMessage(ctx context.Context, message domain.Message) (string, error)
	GetMessageStatus(ctx context.Context, id string) (string, error)
	ListMessages(ctx context.Context) ([]domain.Message, error)
	DeleteMessage(ctx context.Context, id string) error
}
