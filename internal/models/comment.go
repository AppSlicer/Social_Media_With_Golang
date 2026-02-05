package models

import "gorm.io/gorm"

// Comment represents a comment on a post
type Comment struct {
	gorm.Model
	PostID    string `json:"post_id" gorm:"index"` // ID of the post the comment belongs to (MongoDB ObjectID as string)
	UserID    uint   `json:"user_id" gorm:"index"` // ID of the user who made the comment
	Content   string `json:"content" validate:"required,min=1,max=500"`
}

// CreateCommentRequest defines the request body for creating a new comment
type CreateCommentRequest struct {
	PostID  string `json:"post_id" validate:"required"`
	Content string `json:"content" validate:"required,min=1,max=500"`
}

// UpdateCommentRequest defines the request body for updating an existing comment
type UpdateCommentRequest struct {
	Content string `json:"content" validate:"required,min=1,max=500"`
}
