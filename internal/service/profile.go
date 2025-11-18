package service

import (
	"context"
	"errors"
	"konnect/internal/cache"
	"konnect/internal/database"
	"konnect/internal/logger"
	"konnect/internal/model"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrProfileNotFound     = errors.New("profile not found")
	ErrProfileExists       = errors.New("profile already exists")
	ErrProfileUserNotFound = errors.New("user not found")
)

type ProfileService struct {
	db            *database.DB
	interestCache *cache.InterestCache
	logger        *zap.Logger
}

func NewProfileService(db *database.DB, interestCache *cache.InterestCache, logger *logger.Logger) *ProfileService {
	return &ProfileService{
		db:            db,
		interestCache: interestCache,
		logger:        logger.With(zap.String("component", "profile_service")),
	}
}

// CreateProfile creates a new profile for a user
func (s *ProfileService) CreateProfile(profile *model.Profile) error {
	query := `
		INSERT INTO profiles(user_id, fullname, interests, bio, photo_url, photo_public_id, is_verified, dob, gender, is_gender_public, relationship_intent, latitude, longitude, location)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, ST_Point($14, $15))
		RETURNING id, user_id, fullname, interests, bio, photo_url, photo_public_id, is_verified, dob, gender, is_gender_public, relationship_intent, latitude, longitude,
		created_at, updated_at
	`

	if err := s.db.Raw(query,
		profile.UserID,
		profile.Fullname,
		profile.Interests,
		profile.Bio,
		profile.PhotoURL,
		profile.PhotoPublicID,
		profile.IsVerified,
		profile.DOB,
		profile.Gender,
		profile.IsGenderPublic,
		profile.RelationshipIntent,
		profile.Latitude,
		profile.Longitude,
		profile.Longitude,
		profile.Latitude,
	).Scan(profile).Error; err != nil {
		s.logError(err, "failed to create profile", zap.String("user_id", profile.UserID.String()))
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return ErrProfileExists
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProfileUserNotFound
		}

		return err
	}

	// add interests to cache in background
	go func() {
		if err := s.interestCache.AddUserInterests(context.Background(), profile.UserID.String(), profile.Interests); err != nil {
			s.logger.Warn("failed to add user interests to cache on profile creation", zap.Error(err), zap.String("user_id", profile.UserID.String()))
		}
	}()
	return nil
}

// GetProfile retrieves a profile by ID
func (s *ProfileService) GetProfile(id uuid.UUID) (*model.Profile, error) {
	var profile model.Profile
	if err := s.db.Where("id = ?", id).Take(&profile); err != nil {
		return nil, ErrProfileNotFound
	}
	return &profile, nil
}

// GetProfileByUserID retrieves a profile by user ID
func (s *ProfileService) GetProfileByUserID(userID uuid.UUID) (*model.Profile, error) {
	var profile model.Profile
	if err := s.db.Where("user_id = ?", userID).Take(&profile).Error; err != nil {
		return nil, ErrProfileNotFound
	}
	return &profile, nil
}

// UpdateProfile updates an existing profile
func (s *ProfileService) UpdateProfileByUserID(userID uuid.UUID, updates *model.Profile) (*model.Profile, error) {
	// retrieve current profile state
	currentProfile, err := s.GetProfileByUserID(userID)
	if err != nil {
		return nil, err
	}

	var profile model.Profile
	if err := s.db.Model(&profile).
		Where("user_id = ?", userID).
		Clauses(clause.Returning{}).
		Updates(updates).
		Scan(&profile).Error; err != nil {
		s.logError(err, "failed to update profile", zap.String("user_id", userID.String()))

		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrProfileExists
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProfileUserNotFound
		}
		return nil, err
	}

	// sync with cache if interests were provided
	if updates.Interests != nil {
		go func() {
			err := s.interestCache.UpdateUserInterests(context.Background(), userID.String(), currentProfile.Interests, updates.Interests)
			if err != nil {
				s.logger.Warn("failed to update user interests in cache on profile update",
					zap.Error(err),
					zap.String("user_id", userID.String()),
				)
			}
		}()
	}

	return &profile, nil
}

// GetNearbyProfiles gets profiles within a radius (in meters) of given coordinates
func (s *ProfileService) GetNearbyProfiles(userID uuid.UUID, lat, lng float64, radiusMeters float64, offset int, limit int) ([]model.Profile, error) {
	var profiles []model.Profile

	// nearby distance relative to the location. (lon, lat)
	query := s.db.Where("user_id != ?", userID).
		Where("ST_DWithin(location, ST_Point(?, ?)::GEOGRAPHY, ?)", lng, lat, radiusMeters)
	if err := query.Limit(limit).Offset(offset).Find(&profiles).Error; err != nil {
		s.logError(err, "failed to get nearby profiles")
		return nil, err
	}

	return profiles, nil
}

func (s *ProfileService) logError(err error, msg string, fields ...zap.Field) {
	s.logger.Error(msg, append(fields, zap.Error(err))...)
}
