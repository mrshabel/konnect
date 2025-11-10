package model

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	AppUser UserRole = "user"
	Admin   UserRole = "admin"
)

type OAuthProvider string

const (
	Google OAuthProvider = "google"
)

type User struct {
	Model
	Email      string     `gorm:"uniqueIndex;not null" json:"email"`
	Username   string     `gorm:"uniqueIndex;not null" json:"username"`
	Provider   string     `gorm:"not null" json:"provider"`
	Role       UserRole   `gorm:"type:varchar(100);default:'user'" json:"role"`
	LastActive *time.Time `json:"lastActive"`

	// relations
	Profile *Profile `json:"profile,omitempty"`
}

type AuthenticatedUser struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Role     UserRole  `json:"role"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
