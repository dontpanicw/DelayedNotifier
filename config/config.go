package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Config struct {
	HTTPPort  string
	MasterDSN string
	SlaveDSNs []string
	RabbitURL string
	RedisAddr string
}

const DefaultHTTPPort = ":8080"

func NewConfig() (*Config, error) {
	cfg := Config{}

	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		cfg.HTTPPort = DefaultHTTPPort
	} else {
		// Ensure port starts with ':' if not already present
		if len(httpPort) > 0 && httpPort[0] != ':' {
			cfg.HTTPPort = ":" + httpPort
		} else {
			cfg.HTTPPort = httpPort
		}
	}
	masterDSN := os.Getenv("MASTER_DSN")
	if masterDSN != "" {
		cfg.MasterDSN = masterDSN
	}
	slaveDSNs := make([]string, 0)
	slaveDSN := os.Getenv("SLAVE_DSN")
	slaveDSNs = append(slaveDSNs, slaveDSN)

	rabbitURL := os.Getenv("RABBIT_URL")
	if rabbitURL != "" {
		cfg.RabbitURL = rabbitURL
	} else {
		log.Print("No RabbitMQ URL found")
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr != "" {
		cfg.RedisAddr = redisAddr
	} else {
		cfg.RedisAddr = "localhost:6379"
	}

	return &cfg, nil
}
