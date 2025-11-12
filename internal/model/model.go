package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Model struct {
	ID        uuid.UUID      `gorm:"primaryKey;not null;type:uuid;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time      `json:"createdAt" gorm:"not null;default:now()"`
	UpdatedAt time.Time      `json:"updatedAt" gorm:"not null;default:now()"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
