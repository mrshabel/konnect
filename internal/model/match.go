package model

import (
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Match struct {
	// id is be a sorted concatenation of user1_id and user2_id
	ID        string         `gorm:"primaryKey;type:varchar(255)" json:"id"`
	User1ID   uuid.UUID      `gorm:"not null" json:"user1Id"`
	User2ID   uuid.UUID      `gorm:"not null" json:"user2Id"`
	IsActive  bool           `gorm:"not null;default:true" json:"isActive"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// relations
	User1    *User     `json:"user1,omitempty"`
	User2    *User     `json:"user2,omitempty"`
	Messages []Message `json:"messages,omitempty"`
}

// BeforeCreate GORM hook to generate the match ID from user IDs. The IDs are sorted and separated with an underscore
func (m *Match) BeforeCreate(tx *gorm.DB) error {
	ids := []string{m.User1ID.String(), m.User2ID.String()}
	sort.Strings(ids)

	m.ID = strings.Join(ids, "_")
	return nil
}

type CreateMatchRequest struct {
	UserID uuid.UUID `json:"userId" binding:"required"`
}

type UpdateMatchRequest struct {
	IsActive bool `json:"isActive"`
}
