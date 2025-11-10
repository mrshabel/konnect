package database

import (
	"fmt"
	"konnect/internal/config"
	"konnect/internal/logger"
	"konnect/internal/model"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct {
	*gorm.DB
}

func New(cfg *config.Config, logger *logger.Logger) (*DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", cfg.DbHost, cfg.DbUsername, cfg.DbPassword, cfg.DbName, cfg.DbPort)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	logger.Info("database connected successfully")

	// enable postgis extension
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS postgis;").Error; err != nil {
		logger.Error("failed to enable postgis extension", zap.Error(err))
		return nil, err
	}

	// run db migrations
	if err := db.AutoMigrate(
		&model.User{},
		&model.Profile{},
		&model.Swipe{},
		&model.Match{},
		&model.Message{},
	); err != nil {
		logger.Error("failed to run migrations", zap.Error(err))
		return nil, err
	}

	return &DB{db}, nil
}
