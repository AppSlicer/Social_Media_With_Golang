package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Story represents a user's story stored in MongoDB
type Story struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	UserID    string             `json:"user_id" bson:"user_id"`
	Items     []StoryItem        `json:"items" bson:"items"`
	ExpiresAt time.Time          `json:"expires_at" bson:"expires_at"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
}

// StoryItem represents a single item in a story
type StoryItem struct {
	ID        string    `json:"id" bson:"id"`
	Type      string    `json:"type" bson:"type"` // "image" or "video"
	URL       string    `json:"url" bson:"url"`
	Duration  int       `json:"duration" bson:"duration"` // seconds
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

// StorySeen tracks which stories a user has seen (PostgreSQL)
type StorySeen struct {
	ID      uint      `json:"id" gorm:"primaryKey"`
	StoryID string    `json:"story_id" gorm:"index;uniqueIndex:idx_story_user_seen"`
	UserID  uint      `json:"user_id" gorm:"index;uniqueIndex:idx_story_user_seen"`
	SeenAt  time.Time `json:"seen_at"`
}

// StoryReaction tracks reactions to stories (PostgreSQL)
type StoryReaction struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	StoryID   string    `json:"story_id" gorm:"index"`
	UserID    uint      `json:"user_id" gorm:"index"`
	Reaction  string    `json:"reaction"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateStoryRequest defines the request body for creating a story
type CreateStoryRequest struct {
	MediaURL string `json:"media_url" validate:"required"`
	Type     string `json:"type" validate:"required,oneof=image video"`
}
