package cache

import (
	"context"
	"konnect/internal/database"
	"konnect/internal/logger"
	"konnect/internal/model"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type InterestCache struct {
	client *Client
	logger *zap.Logger
}

func NewInterests(client *Client, logger *logger.Logger) *InterestCache {
	return &InterestCache{
		client: client,
		logger: logger.With(zap.String("component", "interest_cache")),
	}
}

func (i *InterestCache) GetUserInterests(ctx context.Context, userID string) ([]string, error) {
	return i.client.SMembers(ctx, GetUserInterestsKey(userID)).Result()
}

func (i *InterestCache) AddUserInterests(ctx context.Context, userID string, interests []string) error {
	// start in redis transaction
	tx := i.client.TxPipeline()

	// record interest in user bucket then user in interest bucket
	for _, interest := range interests {
		tx.SAdd(ctx, GetUserInterestsKey(userID), interest)
		tx.SAdd(ctx, GetInterestBucketKey(interest), userID)
	}

	if _, err := tx.Exec(ctx); err != nil {
		return err
	}
	i.logger.Info("User interests added successfully")
	return nil
}

func (i *InterestCache) RemoveUserInterests(ctx context.Context, userID string, interests []string) error {
	// start in redis transaction
	tx := i.client.TxPipeline()

	for _, interest := range interests {
		tx.SRem(ctx, GetUserInterestsKey(userID), interest)
		tx.SRem(ctx, GetInterestBucketKey(interest), userID)
	}

	if _, err := tx.Exec(ctx); err != nil {
		return err
	}
	i.logger.Info("User interests removed successfully")
	return nil
}

// UpdateUserInterests removes all old user interests and replaces them with new ones
func (i *InterestCache) UpdateUserInterests(ctx context.Context, userID string, oldInterests, newInterests []string) error {
	// start in redis transaction
	tx := i.client.TxPipeline()

	// clear user interests bucket
	userKey := GetUserInterestsKey(userID)
	tx.Del(ctx, userKey)

	for _, interest := range oldInterests {
		tx.SRem(ctx, GetInterestBucketKey(interest), userID)
	}
	for _, interest := range newInterests {
		tx.SAdd(ctx, GetInterestBucketKey(interest), userID)
		tx.SAdd(ctx, userKey, interest)
	}

	if _, err := tx.Exec(ctx); err != nil {
		return err
	}
	i.logger.Info("User interests updated successfully")
	return nil
}

// GetCommonInterests retrieves the common interests among two users
func (i *InterestCache) GetCommonInterests(ctx context.Context, userID1, userID2 string) ([]string, error) {
	// get set intersection from user buckets
	return i.client.SInter(ctx, GetUserInterestsKey(userID1), GetUserInterestsKey(userID2)).Result()
}

// GetUsersWithMatchingInterests retrieves all user ids with matching interests
func (i *InterestCache) GetUsersWithMatchingInterests(ctx context.Context, userID string, interests []string, limit int) ([]string, error) {
	if len(interests) == 0 {
		return []string{}, nil
	}

	if limit <= 0 {
		limit = 100
	}

	// get interest bucket keys
	keys := make([]string, 0, len(interests))
	for _, interest := range interests {
		keys = append(keys, GetInterestBucketKey(interest))
	}

	// load all members while filtering out requested user
	users, err := i.client.SUnion(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	usersOut := make([]string, 0, len(users))
	for _, user := range users {
		if user != userID {
			usersOut = append(usersOut, user)
		}
	}

	return usersOut, nil
}

// GetMultipleUserInterests fetches interests for multiple users at once. their id to interests bucket mapping is returned
func (i *InterestCache) GetMultipleUserInterests(ctx context.Context, ids []string) (map[string][]string, error) {
	pipe := i.client.Pipeline()

	// load the commands to get user interests in the pipeline
	cmds := make(map[string]*redis.StringSliceCmd)
	for _, id := range ids {
		cmds[id] = pipe.SMembers(ctx, GetUserInterestsKey(id))
	}

	// execute all pipelined commands
	if _, err := pipe.Exec(ctx); err != nil {
		return nil, err
	}

	// get results
	results := make(map[string][]string)
	for id, cmd := range cmds {
		interests, err := cmd.Result()
		// skip unavailable user interests
		if err != nil {
			continue
		}
		results[id] = interests
	}

	return results, nil
}

// GetBucketMembers returns all userIDs in an interest bucket
func (i *InterestCache) GetInterestMembers(ctx context.Context, interest string) ([]string, error) {
	return i.client.SMembers(ctx, GetInterestBucketKey(interest)).Result()
}

func (i *InterestCache) GetInterestMembersCount(ctx context.Context, interest string) (int64, error) {
	return i.client.SCard(ctx, GetInterestBucketKey(interest)).Result()
}

// SeedActiveProfiles seeds interest buckets from profiles that were active within the last n days
func (i *InterestCache) SeedActiveProfiles(ctx context.Context, db *database.DB, interests []string, days int) error {
	// get least allowed date
	allowedTime := time.Now().AddDate(0, 0, -days)

	// fetch profiles in batches of 100
	var profiles []model.Profile
	err := db.Joins("JOIN users ON users.id = profiles.user_id").
		Where("users.last_active >= ?", allowedTime).
		Select("profiles.user_id", "profiles.interests").
		FindInBatches(&profiles, 100, func(tx *gorm.DB, batch int) error {

			pipe := i.client.Pipeline()
			for _, p := range profiles {
				userID := p.UserID.String()
				userInterestsKey := GetUserInterestsKey(userID)

				// clear old interests for the user
				pipe.Del(ctx, userInterestsKey)

				// add to interest and user bucket
				for _, interest := range p.Interests {
					pipe.SAdd(ctx, userInterestsKey, interest)
					pipe.SAdd(ctx, GetInterestBucketKey(interest), userID)
				}
			}

			// execute pipeline for current batch
			if _, err := pipe.Exec(ctx); err != nil {
				i.logger.Warn("Seeding active profile failed. Skipping remaining batches...", zap.Error(err))
				// skip remaining batches
				return err
			}

			return nil
		}).Error

	if err != nil {
		i.logger.Warn("Seeding active profiles failed", zap.Error(err))
		return err
	}

	i.logger.Info("Seeded active profiles into redis", zap.Int("batch_size", 100), zap.Int("days", days))
	return nil
}

func GetInterestBucketKey(interest string) string {
	return "interests:" + interest
}

func GetUserInterestsKey(userID string) string {
	return "interests:user:" + userID
}

func GetUserFeedKey(userID string) string {
	return "feeds:" + userID
}

func GetUserSwipesKey(userID string) string {
	return "swipes:" + userID
}
