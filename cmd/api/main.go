package main

import (
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"github.com/vesselchuckk/go-social/cmd/api/config"
	"github.com/vesselchuckk/go-social/cmd/api/server"
	"github.com/vesselchuckk/go-social/internal/auth"
	"github.com/vesselchuckk/go-social/internal/mails"
	"github.com/vesselchuckk/go-social/internal/store"
	"github.com/vesselchuckk/go-social/internal/store/cache"
	"go.uber.org/zap"
	"log"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Panicf("warning: error occured when loading config: %s", err)
	}

	cfg, err := config.New()
	if err != nil {
		log.Fatalf("failed to load config: %s", err)
	}

	// logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()

	db, err := store.NewPostgresDB()
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()
	logger.Info("database connection established")

	var redisDB *redis.Client
	if cfg.RedisEnabled {
		redisDB = cache.NewRedisClient(
			cfg.RedisAddr,
			cfg.RedisPW,
			cfg.RedisDB,
		)
		logger.Info("Redis caching is enabled")
	}

	dataStorage := store.NewStorage(db)

	mailer := mails.NewMailer(cfg.MailAPI, cfg.FromEmail)

	JWTauth := auth.NewJWTAuth(cfg.JWTSecret, cfg.JWTiss, cfg.JWTiss)

	srv := server.NewServer(cfg, dataStorage, logger, mailer, JWTauth, redisDB)
	if err := srv.Run(); err != nil {
		logger.Fatal(err)
	}
}
