package rabbitmq

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/dontpanicw/DelayedNotifier/internal/domain"
	"github.com/dontpanicw/DelayedNotifier/internal/port"
	"github.com/rabbitmq/amqp091-go"
	"github.com/wb-go/wbf/rabbitmq"
	"github.com/wb-go/wbf/retry"
)

const (
	defaultExchangeName   = "notifications.exchange"
	defaultRoutingKeyName = "notifications.create"
)

var _ port.MessageQueue = (*MessageQueueProducer)(nil)

type MessageQueueProducer struct {
	Client   *rabbitmq.RabbitClient
	Exchange string
}

func NewMessageQueueProducer(rabbitURL string) (*MessageQueueProducer, error) {
	strategy := retry.Strategy{
		Attempts: 3,
		Delay:    3 * time.Second,
		Backoff:  2,
	}
	cfg := rabbitmq.ClientConfig{
		URL:            rabbitURL,
		ConnectionName: "delayed_notifier",
		ConnectTimeout: 5 * time.Second,
		Heartbeat:      10 * time.Second,
		ReconnectStrat: strategy,
		ProducingStrat: strategy,
		ConsumingStrat: strategy,
	}

	client, err := rabbitmq.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	err = client.DeclareExchange(
		defaultExchangeName,
		"direct", // простой direct‑exchange
		true,     // durable
		false,    // autoDelete
		false,    // internal
		nil,      // no extra args
	)

	if err != nil {
		client.Close()
		return nil, err
	}

	return &MessageQueueProducer{
		Client:   client,
		Exchange: defaultExchangeName,
	}, nil

}

func (mp *MessageQueueProducer) SendMessage(ctx context.Context, message domain.Message) error {
	publisher := rabbitmq.NewPublisher(mp.Client, mp.Exchange, "application/json")

	bodyMsg, err := json.Marshal(message)
	if err != nil {
		return err
	}
	routingKey := defaultRoutingKeyName
	err = publisher.Publish(
		ctx,
		bodyMsg,
		routingKey,
		rabbitmq.WithExpiration(5*time.Minute),
		rabbitmq.WithHeaders(amqp091.Table{"x-service": "auth"}),
	)
	if err != nil {
		log.Printf("publish to RabbitMQ failed: %v", err)
		return err
	}
	return nil
}

func (mp *MessageQueueProducer) Close() {
	if mp.Client != nil {
		mp.Client.Close()
	}
}
