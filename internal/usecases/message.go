package usecases

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/dontpanicw/DelayedNotifier/internal/domain"
	"github.com/dontpanicw/DelayedNotifier/internal/port"
	"github.com/google/uuid"
)

var _ port.Usecases = (*MessageUsecases)(nil)

type MessageUsecases struct {
	repo  port.Repository
	queue port.MessageQueue
	cache port.StatusCache
}

func NewMessageUsecases(repo port.Repository, queue port.MessageQueue, cache port.StatusCache) *MessageUsecases {
	return &MessageUsecases{
		repo:  repo,
		queue: queue,
		cache: cache,
	}
}

func (m *MessageUsecases) CreateAndSendMessage(ctx context.Context, message domain.Message) (string, error) {
	if message.UserId <= 0 {
		return "", errors.New("userId should be greater than zero")
	}
	message.Id = uuid.NewString()
	message.Status = domain.JobStatusScheduled
	err := m.repo.CreateMessage(ctx, message)
	if err != nil {
		return "", err
	}
	//TODO: отправить в rabbitmq
	err = m.queue.SendMessage(ctx, message)
	if err != nil {
		return "", err
	}
	log.Printf("message %s sent", message.Id)
	return message.Id, nil
}

func (m *MessageUsecases) GetMessageStatus(ctx context.Context, id string) (string, error) {
	if m.cache != nil {
		if status, err := m.cache.GetStatus(ctx, id); err == nil && status != "" {
			return status, nil
		}
	}

	messageStatus, err := m.repo.GetMessageStatus(ctx, id)
	if err != nil {
		return "", err
	}
	if m.cache != nil {
		_ = m.cache.SetStatus(ctx, id, messageStatus, 5*time.Minute)
	}
	return messageStatus, nil
}

func (m *MessageUsecases) ListMessages(ctx context.Context) ([]domain.Message, error) {
	return m.repo.ListMessages(ctx)
}

func (m *MessageUsecases) DeleteMessage(ctx context.Context, id string) error {
	return m.repo.DeleteMessage(ctx, id)
}
