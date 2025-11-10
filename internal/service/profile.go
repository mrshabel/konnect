package service

import (
	"errors"
	"fmt"
	"konnect/internal/database"
	"konnect/internal/logger"
	"konnect/internal/model"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm/clause"
)

var (
	ErrProfileNotFound = errors.New("profile not found")
	ErrProfileExists   = errors.New("profile already exists")
)

type ProfileService struct {
	db     *database.DB
	logger *zap.Logger
}

func NewProfileService(db *database.DB, logger *logger.Logger) *ProfileService {
	return &ProfileService{
		db:     db,
		logger: logger.With(zap.String("component", "profile_service")),
	}
}

// CreateProfile creates a new profile for a user
func (s *ProfileService) CreateProfile(profile *model.Profile) error {
	// check for profile existence
	existingProfile, _ := s.GetProfileByUserID(profile.UserID)
	if existingProfile != nil {
		s.logError(ErrProfileExists, "profile already exists", zap.String("user_id", profile.UserID.String()))
		return ErrProfileExists
	}

	// set db point from coordinates
	profile.Location = fmt.Sprintf("ST_Point(%f, %f)", profile.Longitude, profile.Latitude)

	if err := s.db.Create(profile).Error; err != nil {
		s.logError(err, "failed to create profile", zap.String("user_id", profile.UserID.String()))
		return err
	}
	return nil
}

// GetProfile retrieves a profile by ID
func (s *ProfileService) GetProfile(id uuid.UUID) (*model.Profile, error) {
	var profile model.Profile
	if err := s.db.First(&profile, "id = ?", id).Error; err != nil {
		return nil, ErrProfileNotFound
	}
	return &profile, nil
}

// GetProfileByUserID retrieves a profile by user ID
func (s *ProfileService) GetProfileByUserID(userID uuid.UUID) (*model.Profile, error) {
	var profile model.Profile
	if err := s.db.First(&profile, "user_id = ?", userID).Error; err != nil {
		return nil, ErrProfileNotFound
	}
	return &profile, nil
}

// UpdateProfile updates an existing profile
func (s *ProfileService) UpdateProfile(id uuid.UUID, updates *model.Profile) (*model.Profile, error) {
	// set db point if coordinates are updated
	if updates.Latitude != 0 && updates.Longitude != 0 {
		updates.Location = fmt.Sprintf("ST_POINT(%f, %f)", updates.Longitude, updates.Latitude)
	}

	var profile model.Profile
	if err := s.db.Model(&model.Profile{}).
		Where("id = ?", id).
		Clauses(clause.Returning{}).
		Updates(updates).
		Scan(&profile).Error; err != nil {
		s.logError(err, "failed to update profile", zap.String("profile_id", id.String()))
		return nil, err
	}
	return &profile, nil
}

// GetNearbyProfiles gets profiles within a radius (in meters) of given coordinates
func (s *ProfileService) GetNearbyProfiles(lat, lng float64, radiusMeters float64, offset int, limit int) ([]model.Profile, error) {
	var profiles []model.Profile

	// nearby distance relative to the location. (lon, lat)
	query := s.db.Where("ST_DWithin(location, ST_Point(?, ?)::GEOGRAPHY, ?)", lng, lat, radiusMeters)
	if err := query.Limit(limit).Offset(offset).Find(&profiles).Error; err != nil {
		s.logError(err, "failed to get nearby profiles")
		return nil, err
	}

	return profiles, nil
}

func (s *ProfileService) logError(err error, msg string, fields ...zap.Field) {
	s.logger.Error(msg, append(fields, zap.Error(err))...)
}
