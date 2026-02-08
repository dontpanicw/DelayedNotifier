package app

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/lib/pq"

	"github.com/dontpanicw/DelayedNotifier/config"
	"github.com/dontpanicw/DelayedNotifier/internal/adapter/repository/postgres"
	"github.com/dontpanicw/DelayedNotifier/internal/input"
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
	messageUsecase := usecases.NewMessageUsecases(messageRepo)
	srv := input.NewServer(messageUsecase)

	log.Printf("Starting server on %s", cfg.HTTPPort)
	return http.ListenAndServe(cfg.HTTPPort, srv)
}
