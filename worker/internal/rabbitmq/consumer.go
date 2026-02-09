package rabbitmq

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/dontpanicw/DelayedNotifier/internal/domain"
	"github.com/dontpanicw/DelayedNotifier/internal/port"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	workerExchangeName = "notifications.exchange"
	workerQueueName    = "notifications.queue"
	workerRoutingKey   = "notifications.create"
)

// retryDelays задаёт экспоненциальную политику ретраев:
// первая попытка сразу, затем 10с, 40с, 90с.
var retryDelays = []time.Duration{
	10 * time.Second,
	40 * time.Second,
	90 * time.Second,
}

type MessageQueueConsumer struct {
	conn  *amqp.Connection
	ch    *amqp.Channel
	repo  port.Repository
	cache port.StatusCache
}

func NewMessageQueueConsumer(rabbitURL string, repo port.Repository, cache port.StatusCache) (*MessageQueueConsumer, error) {
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	if err := ch.ExchangeDeclare(
		workerExchangeName,
		"direct",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}

	q, err := ch.QueueDeclare(
		workerQueueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}

	if err := ch.QueueBind(
		q.Name,
		workerRoutingKey,
		workerExchangeName,
		false,
		nil,
	); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}

	return &MessageQueueConsumer{
		conn:  conn,
		ch:    ch,
		repo:  repo,
		cache: cache,
	}, nil
}

func (c *MessageQueueConsumer) Start(ctx context.Context) error {
	if err := c.ch.Qos(1, 0, false); err != nil {
		return err
	}

	msgs, err := c.ch.Consume(
		workerQueueName,
		"delayed_notifier_worker",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case d, ok := <-msgs:
			if !ok {
				return nil
			}
			go c.handleDelivery(ctx, d)
		}
	}
}

func (c *MessageQueueConsumer) handleDelivery(ctx context.Context, d amqp.Delivery) {
	var msg domain.Message
	if err := json.Unmarshal(d.Body, &msg); err != nil {
		log.Printf("failed to unmarshal message: %v", err)
		_ = d.Nack(false, false)
		return
	}

	delay := time.Until(msg.ScheduledAt)
	if delay > 0 {
		timer := time.NewTimer(delay)
		defer timer.Stop()

		select {
		case <-ctx.Done():
			_ = d.Nack(false, true)
			return
		case <-timer.C:
		}
	}

	// Отправка сообщения с экспоненциальной политикой ретраев.
	if err := c.sendWithRetry(ctx, &msg); err != nil {
		log.Printf("failed to send message after retries: %v", err)
		// помечаем как терминально упавшее
		if err2 := c.repo.UpdateMessageStatus(ctx, msg.Id, domain.JobStatusTerminallyFailed); err2 != nil {
			log.Printf("failed to mark message terminally failed: %v", err2)
		}
		if c.cache != nil {
			_ = c.cache.SetStatus(ctx, msg.Id, domain.JobStatusTerminallyFailed, 5*time.Minute)
		}
		// сообщение обработано (больше не будет ретраев из очереди)
		_ = d.Ack(false)
		return
	}

	// успешная отправка
	if c.cache != nil {
		_ = c.cache.SetStatus(ctx, msg.Id, domain.JobStatusSent, 5*time.Minute)
	}

	if err := d.Ack(false); err != nil {
		log.Printf("failed to ack message: %v", err)
	}
}

func (c *MessageQueueConsumer) Close() {
	if c.ch != nil {
		_ = c.ch.Close()
	}
	if c.conn != nil {
		_ = c.conn.Close()
	}
}

// sendWithRetry реализует экспоненциальную политику повторных попыток отправки.
// Сейчас в качестве "отправки" используется обновление статуса в БД на Sent —
// сюда же можно встроить реальный вызов внешнего сервиса (Telegram/email),
// а UpdateMessageStatus вызывать после успешной доставки.
func (c *MessageQueueConsumer) sendWithRetry(ctx context.Context, msg *domain.Message) error {
	var lastErr error
	attempts := len(retryDelays) + 1 // первая попытка без задержки + ретраи

	for attempt := 0; attempt < attempts; attempt++ {
		if attempt > 0 {
			delay := retryDelays[attempt-1]
			log.Printf("retry attempt %d for message %s after %s", attempt+1, msg.Id, delay)
			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
		}

		// здесь должно быть реальное отправление во внешний сервис;
		// сейчас считаем, что успешное обновление статуса = успех отправки.
		if err := c.repo.UpdateMessageStatus(ctx, msg.Id, domain.JobStatusSent); err == nil {
			return nil
		} else {
			lastErr = err
			log.Printf("attempt %d to update status for %s failed: %v", attempt+1, msg.Id, err)
		}
	}

	return lastErr
}


