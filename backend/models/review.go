package models

import (
	"time"

	"gorm.io/gorm"
)

// ReviewStatus represents the status of a review
type ReviewStatus string

const (
	ReviewStatusPending  ReviewStatus = "pending"
	ReviewStatusApproved ReviewStatus = "approved"
	ReviewStatusRejected ReviewStatus = "rejected"
)

// Review represents a review of an album or track
type Review struct {
	ID                   uint           `json:"id" gorm:"primaryKey"`
	UserID               uint           `json:"user_id" gorm:"not null"`
	AlbumID              *uint          `json:"album_id"` // Nullable - either album_id or track_id must be set
	TrackID              *uint          `json:"track_id"` // Nullable - either album_id or track_id must be set
	Text                 string         `json:"text" gorm:"type:text"`
	RatingRhymes         int            `json:"rating_rhymes" gorm:"not null;check:rating_rhymes >= 1 AND rating_rhymes <= 10"`
	RatingStructure      int            `json:"rating_structure" gorm:"not null;check:rating_structure >= 1 AND rating_structure <= 10"`
	RatingImplementation int            `json:"rating_implementation" gorm:"not null;check:rating_implementation >= 1 AND rating_implementation <= 10"`
	RatingIndividuality  int            `json:"rating_individuality" gorm:"not null;check:rating_individuality >= 1 AND rating_individuality <= 10"`
	AtmosphereMultiplier float64        `json:"atmosphere_multiplier" gorm:"not null;check:atmosphere_multiplier >= 1.0000 AND atmosphere_multiplier <= 1.6072"`
	FinalScore           float64        `json:"final_score" gorm:"not null"`
	Status               ReviewStatus   `json:"status" gorm:"default:'pending'"`
	ModeratedBy          *uint          `json:"moderated_by"`
	ModeratedAt          *time.Time     `json:"moderated_at"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	DeletedAt            gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	User      User        `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Album     *Album      `json:"album,omitempty" gorm:"foreignKey:AlbumID"`
	Track     *Track      `json:"track,omitempty" gorm:"foreignKey:TrackID"`
	Moderator *User       `json:"moderator,omitempty" gorm:"foreignKey:ModeratedBy"`
	Likes     []ReviewLike `json:"likes,omitempty" gorm:"foreignKey:ReviewID"`
}

// TableName specifies the table name for Review
func (Review) TableName() string {
	return "reviews"
}

// CalculateFinalScore calculates the final score based on the rating formula
// Formula: (Рифмы+Структура+Реализация+Индивидуальность) × 1.4 × Атмосфера/Вайб
// Result is rounded to the nearest integer
func (r *Review) CalculateFinalScore() {
	baseScore := float64(r.RatingRhymes + r.RatingStructure + r.RatingImplementation + r.RatingIndividuality)
	score := baseScore * 1.4 * r.AtmosphereMultiplier
	r.FinalScore = float64(int(score + 0.5)) // Round to nearest integer
}

