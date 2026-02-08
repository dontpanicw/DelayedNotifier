package postgres

import (
	"context"
	"github.com/dontpanicw/DelayedNotifier/config"
	"github.com/dontpanicw/DelayedNotifier/internal/domain"
	"github.com/dontpanicw/DelayedNotifier/internal/port"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"
	"time"
)

var (
	_ port.Repository = (*MessageRepository)(nil)
)

const (
	getMessageQuery = `
		select status
		from messages 
		where id = $1
		`
	createMessageQuery = `INSERT INTO messages (id, text, status, scheduled_at, user_id, telegram_chat_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		`
	deleteMessageQuery = `DELETE FROM messages 
       WHERE id = $1
       `
	listMessagesQuery = `SELECT id, text, status, scheduled_at, user_id, telegram_chat_id FROM messages ORDER BY created_at DESC`
)

type MessageRepository struct {
	PostgresDB *dbpg.DB
}

func NewMessageRepository(config *config.Config) *MessageRepository {
	opts := &dbpg.Options{MaxOpenConns: 10, MaxIdleConns: 5}
	db, err := dbpg.New(config.MasterDSN, config.SlaveDSNs, opts)
	if err != nil {
		panic(err)
	}

	return &MessageRepository{
		PostgresDB: db,
	}
}

func (m *MessageRepository) CreateMessage(ctx context.Context, message domain.Message) error {
	_, err := m.PostgresDB.ExecWithRetry(ctx, createRetryStrategy(), createMessageQuery, message.Id, message.Text, message.Status, message.ScheduledAt, message.UserId, message.TelegramChatId)
	if err != nil {
		return err
	}
	return nil
}

func (m *MessageRepository) GetMessageStatus(ctx context.Context, id string) (string, error) {
	var messageStatus string
	err := m.PostgresDB.QueryRowContext(ctx, getMessageQuery, id).Scan(&messageStatus)
	if err != nil {
		return "", err
	}
	return messageStatus, nil
}

func (m *MessageRepository) ListMessages(ctx context.Context) ([]domain.Message, error) {
	rows, err := m.PostgresDB.QueryContext(ctx, listMessagesQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []domain.Message
	for rows.Next() {
		var msg domain.Message
		var userID, chatID int64
		if err := rows.Scan(&msg.Id, &msg.Text, &msg.Status, &msg.ScheduledAt, &userID, &chatID); err != nil {
			return nil, err
		}
		msg.UserId = uint32(userID)
		msg.TelegramChatId = uint32(chatID)
		messages = append(messages, msg)
	}
	return messages, rows.Err()
}

func (m *MessageRepository) DeleteMessage(ctx context.Context, id string) error {
	_, err := m.PostgresDB.ExecWithRetry(ctx, createRetryStrategy(), deleteMessageQuery, id)
	if err != nil {
		return err
	}
	return nil
}

func createRetryStrategy() retry.Strategy {
	return retry.Strategy{
		Attempts: 3,
		Delay:    5 * time.Second,
		Backoff:  2}
}
