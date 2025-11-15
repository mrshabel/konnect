package main

import (
	"konnect/internal/config"
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

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: cfg.RedisAddr},
		asynq.Config{
			// concurrent workers to use
			Concurrency: 10,
			// queues with priorities
			Queues: worker.Queues,
		},
	)

	// handlers and services
	emailService := service.NewEmailService(cfg)

	emailProcessor := worker.NewEmailProcessor(emailService)
	// smsProcessor := worker.NewSMSProcessor()

	// mux maps a type to a handler
	mux := asynq.NewServeMux()
	mux.Handle(worker.TypeEmailDelivery, emailProcessor)
	// mux.Handle(worker.TypeSMSDelivery, smsProcessor)

	if err := srv.Run(mux); err != nil {
		logger.Fatal("could not run server", zap.Error(err))
	}
}
