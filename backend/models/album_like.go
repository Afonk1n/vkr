package models

import (
	"time"

	"gorm.io/gorm"
)

// AlbumLike represents a like on an album
type AlbumLike struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UserID    uint           `json:"user_id" gorm:"not null"`
	AlbumID   uint           `json:"album_id" gorm:"not null"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	User  User  `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Album Album `json:"album,omitempty" gorm:"foreignKey:AlbumID"`
}

// TableName specifies the table name for AlbumLike
func (AlbumLike) TableName() string {
	return "album_likes"
}

// BeforeCreate ensures unique like per user per album
func (al *AlbumLike) BeforeCreate(tx *gorm.DB) error {
	var count int64
	tx.Model(&AlbumLike{}).
		Where("user_id = ? AND album_id = ?", al.UserID, al.AlbumID).
		Count(&count)
	
	if count > 0 {
		return gorm.ErrDuplicatedKey
	}
	return nil
}

