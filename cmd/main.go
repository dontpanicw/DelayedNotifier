package main

import (
	"github.com/dontpanicw/DelayedNotifier/config"
	"github.com/dontpanicw/DelayedNotifier/internal/app"
	"log"
)

func main() {

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal("error creating config")
	}

	if err := app.Start(cfg); err != nil {
		log.Fatal("failed to start application")
	}

}
