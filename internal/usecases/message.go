package usecases

import (
	"context"
	"errors"
	"github.com/dontpanicw/DelayedNotifier/internal/domain"
	"github.com/dontpanicw/DelayedNotifier/internal/port"
	"github.com/google/uuid"
)

var _ port.Usecases = (*MessageUsecases)(nil)

type MessageUsecases struct {
	repo port.Repository
}

func NewMessageUsecases(repo port.Repository) *MessageUsecases {
	return &MessageUsecases{repo: repo}
}

func (m MessageUsecases) CreateAndSendMessage(ctx context.Context, message domain.Message) (string, error) {
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
	return message.Id, nil
}

func (m MessageUsecases) GetMessageStatus(ctx context.Context, id string) (string, error) {
	messageStatus, err := m.repo.GetMessageStatus(ctx, id)
	if err != nil {
		return "", err
	}
	return messageStatus, nil
}

func (m MessageUsecases) ListMessages(ctx context.Context) ([]domain.Message, error) {
	return m.repo.ListMessages(ctx)
}

func (m MessageUsecases) DeleteMessage(ctx context.Context, id string) error {
	return m.repo.DeleteMessage(ctx, id)
}
