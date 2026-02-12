package models

import "time"

// Follow represents an Instagram-style follow relationship
type Follow struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	FollowerID  uint      `json:"follower_id" gorm:"index;uniqueIndex:idx_follower_following"`
	FollowingID uint      `json:"following_id" gorm:"index;uniqueIndex:idx_follower_following"`
	CreatedAt   time.Time `json:"created_at"`
}
