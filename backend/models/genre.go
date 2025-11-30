package models

import (
	"time"

	"gorm.io/gorm"
)

// Genre represents a music genre
type Genre struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"uniqueIndex;not null"`
	Description string         `json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Albums []Album `json:"albums,omitempty" gorm:"foreignKey:GenreID"`
	Tracks []Track `json:"tracks,omitempty" gorm:"many2many:track_genres;"`
}

// TableName specifies the table name for Genre
func (Genre) TableName() string {
	return "genres"
}

