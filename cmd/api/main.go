package main

import (
	"context"
	"fmt"
	"konnect/internal/cache"
	"konnect/internal/config"
	"konnect/internal/database"
	"konnect/internal/handler"
	"konnect/internal/logger"
	"konnect/internal/router"
	"konnect/internal/service"
	"konnect/internal/worker"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
	"go.uber.org/zap"
)

// @title           Konnect API
// @version         1.0
// @description     Match-making platform for all personalities

// @BasePath  /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {

	// initialize logger
	logger, err := logger.NewLogger()
	if err != nil {
		// standard log as fallback
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer logger.Close()

	// get config
	cfg, err := config.New()
	if err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}

	// db
	db, err := database.New(cfg, logger)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.String("component", "main"), zap.Error(err))
	}

	// worker client
	workerClient := worker.NewWorkerClient(cfg)
	defer workerClient.Close()

	// cache
	cacheClient, err := cache.New(cfg)
	if err != nil {
		logger.Fatal("failed to connect to redis cache", zap.Error(err))
	}
	defer cacheClient.Close()

	// setup goth
	goth.UseProviders(
		google.New(cfg.GoogleClientID, cfg.GoogleClientSecret, cfg.GoogleCallbackURL, "email", "profile"),
	)
	// cloudinary
	cloudinaryService, err := service.NewCloudinaryService(logger)
	if err != nil {
		logger.Fatal("failed to initialize cloudinary service", zap.String("component", "main"), zap.Error(err))
	}

	// cache services
	interestCache := cache.NewInterests(cacheClient, logger)

	// services
	authService := service.NewAuthService(db, cfg, logger)
	// profile service now depends on the interest cache
	profileService := service.NewProfileService(db, interestCache, logger)
	swipeService := service.NewSwipeService(db, workerClient.Client, logger)

	// handlers
	authHandler := handler.NewAuthHandler(authService)
	profileHandler := handler.NewProfileHandler(profileService, cloudinaryService, logger)
	swipeHandler := handler.NewSwipeHandler(swipeService, logger)

	// middleware
	middleware := handler.NewMiddleware(authService, logger)

	// enqueue seeder jobs
	if err := worker.NewCacheSeedingJob(workerClient.Client); err != nil {
		logger.Fatal("failed to enqueue cache seeding job", zap.Error(err))
	}

	// server router
	r := gin.Default()

	router.RegisterRoutes(r, middleware, authHandler, profileHandler, swipeHandler)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%v", cfg.Port),
		Handler:      r,
		IdleTimeout:  time.Minute,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	done := make(chan struct{}, 1)
	go gracefulShutdown(server, logger, done)

	logger.Info("Server starting", zap.String("addr", server.Addr))
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("http server error", zap.Error(err))
	}

	// wait for the graceful shutdown to complete
	<-done
	logger.Info("Graceful shutdown complete")
}

func gracefulShutdown(apiServer *http.Server, logger *logger.Logger, done chan struct{}) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	log.Println("shutting down gracefully, press Ctrl+C again to force")
	stop()
	logger.Info("shutting down gracefully, press Ctrl+C again to force")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := apiServer.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exiting")
	// notify the main goroutine that the shutdown is complete
	done <- struct{}{}
}
