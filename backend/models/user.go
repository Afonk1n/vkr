package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Username  string         `json:"username" gorm:"uniqueIndex;not null"`
	Email     string         `json:"email" gorm:"uniqueIndex;not null"`
	Password  string         `json:"-" gorm:"not null"` // Password hash, not exposed in JSON
	AvatarPath string        `json:"avatar_path" gorm:"type:text"`
	Bio       string         `json:"bio" gorm:"type:text"`
	SocialLinks string       `json:"social_links" gorm:"type:jsonb"` // JSON: {"vk": "", "telegram": "", "instagram": ""}
	IsAdmin   bool           `json:"is_admin" gorm:"default:false"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Reviews []Review `json:"reviews,omitempty" gorm:"foreignKey:UserID"`
}

// TableName specifies the table name for User
func (User) TableName() string {
	return "users"
}

