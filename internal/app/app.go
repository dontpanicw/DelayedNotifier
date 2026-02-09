package app

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/lib/pq"

	"github.com/dontpanicw/DelayedNotifier/config"
	redisCache "github.com/dontpanicw/DelayedNotifier/internal/adapter/cache/redis"
	"github.com/dontpanicw/DelayedNotifier/internal/adapter/rabbitmq"
	http2 "github.com/dontpanicw/DelayedNotifier/internal/input/http"
	"github.com/dontpanicw/DelayedNotifier/internal/adapter/repository/postgres"
	"github.com/dontpanicw/DelayedNotifier/internal/usecases"
	migrations "github.com/dontpanicw/DelayedNotifier/pkg/migration/postgres"
)

func Start(cfg *config.Config) error {
	db, err := sql.Open("postgres", cfg.MasterDSN)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := migrations.Migrate(db); err != nil {
		return err
	}
	log.Print("Migrations applied successfully")

	messageRepo := postgres.NewMessageRepository(cfg)
	statusCache := redisCache.NewStatusCache(cfg.RedisAddr)

	messageQueue, err := rabbitmq.NewMessageQueueProducer(cfg.RabbitURL)
	if err != nil {
		log.Fatal(err)
	}
	defer messageQueue.Close()

	messageUsecase := usecases.NewMessageUsecases(messageRepo, messageQueue, statusCache)

	srv := http2.NewServer(messageUsecase)

	log.Printf("Starting server on %s", cfg.HTTPPort)
	return http.ListenAndServe(cfg.HTTPPort, srv)
}
