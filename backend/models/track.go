package models

import (
	"time"

	"gorm.io/gorm"
)

// Track represents a track in an album
type Track struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	AlbumID    uint           `json:"album_id" gorm:"not null"`
	Title      string         `json:"title" gorm:"not null"`
	Duration   *int           `json:"duration"` // Duration in seconds
	TrackNumber *int          `json:"track_number"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Album Album `json:"album,omitempty" gorm:"foreignKey:AlbumID"`
}

// TableName specifies the table name for Track
func (Track) TableName() string {
	return "tracks"
}

