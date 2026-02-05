package handlers

import (
	"net/http"
	"strconv"

	"github.com/anonto42/nano-midea/backend/internal/models"
	"github.com/anonto42/nano-midea/backend/internal/repositories"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// FriendshipHandler handles HTTP requests related to friendships
type FriendshipHandler struct {
	friendshipRepository repositories.FriendshipRepository
	userRepository       repositories.UserRepository // To fetch user details for friends list
}

// NewFriendshipHandler creates a new FriendshipHandler
func NewFriendshipHandler(friendshipRepo repositories.FriendshipRepository, userRepo repositories.UserRepository) *FriendshipHandler {
	return &FriendshipHandler{
		friendshipRepository: friendshipRepo,
		userRepository:       userRepo,
	}
}

// RegisterFriendshipRoutes registers friendship-related routes
func (h *FriendshipHandler) RegisterFriendshipRoutes(g *echo.Group) {
	g.POST("/friends/request", h.SendFriendRequest)
	g.GET("/friends/requests/pending", h.GetPendingFriendRequests)
	g.PUT("/friends/request/:id/status", h.UpdateFriendRequestStatus)
	g.GET("/friends", h.GetFriends)
	g.DELETE("/friends/:id", h.DeleteFriend) // Unfriend
}

// SendFriendRequest handles sending a friend request
func (h *FriendshipHandler) SendFriendRequest(c echo.Context) error {
	firebaseUID := c.Get("firebaseUID").(string) // Get Firebase UID from middleware

	var req models.CreateFriendRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Get sender's user ID from our PostgreSQL database using Firebase UID
	senderUser, err := h.userRepository.GetUserByFirebaseUID(firebaseUID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Authenticated user not found in database")
	}

	// Check if receiver exists
	_, err = h.userRepository.GetUserByID(req.ReceiverID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "Receiver user not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if senderUser.ID == req.ReceiverID {
		return echo.NewHTTPError(http.StatusBadRequest, "Cannot send a friend request to yourself")
	}

	friendRequest := &models.FriendRequest{
		SenderID:   senderUser.ID,
		ReceiverID: req.ReceiverID,
		Status:     "pending", // Default status
	}

	if err := h.friendshipRepository.SendFriendRequest(friendRequest); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, friendRequest)
}

// GetPendingFriendRequests retrieves pending friend requests for the authenticated user
func (h *FriendshipHandler) GetPendingFriendRequests(c echo.Context) error {
	firebaseUID := c.Get("firebaseUID").(string) // Get Firebase UID from middleware

	receiverUser, err := h.userRepository.GetUserByFirebaseUID(firebaseUID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Authenticated user not found in database")
	}

	requests, err := h.friendshipRepository.GetUserPendingFriendRequests(receiverUser.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, requests)
}

// UpdateFriendRequestStatus updates the status of a friend request (accept/reject)
func (h *FriendshipHandler) UpdateFriendRequestStatus(c echo.Context) error {
	firebaseUID := c.Get("firebaseUID").(string) // Get Firebase UID from middleware
	requestID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request ID")
	}

	var req models.UpdateFriendRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	receiverUser, err := h.userRepository.GetUserByFirebaseUID(firebaseUID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Authenticated user not found in database")
	}

	friendRequest, err := h.friendshipRepository.GetFriendRequestByID(uint(requestID))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "Friend request not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Ensure the authenticated user is the receiver of the request
	if friendRequest.ReceiverID != receiverUser.ID {
		return echo.NewHTTPError(http.StatusForbidden, "You are not authorized to modify this friend request")
	}

	if err := h.friendshipRepository.UpdateFriendRequestStatus(uint(requestID), req.Status); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	friendRequest.Status = req.Status
	return c.JSON(http.StatusOK, friendRequest)
}

// GetFriends retrieves the list of friends for the authenticated user
func (h *FriendshipHandler) GetFriends(c echo.Context) error {
	firebaseUID := c.Get("firebaseUID").(string) // Get Firebase UID from middleware

	currentUser, err := h.userRepository.GetUserByFirebaseUID(firebaseUID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Authenticated user not found in database")
	}

	friends, err := h.friendshipRepository.GetUserFriends(currentUser.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, friends)
}

// DeleteFriend handles unfriending (deleting an accepted friend request)
func (h *FriendshipHandler) DeleteFriend(c echo.Context) error {
	firebaseUID := c.Get("firebaseUID").(string) // Get Firebase UID from middleware
	friendUserID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid friend user ID")
	}

	currentUser, err := h.userRepository.GetUserByFirebaseUID(firebaseUID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Authenticated user not found in database")
	}

	// Find the accepted friend request between current user and friendUserID
	var friendRequest *models.FriendRequest
	friendRequest, err = h.friendshipRepository.GetFriendRequestBySenderReceiver(currentUser.ID, uint(friendUserID))
	if err != nil {
		friendRequest, err = h.friendshipRepository.GetFriendRequestBySenderReceiver(uint(friendUserID), currentUser.ID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return echo.NewHTTPError(http.StatusNotFound, "Friendship not found")
			}
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}

	if friendRequest.Status != "accepted" {
		return echo.NewHTTPError(http.StatusBadRequest, "Users are not friends")
	}

	if err := h.friendshipRepository.DeleteFriendRequest(friendRequest.ID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}
