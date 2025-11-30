package models

import (
	"time"

	"gorm.io/gorm"
)

// Track represents a track in an album
type Track struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	AlbumID        uint           `json:"album_id" gorm:"not null"`
	Title          string         `json:"title" gorm:"not null"`
	Duration       *int           `json:"duration"` // Duration in seconds
	TrackNumber    *int           `json:"track_number"`
	CoverImagePath string         `json:"cover_image_path"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Album   Album       `json:"album,omitempty" gorm:"foreignKey:AlbumID"`
	Likes   []TrackLike `json:"likes,omitempty" gorm:"foreignKey:TrackID"`
	Genres  []Genre     `json:"genres,omitempty" gorm:"many2many:track_genres;"`
	Reviews []Review    `json:"reviews,omitempty" gorm:"foreignKey:TrackID"`
}

// TableName specifies the table name for Track
func (Track) TableName() string {
	return "tracks"
}
