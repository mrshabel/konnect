package script

import (
	"context"
	"konnect/internal/cache"
	"konnect/internal/database"
	"konnect/internal/logger"
	"konnect/internal/model"
	"time"

	"go.uber.org/zap"
)

func SeedInterestCache(db *database.DB, interestCache *cache.InterestCache, logger *logger.Logger) error {
	logger.Info("Starting cache seeding process...")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// seed interests for users active in the 30 days
	if err := interestCache.SeedActiveProfiles(ctx, db, model.SystemInterests, 30); err != nil {
		logger.Warn("Failed to seed active user interests", zap.Error(err))
		return err
	}

	logger.Info("Cache seeding process complete")
	return nil
}
