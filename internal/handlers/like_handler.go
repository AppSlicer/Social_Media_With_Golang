package handlers

import (
	"net/http"

	"github.com/anonto42/nano-midea/backend/internal/models"
	"github.com/anonto42/nano-midea/backend/internal/repositories"
	"github.com/labstack/echo/v4"
)

// LikeHandler handles HTTP requests related to likes
type LikeHandler struct {
	likeRepository repositories.LikeRepository
	postRepository repositories.PostRepository // To update like counts in posts
	userRepository repositories.UserRepository // To fetch user details for likes
}

// NewLikeHandler creates a new LikeHandler
func NewLikeHandler(likeRepo repositories.LikeRepository, postRepo repositories.PostRepository, userRepo repositories.UserRepository) *LikeHandler {
	return &LikeHandler{
		likeRepository: likeRepo,
		postRepository: postRepo,
		userRepository: userRepo,
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
	firebaseUID := c.Get("firebaseUID").(string) // Get Firebase UID from middleware
	postID := c.Param("post_id")

	// Verify post exists
	_, err := h.postRepository.GetPostByID(c.Request().Context(), postID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Post not found")
	}

	// Get user ID from PostgreSQL using Firebase UID
	user, err := h.userRepository.GetUserByFirebaseUID(firebaseUID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Authenticated user not found in database")
	}

	// Check if user has already liked the post
	hasLiked, err := h.likeRepository.HasUserLikedPost(postID, user.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if hasLiked {
		return echo.NewHTTPError(http.StatusConflict, "Post already liked by this user")
	}

	like := &models.Like{
		PostID: postID,
		UserID: user.ID,
	}

	if err := h.likeRepository.CreateLike(like); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Increment likes count in the post
	go h.postRepository.IncrementLikesCount(c.Request().Context(), postID)

	return c.JSON(http.StatusCreated, like)
}

// UnlikePost handles unliking a post
func (h *LikeHandler) UnlikePost(c echo.Context) error {
	firebaseUID := c.Get("firebaseUID").(string) // Get Firebase UID from middleware
	postID := c.Param("post_id")

	// Verify post exists
	_, err := h.postRepository.GetPostByID(c.Request().Context(), postID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Post not found")
	}

	// Get user ID from PostgreSQL using Firebase UID
	user, err := h.userRepository.GetUserByFirebaseUID(firebaseUID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Authenticated user not found in database")
	}

	if err := h.likeRepository.DeleteLike(postID, user.ID); err != nil {
		if err.Error() == "like not found" {
			return echo.NewHTTPError(http.StatusNotFound, "Like not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Decrement likes count in the post
	go h.postRepository.DecrementLikesCount(c.Request().Context(), postID)

	return c.NoContent(http.StatusNoContent)
}

// GetLikesCountForPost retrieves the total number of likes for a specific post
func (h *LikeHandler) GetLikesCountForPost(c echo.Context) error {
	postID := c.Param("post_id")

	// Verify post exists
	_, err := h.postRepository.GetPostByID(c.Request().Context(), postID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Post not found")
	}

	count, err := h.likeRepository.GetLikesCountByPostID(postID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, echo.Map{"post_id": postID, "likes_count": count})
}

// GetUserLikeStatusForPost checks if the authenticated user has liked a specific post
func (h *LikeHandler) GetUserLikeStatusForPost(c echo.Context) error {
	firebaseUID := c.Get("firebaseUID").(string) // Get Firebase UID from middleware
	postID := c.Param("post_id")

	// Verify post exists
	_, err := h.postRepository.GetPostByID(c.Request().Context(), postID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Post not found")
	}

	// Get user ID from PostgreSQL using Firebase UID
	user, err := h.userRepository.GetUserByFirebaseUID(firebaseUID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Authenticated user not found in database")
	}

	hasLiked, err := h.likeRepository.HasUserLikedPost(postID, user.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, echo.Map{"post_id": postID, "user_id": user.ID, "has_liked": hasLiked})
}
