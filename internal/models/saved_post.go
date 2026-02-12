package models

import "time"

// SavedPost represents a bookmarked/saved post by a user
type SavedPost struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"index;uniqueIndex:idx_user_post_save"`
	PostID    string    `json:"post_id" gorm:"index;uniqueIndex:idx_user_post_save"`
	CreatedAt time.Time `json:"created_at"`
}
