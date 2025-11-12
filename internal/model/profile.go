package model

import (
	"database/sql/driver"
	"encoding/json"
	"konnect/internal/util"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Gender string
type RelationshipIntent string

const (
	Male   Gender = "male"
	Female Gender = "female"

	Friendship RelationshipIntent = "friendship"
	Dating     RelationshipIntent = "dating"
	Casual     RelationshipIntent = "casual"
	Marriage   RelationshipIntent = "marriage"
)

// Interests is a custom type for handling postgres JSONB
type Interests []string

// Scan implements sql.Scanner interface for gorm capatibility
func (i *Interests) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, i)
}

// Value implements driver.Valuer interface for gorm compatibility
func (i Interests) Value() (driver.Value, error) {
	if len(i) == 0 {
		return nil, nil
	}
	return json.Marshal(i)
}

type Profile struct {
	Model
	UserID             uuid.UUID          `gorm:"not null;uniqueIndex" json:"userId"`
	Fullname           string             `gorm:"type:varchar(255);not null" json:"fullname"`
	Interests          Interests          `gorm:"type:jsonb;not null" json:"interests"`
	Bio                string             `gorm:"type:varchar(5000);not null" json:"bio"`
	PhotoURL           *string            `gorm:"type:varchar(500)" json:"photoUrl"`
	PhotoPublicID      *string            `gorm:"type:varchar(255)" json:"photoPublicId"`
	IsVerified         bool               `gorm:"not null;default:false" json:"isVerified"`
	DOB                time.Time          `gorm:"type:date;not null;check:dob < NOW()" json:"dob"`
	Gender             Gender             `gorm:"type:varchar(10);not null" json:"gender"`
	IsGenderPublic     bool               `gorm:"not null;default:true" json:"isGenderPublic"`
	RelationshipIntent RelationshipIntent `gorm:"type:varchar(100);not null" json:"relationshipIntent"`
	Latitude           float64            `gorm:"type:decimal(9,6);not null" json:"latitude"`
	Longitude          float64            `gorm:"type:decimal(9,6);not null" json:"longitude"`
	// postgis point. it is used internally for querying
	Location string `gorm:"type:GEOGRAPHY(POINT);not null;index:idx_profiles_location,type:gist" json:"-"`

	// relations
	User *User `json:"user,omitempty"`
}

type CreateProfileRequest struct {
	Fullname           string             `json:"fullname" binding:"required,min=2,max=255"`
	Interests          Interests          `json:"interests" binding:"required,min=1"`
	Bio                string             `json:"bio" binding:"required,min=10,max=5000"`
	DOB                string             `json:"dob" binding:"required"`
	Gender             Gender             `json:"gender" binding:"required,oneof=male female"`
	IsGenderPublic     bool               `json:"isGenderPublic"`
	RelationshipIntent RelationshipIntent `json:"relationshipIntent" binding:"required,oneof=friendship dating casual marriage"`
	Latitude           float64            `json:"latitude" binding:"required,latitude"`
	Longitude          float64            `json:"longitude" binding:"required,longitude"`
}

type UpdateProfileRequest struct {
	Fullname           *string             `json:"fullname,omitempty" binding:"omitempty,min=2,max=255"`
	Interests          Interests           `json:"interests,omitempty" binding:"omitempty,min=1"`
	Bio                *string             `json:"bio,omitempty" binding:"omitempty,min=10,max=5000"`
	DOB                *string             `json:"dob,omitempty" binding:"omitempty"`
	Gender             *Gender             `json:"gender,omitempty" binding:"omitempty,oneof=male female"`
	IsGenderPublic     *bool               `json:"isGenderPublic,omitempty"`
	RelationshipIntent *RelationshipIntent `json:"relationshipIntent,omitempty" binding:"omitempty,oneof=friendship dating casual marriage"`
	Latitude           *float64            `json:"latitude,omitempty" binding:"omitempty,latitude"`
	Longitude          *float64            `json:"longitude,omitempty" binding:"omitempty,longitude"`
}

// Compose converts the update model into the main model with non-zero fields
func (u *UpdateProfileRequest) Compose() (*Profile, error) {
	p := &Profile{}

	if u.Fullname != nil {
		p.Fullname = *u.Fullname
	}
	if len(u.Interests) > 0 {
		p.Interests = u.Interests
	}
	if u.Bio != nil {
		p.Bio = *u.Bio
	}
	if u.DOB != nil {
		t, err := util.ValidateDate(*u.DOB)
		if err != nil {
			return nil, err
		}
		p.DOB = t
	}
	if u.Gender != nil {
		p.Gender = *u.Gender
	}
	if u.IsGenderPublic != nil {
		p.IsGenderPublic = *u.IsGenderPublic
	}
	if u.RelationshipIntent != nil {
		p.RelationshipIntent = *u.RelationshipIntent
	}
	if u.Latitude != nil {
		p.Latitude = *u.Latitude
	}
	if u.Longitude != nil {
		p.Longitude = *u.Longitude
	}
	if u.Latitude != nil && u.Longitude != nil {
		p.Location = gorm.Expr("ST_Point(?, ?)", u.Longitude, u.Latitude).SQL
	}

	return p, nil
}

type GetNearbyProfilesRequest struct {
	Lat    float64 `form:"lat" binding:"required,latitude"`
	Lng    float64 `form:"lng" binding:"required,longitude"`
	Radius float64 `form:"radius,default=5000" binding:"min=100,max=50000"`
	Limit  int     `form:"limit,default=20" binding:"min=1,max=100"`
	Offset int     `form:"offset,default=0" binding:"min=0"`
}
