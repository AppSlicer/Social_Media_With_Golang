package models

import "gorm.io/gorm"

// Like represents a like on a post
type Like struct {
	gorm.Model
	PostID string `json:"post_id" gorm:"index"` // ID of the post that was liked (MongoDB ObjectID as string)
	UserID uint   `json:"user_id" gorm:"index"` // ID of the user who liked the post
}

// CreateLikeRequest defines the request body for liking a post
type CreateLikeRequest struct {
	PostID string `json:"post_id" validate:"required"`
}