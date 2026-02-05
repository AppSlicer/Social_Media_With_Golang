package repositories

import (
	"fmt"

	"github.com/anonto42/nano-midea/backend/internal/models"
	"gorm.io/gorm"
)

// FriendshipRepository defines the interface for friendship data operations
type FriendshipRepository interface {
	SendFriendRequest(req *models.FriendRequest) error
	GetFriendRequestByID(id uint) (*models.FriendRequest, error)
	GetFriendRequestBySenderReceiver(senderID, receiverID uint) (*models.FriendRequest, error)
	GetUserPendingFriendRequests(userID uint) ([]models.FriendRequest, error)
	GetUserFriends(userID uint) ([]models.User, error)
	UpdateFriendRequestStatus(id uint, status string) error
	DeleteFriendRequest(id uint) error
}

// PostgresFriendshipRepository implements FriendshipRepository for PostgreSQL
type PostgresFriendshipRepository struct {
	db *gorm.DB
}

// NewPostgresFriendshipRepository creates a new PostgresFriendshipRepository
func NewPostgresFriendshipRepository(db *gorm.DB) *PostgresFriendshipRepository {
	return &PostgresFriendshipRepository{db: db}
}

// SendFriendRequest creates a new friend request
func (r *PostgresFriendshipRepository) SendFriendRequest(req *models.FriendRequest) error {
	// Check if a request already exists or if they are already friends
	var existingRequest models.FriendRequest
	err := r.db.Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
		req.SenderID, req.ReceiverID, req.ReceiverID, req.SenderID).First(&existingRequest).Error

	if err == nil {
		if existingRequest.Status == "pending" {
			return fmt.Errorf("a pending friend request already exists between these users")
		} else if existingRequest.Status == "accepted" {
			return fmt.Errorf("users are already friends")
		}
	} else if err != gorm.ErrRecordNotFound {
		return err
	}

	req.Status = "pending"
	return r.db.Create(req).Error
}

// GetFriendRequestByID retrieves a friend request by ID
func (r *PostgresFriendshipRepository) GetFriendRequestByID(id uint) (*models.FriendRequest, error) {
	var req models.FriendRequest
	if err := r.db.First(&req, id).Error; err != nil {
		return nil, err
	}
	return &req, nil
}

// GetFriendRequestBySenderReceiver retrieves a friend request by sender and receiver IDs
func (r *PostgresFriendshipRepository) GetFriendRequestBySenderReceiver(senderID, receiverID uint) (*models.FriendRequest, error) {
	var req models.FriendRequest
	if err := r.db.Where("sender_id = ? AND receiver_id = ?", senderID, receiverID).First(&req).Error; err != nil {
		return nil, err
	}
	return &req, nil
}


// GetUserPendingFriendRequests retrieves all pending friend requests for a user
func (r *PostgresFriendshipRepository) GetUserPendingFriendRequests(userID uint) ([]models.FriendRequest, error) {
	var requests []models.FriendRequest
	if err := r.db.Where("receiver_id = ? AND status = ?", userID, "pending").Find(&requests).Error; err != nil {
		return nil, err
	}
	return requests, nil
}

// GetUserFriends retrieves all accepted friends for a user
func (r *PostgresFriendshipRepository) GetUserFriends(userID uint) ([]models.User, error) {
	var friends []models.User
	// Find requests where current user is sender and status is accepted
	// Or where current user is receiver and status is accepted
	subQuery1 := r.db.Table("friend_requests").Select("receiver_id").Where("sender_id = ? AND status = ?", userID, "accepted")
	subQuery2 := r.db.Table("friend_requests").Select("sender_id").Where("receiver_id = ? AND status = ?", userID, "accepted")

	if err := r.db.Where("id IN (?) OR id IN (?)", subQuery1, subQuery2).Find(&friends).Error; err != nil {
		return nil, err
	}
	return friends, nil
}


// UpdateFriendRequestStatus updates the status of a friend request
func (r *PostgresFriendshipRepository) UpdateFriendRequestStatus(id uint, status string) error {
	return r.db.Model(&models.FriendRequest{}).Where("id = ?", id).Update("status", status).Error
}

// DeleteFriendRequest deletes a friend request
func (r *PostgresFriendshipRepository) DeleteFriendRequest(id uint) error {
	return r.db.Delete(&models.FriendRequest{}, id).Error
}
