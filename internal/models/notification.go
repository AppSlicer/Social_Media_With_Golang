package models

import "time"

// Notification represents a user notification (PostgreSQL)
type Notification struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	Type            string    `json:"type" gorm:"size:30;index"` // like, comment, follow, mention, tag, story_reaction
	ActorID         uint      `json:"actor_id" gorm:"index"`
	RecipientID     uint      `json:"recipient_id" gorm:"index"`
	TargetID        string    `json:"target_id"`                    // post ID, comment ID, etc.
	TargetType      string    `json:"target_type" gorm:"size:20"`   // post, comment, user
	PreviewImageURL string    `json:"preview_image_url"`
	Message         string    `json:"message"`
	IsRead          bool      `json:"is_read" gorm:"default:false;index"`
	CreatedAt       time.Time `json:"created_at" gorm:"index"`
}
