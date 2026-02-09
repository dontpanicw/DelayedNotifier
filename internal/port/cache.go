package port

import (
	"context"
	"time"
)

// StatusCache описывает кэш для быстрого получения статуса сообщения.
type StatusCache interface {
	GetStatus(ctx context.Context, id string) (string, error)
	SetStatus(ctx context.Context, id, status string, ttl time.Duration) error
}

