package models

import (
	"time"

	"gorm.io/gorm"
)

// TrackLike represents a like on a track
type TrackLike struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UserID    uint           `json:"user_id" gorm:"not null"`
	TrackID   uint           `json:"track_id" gorm:"not null"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	User  User  `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Track Track `json:"track,omitempty" gorm:"foreignKey:TrackID"`
}

// TableName specifies the table name for TrackLike
func (TrackLike) TableName() string {
	return "track_likes"
}

// BeforeCreate ensures unique like per user per track
func (tl *TrackLike) BeforeCreate(tx *gorm.DB) error {
	var count int64
	tx.Model(&TrackLike{}).
		Where("user_id = ? AND track_id = ?", tl.UserID, tl.TrackID).
		Count(&count)
	
	if count > 0 {
		return gorm.ErrDuplicatedKey
	}
	return nil
}

