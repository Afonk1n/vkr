package models

import (
	"gorm.io/gorm"
)

// TrackGenre represents the many-to-many relationship between tracks and genres
type TrackGenre struct {
	ID      uint `json:"id" gorm:"primaryKey"`
	TrackID uint `json:"track_id" gorm:"not null;index"`
	GenreID uint `json:"genre_id" gorm:"not null;index"`

	// Relationships
	Track Track `json:"track,omitempty" gorm:"foreignKey:TrackID"`
	Genre Genre `json:"genre,omitempty" gorm:"foreignKey:GenreID"`
}

// TableName specifies the table name for TrackGenre
func (TrackGenre) TableName() string {
	return "track_genres"
}

// BeforeCreate ensures unique track-genre combination
func (tg *TrackGenre) BeforeCreate(tx *gorm.DB) error {
	var count int64
	tx.Model(&TrackGenre{}).
		Where("track_id = ? AND genre_id = ?", tg.TrackID, tg.GenreID).
		Count(&count)

	if count > 0 {
		return gorm.ErrDuplicatedKey
	}
	return nil
}

