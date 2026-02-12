package handlers

import (
	"net/http"
	"strconv"

	"github.com/anonto42/nano-midea/backend/internal/models"
	"github.com/anonto42/nano-midea/backend/internal/repositories"
	"github.com/labstack/echo/v4"
)

// FollowHandler handles follow/unfollow HTTP requests
type FollowHandler struct {
	followRepository       repositories.FollowRepository
	userRepository         repositories.UserRepository
	notificationRepository repositories.NotificationRepository
}

// NewFollowHandler creates a new FollowHandler
func NewFollowHandler(followRepo repositories.FollowRepository, userRepo repositories.UserRepository, notifRepo repositories.NotificationRepository) *FollowHandler {
	return &FollowHandler{
		followRepository:       followRepo,
		userRepository:         userRepo,
		notificationRepository: notifRepo,
	}
}

// RegisterFollowRoutes registers follow-related routes
func (h *FollowHandler) RegisterFollowRoutes(g *echo.Group) {
	g.POST("/users/:id/follow", h.FollowUser)
	g.DELETE("/users/:id/follow", h.UnfollowUser)
}

// FollowUser follows a user
func (h *FollowHandler) FollowUser(c echo.Context) error {
	currentUserID := getUserIDFromContext(c)
	if currentUserID == 0 {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	targetID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID")
	}

	if currentUserID == uint(targetID) {
		return echo.NewHTTPError(http.StatusBadRequest, "Cannot follow yourself")
	}

	// Check if already following
	isFollowing, err := h.followRepository.IsFollowing(currentUserID, uint(targetID))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if isFollowing {
		return echo.NewHTTPError(http.StatusConflict, "Already following this user")
	}

	follow := &models.Follow{
		FollowerID:  currentUserID,
		FollowingID: uint(targetID),
	}

	if err := h.followRepository.CreateFollow(follow); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Update counts
	h.userRepository.IncrementFollowingCount(currentUserID)
	h.userRepository.IncrementFollowersCount(uint(targetID))

	// Create notification
	if h.notificationRepository != nil {
		actor, _ := h.userRepository.GetUserByID(currentUserID)
		if actor != nil {
			notif := &models.Notification{
				Type:        "follow",
				ActorID:     currentUserID,
				RecipientID: uint(targetID),
				Message:     actor.DisplayName + " started following you",
			}
			h.notificationRepository.CreateNotification(notif)
		}
	}

	return c.JSON(http.StatusOK, echo.Map{"success": true, "data": echo.Map{"following": true}})
}

// UnfollowUser unfollows a user
func (h *FollowHandler) UnfollowUser(c echo.Context) error {
	currentUserID := getUserIDFromContext(c)
	if currentUserID == 0 {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	targetID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID")
	}

	if err := h.followRepository.DeleteFollow(currentUserID, uint(targetID)); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Update counts
	h.userRepository.DecrementFollowingCount(currentUserID)
	h.userRepository.DecrementFollowersCount(uint(targetID))

	return c.JSON(http.StatusOK, echo.Map{"success": true, "data": echo.Map{"following": false}})
}
