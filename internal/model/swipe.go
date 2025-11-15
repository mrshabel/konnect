package model

import "github.com/google/uuid"

type SwipeType string

const (
	Like SwipeType = "like"
	Pass SwipeType = "pass"
)

type Swipe struct {
	Model
	SwiperID  uuid.UUID `gorm:"not null" json:"swiperId"`
	SwipeeID  uuid.UUID `gorm:"not null" json:"swipeeId"`
	SwipeType SwipeType `gorm:"type:varchar(10);not null" json:"swipeType"`

	// relations
	Swiper *User `json:"swiper,omitempty"`
	Swipee *User `json:"swipee,omitempty"`
}

type CreateSwipeRequest struct {
	SwipeeID  uuid.UUID `json:"swipeeId" binding:"required"`
	SwipeType SwipeType `json:"swipeType" binding:"required,oneof=like pass"`
}

type SwipeResponse struct {
	Swipe Swipe  `json:"swipe"`
	Match *Match `json:"match,omitempty"`
}
