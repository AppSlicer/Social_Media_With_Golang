package handlers

import (
	"net/http"

	"github.com/anonto42/nano-midea/backend/internal/models"
	"github.com/anonto42/nano-midea/backend/internal/repositories"
	"github.com/labstack/echo/v4"
)

// LikeHandler handles HTTP requests related to likes
type LikeHandler struct {
	likeRepository         repositories.LikeRepository
	postRepository         repositories.PostRepository
	userRepository         repositories.UserRepository
	notificationRepository repositories.NotificationRepository
}

// NewLikeHandler creates a new LikeHandler
func NewLikeHandler(likeRepo repositories.LikeRepository, postRepo repositories.PostRepository, userRepo repositories.UserRepository, notifRepo repositories.NotificationRepository) *LikeHandler {
	return &LikeHandler{
		likeRepository:         likeRepo,
		postRepository:         postRepo,
		userRepository:         userRepo,
		notificationRepository: notifRepo,
	}
}

// RegisterLikeRoutes registers like-related routes
func (h *LikeHandler) RegisterLikeRoutes(g *echo.Group) {
	g.POST("/posts/:post_id/likes", h.LikePost)
	g.DELETE("/posts/:post_id/likes", h.UnlikePost)
	g.GET("/posts/:post_id/likes/count", h.GetLikesCountForPost)
	g.GET("/posts/:post_id/likes/status", h.GetUserLikeStatusForPost)
}

// LikePost handles liking a post
func (h *LikeHandler) LikePost(c echo.Context) error {
	currentUserID := getUserIDFromContext(c)
	if currentUserID == 0 {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}
	postID := c.Param("post_id")

	// Verify post exists
	post, err := h.postRepository.GetPostByID(c.Request().Context(), postID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Post not found")
	}

	hasLiked, err := h.likeRepository.HasUserLikedPost(postID, currentUserID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if hasLiked {
		return echo.NewHTTPError(http.StatusConflict, "Post already liked by this user")
	}

	like := &models.Like{
		PostID: postID,
		UserID: currentUserID,
	}

	if err := h.likeRepository.CreateLike(like); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	go h.postRepository.IncrementLikesCount(c.Request().Context(), postID)

	// Create notification for post owner
	if h.notificationRepository != nil {
		actor, _ := h.userRepository.GetUserByID(currentUserID)
		if actor != nil && post.UserID != "" {
			recipient, err := h.userRepository.GetUserByFirebaseUID(post.UserID)
			if err == nil && recipient.ID != currentUserID {
				notif := &models.Notification{
					Type:        "like",
					ActorID:     currentUserID,
					RecipientID: recipient.ID,
					TargetID:    postID,
					TargetType:  "post",
					Message:     actor.DisplayName + " liked your post",
				}
				h.notificationRepository.CreateNotification(notif)
			}
		}
	}

	return c.JSON(http.StatusOK, echo.Map{"success": true, "data": echo.Map{"liked": true}})
}

// UnlikePost handles unliking a post
func (h *LikeHandler) UnlikePost(c echo.Context) error {
	currentUserID := getUserIDFromContext(c)
	if currentUserID == 0 {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}
	postID := c.Param("post_id")

	_, err := h.postRepository.GetPostByID(c.Request().Context(), postID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Post not found")
	}

	if err := h.likeRepository.DeleteLike(postID, currentUserID); err != nil {
		if err.Error() == "like not found" {
			return echo.NewHTTPError(http.StatusNotFound, "Like not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	go h.postRepository.DecrementLikesCount(c.Request().Context(), postID)

	return c.JSON(http.StatusOK, echo.Map{"success": true, "data": echo.Map{"liked": false}})
}

// GetLikesCountForPost retrieves the total number of likes for a specific post
func (h *LikeHandler) GetLikesCountForPost(c echo.Context) error {
	postID := c.Param("post_id")

	_, err := h.postRepository.GetPostByID(c.Request().Context(), postID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Post not found")
	}

	count, err := h.likeRepository.GetLikesCountByPostID(postID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, echo.Map{"success": true, "data": echo.Map{"post_id": postID, "likes_count": count}})
}

// GetUserLikeStatusForPost checks if the authenticated user has liked a specific post
func (h *LikeHandler) GetUserLikeStatusForPost(c echo.Context) error {
	currentUserID := getUserIDFromContext(c)
	if currentUserID == 0 {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}
	postID := c.Param("post_id")

	_, err := h.postRepository.GetPostByID(c.Request().Context(), postID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Post not found")
	}

	hasLiked, err := h.likeRepository.HasUserLikedPost(postID, currentUserID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, echo.Map{"success": true, "data": echo.Map{"post_id": postID, "has_liked": hasLiked}})
}
