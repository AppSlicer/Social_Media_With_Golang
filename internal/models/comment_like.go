package models

import "time"

// CommentLike represents a like on a comment
type CommentLike struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CommentID uint      `json:"comment_id" gorm:"index;uniqueIndex:idx_comment_user_like"`
	UserID    uint      `json:"user_id" gorm:"index;uniqueIndex:idx_comment_user_like"`
	CreatedAt time.Time `json:"created_at"`
}
