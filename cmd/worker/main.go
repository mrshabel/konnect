package main

import (
	"konnect/internal/cache"
	"konnect/internal/config"
	"konnect/internal/database"
	"konnect/internal/logger"
	"konnect/internal/service"
	"konnect/internal/worker"
	"log"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

func main() {
	// initialize logger
	logger, err := logger.NewLogger()
	if err != nil {
		// standard log as fallback
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer logger.Close()

	// load configs
	cfg, err := config.New()
	if err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}

	// db
	db, err := database.New(cfg, logger)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.String("component", "main"), zap.Error(err))
	}

	// cache
	cacheClient, err := cache.New(cfg)
	if err != nil {
		logger.Fatal("failed to connect to redis cache", zap.Error(err))
	}
	defer cacheClient.Close()

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: cfg.RedisAddr, Password: cfg.RedisPassword},
		asynq.Config{
			// concurrent workers to use
			Concurrency: 10,
			// queues with priorities
			Queues: worker.Queues,
		},
	)

	// handlers and services
	emailService := service.NewEmailService(cfg)

	// cache services
	interestCache := cache.NewInterests(cacheClient, logger)

	emailProcessor := worker.NewEmailProcessor(emailService)
	cacheSeederProcessor := worker.NewCacheSeederProcessor(db, interestCache, logger)
	// smsProcessor := worker.NewSMSProcessor()

	// mux maps a type to a handler
	mux := asynq.NewServeMux()
	mux.Handle(worker.TypeEmailDelivery, emailProcessor)
	mux.Handle(worker.TypeSeedCache, cacheSeederProcessor)
	// mux.Handle(worker.TypeSMSDelivery, smsProcessor)

	if err := srv.Run(mux); err != nil {
		logger.Fatal("could not run server", zap.Error(err))
	}
}
