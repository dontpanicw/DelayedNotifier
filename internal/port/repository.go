package port

import (
	"context"
	"github.com/dontpanicw/DelayedNotifier/internal/domain"
)

type Repository interface {
	CreateMessage(ctx context.Context, message domain.Message) error
	GetMessageStatus(ctx context.Context, id string) (string, error)
	ListMessages(ctx context.Context) ([]domain.Message, error)
	UpdateMessageStatus(ctx context.Context, id, status string) error
	DeleteMessage(ctx context.Context, id string) error
}
