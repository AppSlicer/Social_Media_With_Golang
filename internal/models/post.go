package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Post represents a social media post stored in MongoDB
type Post struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	UserID    string             `json:"user_id" bson:"user_id"` // Firebase UID of the user who created the post
	Content   string             `json:"content" bson:"content"`
	ImageURLs []string           `json:"image_urls,omitempty" bson:"image_urls,omitempty"`
	VideoURLs []string           `json:"video_urls,omitempty" bson:"video_urls,omitempty"`
	LikesCount    int                `json:"likes_count" bson:"likes_count"`
	CommentsCount int                `json:"comments_count" bson:"comments_count"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}

// CreatePostRequest defines the request body for creating a new post
type CreatePostRequest struct {
	Content   string   `json:"content" validate:"required,min=1,max=280"`
	ImageURLs []string `json:"image_urls,omitempty" validate:"omitempty,dive,url"`
	VideoURLs []string `json:"video_urls,omitempty" validate:"omitempty,dive,url"`
}

// UpdatePostRequest defines the request body for updating an existing post
type UpdatePostRequest struct {
	Content   string   `json:"content,omitempty" validate:"omitempty,min=1,max=280"`
	ImageURLs []string `json:"image_urls,omitempty" validate:"omitempty,dive,url"`
	VideoURLs []string `json:"video_urls,omitempty" validate:"omitempty,dive,url"`
}
