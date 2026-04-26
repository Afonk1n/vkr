package models

import "time"

// UserFollow is a follower → following edge between users (no soft delete).
type UserFollow struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	FollowerID  uint      `json:"follower_id" gorm:"not null;index;uniqueIndex:ux_user_follow_pair"`
	FollowingID uint      `json:"following_id" gorm:"not null;index;uniqueIndex:ux_user_follow_pair"`
	CreatedAt   time.Time `json:"created_at"`

	Follower  User `json:"-" gorm:"foreignKey:FollowerID"`
	Following User `json:"-" gorm:"foreignKey:FollowingID"`
}

func (UserFollow) TableName() string {
	return "user_follows"
}
