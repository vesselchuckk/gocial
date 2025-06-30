package config

import (
	"fmt"
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"log"
	"time"
)

type Config struct {
	ServerHost       string `env:"SERVER_HOST"`
	ServerPort       string `env:"SERVER_PORT"`
	DatabaseName     string `env:"DB_NAME"`
	DatabaseHost     string `env:"DB_HOST"`
	DatabasePort     string `env:"DB_PORT"`
	DatabaseUser     string `env:"DB_USER"`
	DatabasePassword string `env:"DB_PASSWORD"`
	DSN              string `env:"DB_URL"`

	ENV     string
	Version string
	exp     time.Duration `env:"TOKEN_EXP"`

	MailAPI   string `env:"MAIL_API_KEY"`
	FromEmail string `env:"SENDER_EMAIL"`

	FrontendURL string `env:"FRONTEND_URL"`

	AdminUsername string `env:"USERNAME"`
	AdminPassword string `env:"PASSWORD"`

	JWTSecret string `env:"JWT_SECRET"`
	JWTiss    string `env:"JWT_ISSUER"`

	RedisAddr    string `env:"REDIS_ADDR"`
	RedisPW      string `env:"REDIS_PASSWORD"`
	RedisDB      int    `env:"REDIS_DB"`
	RedisEnabled bool   `env:"REDIS_ENABLED"`
}

func New() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("error loading .env: %v", err)
	}

	var cfg Config
	if err = env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("error getting config: %w", err)
	}

	return &cfg, nil
}
