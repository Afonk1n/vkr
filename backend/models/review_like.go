package models

import (
	"time"

	"gorm.io/gorm"
)

// ReviewLike represents a like on a review
type ReviewLike struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UserID    uint           `json:"user_id" gorm:"not null"`
	ReviewID  uint           `json:"review_id" gorm:"not null"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	User  User  `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Review Review `json:"review,omitempty" gorm:"foreignKey:ReviewID"`
}

// TableName specifies the table name for ReviewLike
func (ReviewLike) TableName() string {
	return "review_likes"
}

// BeforeCreate ensures unique like per user per review
func (rl *ReviewLike) BeforeCreate(tx *gorm.DB) error {
	var count int64
	tx.Model(&ReviewLike{}).
		Where("user_id = ? AND review_id = ?", rl.UserID, rl.ReviewID).
		Count(&count)
	
	if count > 0 {
		return gorm.ErrDuplicatedKey
	}
	return nil
}

