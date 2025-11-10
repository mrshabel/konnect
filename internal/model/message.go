package model

import "github.com/google/uuid"

type Message struct {
	Model
	MatchID  uuid.UUID `gorm:"not null" json:"matchId"`
	SenderID uuid.UUID `gorm:"not null" json:"senderId"`
	Content  string    `gorm:"type:varchar(5000);not null" json:"content"`

	// relations
	Match  *Match `json:"match,omitempty"`
	Sender *User  `json:"sender,omitempty"`
}

type CreateMessageRequest struct {
	Content string `json:"content" binding:"required,min=1,max=5000"`
}
