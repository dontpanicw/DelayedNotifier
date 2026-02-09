package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/dontpanicw/DelayedNotifier/config"
	redisCache "github.com/dontpanicw/DelayedNotifier/internal/adapter/cache/redis"
	"github.com/dontpanicw/DelayedNotifier/internal/adapter/repository/postgres"
	workerRabbit "github.com/dontpanicw/DelayedNotifier/worker/internal/rabbitmq"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	repo := postgres.NewMessageRepository(cfg)
	cache := redisCache.NewStatusCache(cfg.RedisAddr)

	consumer, err := workerRabbit.NewMessageQueueConsumer(cfg.RabbitURL, repo, cache)
	if err != nil {
		log.Fatalf("failed to create RabbitMQ consumer: %v", err)
	}
	defer consumer.Close()

	log.Println("worker started, waiting for messages...")
	if err := consumer.Start(ctx); err != nil {
		log.Fatalf("worker stopped with error: %v", err)
	}
}

