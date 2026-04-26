package models

import (
	"time"

	"gorm.io/gorm"
)

// Album represents a music album
type Album struct {
	ID                          uint           `json:"id" gorm:"primaryKey"`
	Title                       string         `json:"title" gorm:"not null"`
	Artist                      string         `json:"artist" gorm:"not null"`
	GenreID                     uint           `json:"genre_id" gorm:"not null"`
	CoverImagePath              string         `json:"cover_image_path"`
	ReleaseDate                 *time.Time     `json:"release_date"`
	Description                 string         `json:"description" gorm:"type:text"`
	AverageRating               float64        `json:"average_rating" gorm:"default:0"`
	AverageRatingRhymes         float64        `json:"average_rating_rhymes,omitempty" gorm:"-"`
	AverageRatingStructure      float64        `json:"average_rating_structure,omitempty" gorm:"-"`
	AverageRatingImplementation float64        `json:"average_rating_implementation,omitempty" gorm:"-"`
	AverageRatingIndividuality  float64        `json:"average_rating_individuality,omitempty" gorm:"-"`
	AverageAtmosphereRating     float64        `json:"average_atmosphere_rating,omitempty" gorm:"-"`
	ApprovedReviewsCount        int64          `json:"approved_reviews_count,omitempty" gorm:"-"`
	CreatedAt                   time.Time      `json:"created_at"`
	UpdatedAt                   time.Time      `json:"updated_at"`
	DeletedAt                   gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Genre   Genre       `json:"genre,omitempty" gorm:"foreignKey:GenreID"`
	Tracks  []Track     `json:"tracks,omitempty" gorm:"foreignKey:AlbumID"`
	Reviews []Review    `json:"reviews,omitempty" gorm:"foreignKey:AlbumID"`
	Likes   []AlbumLike `json:"likes,omitempty" gorm:"foreignKey:AlbumID"`
}

// TableName specifies the table name for Album
func (Album) TableName() string {
	return "albums"
}
