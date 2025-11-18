package worker

import (
	"context"
	"konnect/internal/cache"
	"konnect/internal/database"
	"konnect/internal/logger"
	"konnect/internal/script"
	"log"

	"github.com/hibiken/asynq"
)

// unique job type for the cache seeding
const (
	TypeSeedCache = "seed:cache"
)

func NewCacheSeedingJob(client *asynq.Client) error {
	task := asynq.NewTask(TypeSeedCache, nil)
	info, err := client.Enqueue(task, asynq.Queue("critical"), asynq.MaxRetry(5))
	if err != nil {
		return err
	}
	log.Printf("enqueued cache seeding job: id=%s queue=%s\n", info.ID, info.Queue)
	return nil
}

// CacheSeederProcessor implements asynq.Handler interface
type CacheSeederProcessor struct {
	db             *database.DB
	interestsCache *cache.InterestCache
	logger         *logger.Logger
}

func (p *CacheSeederProcessor) ProcessTask(ctx context.Context, t *asynq.Task) error {
	// proceed with seeding
	return script.SeedInterestCache(p.db, p.interestsCache, p.logger)
}

func NewCacheSeederProcessor(db *database.DB, interestsCache *cache.InterestCache, logger *logger.Logger) *CacheSeederProcessor {
	return &CacheSeederProcessor{
		db:             db,
		interestsCache: interestsCache,
		logger:         logger,
	}
}
