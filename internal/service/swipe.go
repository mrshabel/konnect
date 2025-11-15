package service

import (
	"errors"
	"fmt"
	"konnect/internal/database"
	"konnect/internal/logger"
	"konnect/internal/model"
	"konnect/internal/worker"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	ErrAlreadySwiped = errors.New("user has already swiped on this profile")
	ErrSwipeNotFound = errors.New("swipe not found")
	ErrSelfSwipe     = errors.New("user cannot swipe on their own profile")
)

type SwipeService struct {
	db     *database.DB
	worker *asynq.Client
	logger *zap.Logger
}

func NewSwipeService(db *database.DB, worker *asynq.Client, logger *logger.Logger) *SwipeService {
	return &SwipeService{
		db:     db,
		worker: worker,
		logger: logger.With(zap.String("component", "swipe_service")),
	}
}

// CreateSwipe creates a new swipe and checks for a match if the swipe is a 'like'. A match is returned if any
func (s *SwipeService) CreateSwipe(swipe *model.Swipe) (*model.Swipe, *model.Match, error) {
	if swipe.SwiperID == swipe.SwipeeID {
		return nil, nil, ErrSelfSwipe
	}

	var match *model.Match

	// check mutual swipe and create a match
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 1. create swipe
		if err := tx.Create(swipe).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return ErrAlreadySwiped
			}

			s.logError(err, "failed to create swipe",
				zap.String("swiperId", swipe.SwiperID.String()),
				zap.String("swipeeId", swipe.SwipeeID.String()),
			)
			return err
		}

		// 2. check for mutual like
		if swipe.SwipeType != model.Like {
			return nil
		}
		var reverseSwipe model.Swipe
		err := tx.Where("swiper_id = ? AND swipee_id = ? AND swipe_type = ?", swipe.SwipeeID, swipe.SwiperID, model.Like).First(&reverseSwipe).Error
		if err != nil {
			// any error aside record not found is a fatal exception
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				s.logError(err, "failed to check for reverse swipe",
					zap.String("swiperId", swipe.SwipeeID.String()),
					zap.String("swipeeId", swipe.SwiperID.String()),
				)
				return err
			}
			return nil
		}

		// 3. create match
		newMatch := &model.Match{
			User1ID: swipe.SwiperID,
			User2ID: swipe.SwipeeID,
		}
		if err := tx.Create(newMatch).Error; err != nil {
			// match exist
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				s.logger.Warn("match already exists",
					zap.String("user1Id", newMatch.User1ID.String()),
					zap.String("user2Id", newMatch.User2ID.String()),
				)
				return nil
			}
			s.logError(err, "failed to create match",
				zap.String("user1Id", newMatch.User1ID.String()),
				zap.String("user2Id", newMatch.User2ID.String()),
			)
			return err
		}
		match = newMatch
		return nil

	})

	if err != nil {
		return nil, nil, err
	}

	return swipe, match, nil
}

// GetSwipeHistory retrieves a history of swipes made by a given user
func (s *SwipeService) GetSwipeHistory(userID uuid.UUID, limit, offset int) ([]model.Swipe, error) {
	var swipes []model.Swipe

	query := s.db.Where("swiper_id = ?", userID).Joins("Swipee")
	if err := query.Limit(limit).Offset(offset).Find(&swipes).Error; err != nil {
		s.logError(err, "failed to get swipe history")
		return nil, err
	}

	return swipes, nil
}

// GetSwipeByID retrieves the swipe details and associated swiper and swipee
func (s *SwipeService) GetSwipeByID(id uuid.UUID) (*model.Swipe, error) {
	var swipe model.Swipe
	if err := s.db.
		Joins("Swiper").
		Joins("Swipee").
		Take(&swipe, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSwipeNotFound
		}
		return nil, err
	}
	return &swipe, nil
}

func (s *SwipeService) SendMatchNotification(swipe *model.Swipe) error {
	// send message to only the user whose profile was swiped on
	message := fmt.Sprintf("It's a match! You and @%s both liked each other. Start chatting now!", swipe.Swiper.Username)
	err := worker.NewEmailDeliveryJob(s.worker, model.EmailPayload{
		Email:   swipe.Swipee.Email,
		Subject: "New Konnect Match!",
		Message: message,
	})
	if err != nil {
		s.logger.Error("Failed to send notification to swiper", zap.Error(err))
	}
	return err
}

func (s *SwipeService) logError(err error, msg string, fields ...zap.Field) {
	s.logger.Error(msg, append(fields, zap.Error(err))...)
}
