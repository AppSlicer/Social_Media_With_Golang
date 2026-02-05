package models

import "gorm.io/gorm"

// FriendRequest represents a friend request between two users
type FriendRequest struct {
	gorm.Model
	SenderID   uint   `json:"sender_id" gorm:"index"`   // User ID of the sender
	ReceiverID uint   `json:"receiver_id" gorm:"index"` // User ID of the receiver
	Status     string `json:"status" gorm:"type:varchar(20);default:'pending'"` // e.g., "pending", "accepted", "rejected"
}

// CreateFriendRequest defines the request body for sending a friend request
type CreateFriendRequest struct {
	ReceiverID uint `json:"receiver_id" validate:"required"`
}

// UpdateFriendRequest defines the request body for accepting/rejecting a friend request
type UpdateFriendRequest struct {
	Status string `json:"status" validate:"required,oneof=accepted rejected"`
}

// Friendship represents an accepted friendship (could be implicit via FriendRequest status, but useful for direct querying)
// For simplicity, we might just query FriendRequest table for status "accepted"
// Or, if we want a separate table for accepted friendships for performance/simplicity:
/*
type Friendship struct {
	gorm.Model
	UserID1 uint `json:"user_id_1" gorm:"index"`
	UserID2 uint `json:"user_id_2" gorm:"index"`
	// Ensure unique pair regardless of order
	// gorm:"uniqueIndex:idx_user_pair;check:user_id_1 < user_id_2"
}
*/