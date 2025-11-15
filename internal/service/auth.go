package service

import (
	"errors"
	"konnect/internal/config"
	"konnect/internal/database"
	"konnect/internal/logger"
	"konnect/internal/model"
	"konnect/internal/util"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/markbates/goth"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// errors
var (
	ErrExpiredToken = errors.New("token has expired")
	ErrInvalidToken = errors.New("invalid token")
	ErrUserNotFound = errors.New("user not found")
)

type AuthService struct {
	db     *database.DB
	cfg    *config.Config
	logger *zap.Logger
}

func NewAuthService(db *database.DB, cfg *config.Config, logger *logger.Logger) *AuthService {
	return &AuthService{
		db:     db,
		cfg:    cfg,
		logger: logger.With(zap.String("component", "auth_service")),
	}
}

// UpsertUserFromProvider finds a user by their email or creates a new one
func (s *AuthService) UpsertUserFromProvider(gothUser goth.User) (*model.User, error) {
	now := time.Now()
	username := util.GenerateRandomUsername()

	user := &model.User{
		Email:      gothUser.Email,
		Username:   username,
		Role:       "user",
		Provider:   gothUser.Provider,
		LastActive: &now,
	}

	// upsert user by email
	if err := s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "email"}},
		DoUpdates: clause.AssignmentColumns([]string{"last_active"}),
	}).Create(&user).Error; err != nil {
		s.logError(err, "failed to upsert user", zap.String("email", gothUser.Email), zap.String("provider", gothUser.Provider))
		return nil, err
	}

	return user, nil
}

// GetUserByUsername retrieves user by ID
func (s *AuthService) GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	if err := s.db.First(&user, "username = ?", username).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByID retrieves user by ID
func (s *AuthService) GetUserByID(id uuid.UUID) (*model.User, error) {
	var user model.User
	if err := s.db.Where("id = ?", id).Joins("Profile").Take(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

// UpdateLastActive updates user's last active timestamp
func (s *AuthService) UpdateLastActive(userID string) error {
	now := time.Now()
	return s.db.Model(&model.User{}).Where("id = ?", userID).Update("last_active", now).Error
}

// token helpers (generate and validate)

func (s *AuthService) GenerateAccessToken(user *model.User) (string, error) {
	now := time.Now()
	expiry := now.Add(time.Duration(s.cfg.JWTExpiryMinutes) * time.Minute)
	isVerified := false
	if user.Profile != nil {
		isVerified = user.Profile.IsVerified
	}

	// create token with claims
	claims := jwt.MapClaims{
		"sub":        user.ID.String(),
		"username":   user.Username,
		"role":       user.Role,
		"isVerified": isVerified,
		"iat":        now.Unix(),
		"exp":        expiry.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// sign token with secret key
	tokenString, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		s.logError(err, "failed to generate token", zap.String("user_id", user.ID.String()))
		return "", err
	}

	return tokenString, nil
}

func (s *AuthService) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	// parse token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(s.cfg.JWTSecret), nil
	})

	if err != nil {
		// token expiry
		if err == jwt.ErrTokenExpired {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// logger helpers
func (s *AuthService) logError(err error, msg string, fields ...zap.Field) {
	s.logger.Error(msg, append(fields, zap.Error(err))...)
}

func (s *AuthService) logInfo(msg string, fields ...zap.Field) {
	s.logger.Info(msg, fields...)
}
