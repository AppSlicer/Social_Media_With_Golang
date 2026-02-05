package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/anonto42/nano-midea/backend/internal/models"
	"github.com/anonto42/nano-midea/backend/internal/repositories"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// CommentHandler handles HTTP requests related to comments
type CommentHandler struct {
	commentRepository repositories.CommentRepository
	postRepository    repositories.PostRepository // To update comment counts in posts
	userRepository    repositories.UserRepository // To fetch user details for comments
}

// NewCommentHandler creates a new CommentHandler
func NewCommentHandler(commentRepo repositories.CommentRepository, postRepo repositories.PostRepository, userRepo repositories.UserRepository) *CommentHandler {
	return &CommentHandler{
		commentRepository: commentRepo,
		postRepository:    postRepo,
		userRepository:    userRepo,
	}
}

// RegisterCommentRoutes registers comment-related routes
func (h *CommentHandler) RegisterCommentRoutes(g *echo.Group) {
	g.POST("/posts/:post_id/comments", h.CreateComment)
	g.GET("/posts/:post_id/comments", h.GetCommentsByPostID)
	g.PUT("/comments/:id", h.UpdateComment)
	g.DELETE("/comments/:id", h.DeleteComment)
}

// CreateComment creates a new comment on a post
func (h *CommentHandler) CreateComment(c echo.Context) error {
	firebaseUID := c.Get("firebaseUID").(string) // Get Firebase UID from middleware
	postID := c.Param("post_id")

	var req models.CreateCommentRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

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

	comment := &models.Comment{
		PostID:  postID,
		UserID:  user.ID,
		Content: req.Content,
	}

	if err := h.commentRepository.CreateComment(comment); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Increment comments count in the post
	go h.postRepository.IncrementCommentsCount(context.Background(), postID)

	return c.JSON(http.StatusCreated, comment)
}

// GetCommentsByPostID retrieves all comments for a specific post
func (h *CommentHandler) GetCommentsByPostID(c echo.Context) error {
	postID := c.Param("post_id")

	// Verify post exists
	_, err := h.postRepository.GetPostByID(c.Request().Context(), postID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Post not found")
	}

	comments, err := h.commentRepository.GetCommentsByPostID(postID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, comments)
}

// UpdateComment updates an existing comment
func (h *CommentHandler) UpdateComment(c echo.Context) error {
	firebaseUID := c.Get("firebaseUID").(string) // Get Firebase UID from middleware
	commentID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid comment ID")
	}

	var req models.UpdateCommentRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	user, err := h.userRepository.GetUserByFirebaseUID(firebaseUID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Authenticated user not found in database")
	}

	comment, err := h.commentRepository.GetCommentByID(uint(commentID))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "Comment not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Ensure the user updating the comment is the owner
	if comment.UserID != user.ID {
		return echo.NewHTTPError(http.StatusForbidden, "You are not authorized to update this comment")
	}

	comment.Content = req.Content

	if err := h.commentRepository.UpdateComment(comment); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, comment)
}

// DeleteComment deletes a comment
func (h *CommentHandler) DeleteComment(c echo.Context) error {
	firebaseUID := c.Get("firebaseUID").(string) // Get Firebase UID from middleware
	commentID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid comment ID")
	}

	user, err := h.userRepository.GetUserByFirebaseUID(firebaseUID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Authenticated user not found in database")
	}

	comment, err := h.commentRepository.GetCommentByID(uint(commentID))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "Comment not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Ensure the user deleting the comment is the owner
	if comment.UserID != user.ID {
		return echo.NewHTTPError(http.StatusForbidden, "You are not authorized to delete this comment")
	}

	if err := h.commentRepository.DeleteComment(uint(commentID)); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Decrement comments count in the post
	go h.postRepository.DecrementCommentsCount(context.Background(), comment.PostID)

	return c.NoContent(http.StatusNoContent)
}
