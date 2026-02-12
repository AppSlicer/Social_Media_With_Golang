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
	commentRepository      repositories.CommentRepository
	postRepository         repositories.PostRepository
	userRepository         repositories.UserRepository
	commentLikeRepository  repositories.CommentLikeRepository
	notificationRepository repositories.NotificationRepository
}

// NewCommentHandler creates a new CommentHandler
func NewCommentHandler(
	commentRepo repositories.CommentRepository,
	postRepo repositories.PostRepository,
	userRepo repositories.UserRepository,
	commentLikeRepo repositories.CommentLikeRepository,
	notifRepo repositories.NotificationRepository,
) *CommentHandler {
	return &CommentHandler{
		commentRepository:      commentRepo,
		postRepository:         postRepo,
		userRepository:         userRepo,
		commentLikeRepository:  commentLikeRepo,
		notificationRepository: notifRepo,
	}
}

// RegisterCommentRoutes registers comment-related routes
func (h *CommentHandler) RegisterCommentRoutes(g *echo.Group) {
	g.POST("/posts/:post_id/comments", h.CreateComment)
	g.GET("/posts/:post_id/comments", h.GetCommentsByPostID)
	g.PUT("/comments/:id", h.UpdateComment)
	g.DELETE("/comments/:id", h.DeleteComment)
	g.POST("/comments/:id/like", h.LikeComment)
	g.DELETE("/comments/:id/like", h.UnlikeComment)
}

// CreateComment creates a new comment on a post
func (h *CommentHandler) CreateComment(c echo.Context) error {
	currentUserID := getUserIDFromContext(c)
	if currentUserID == 0 {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}
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
	post, err := h.postRepository.GetPostByID(c.Request().Context(), postID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Post not found")
	}

	comment := &models.Comment{
		PostID:  postID,
		UserID:  currentUserID,
		Content: req.Content,
	}

	if err := h.commentRepository.CreateComment(comment); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Increment comments count in the post
	go h.postRepository.IncrementCommentsCount(context.Background(), postID)

	// Create notification for post owner
	if h.notificationRepository != nil {
		actor, _ := h.userRepository.GetUserByID(currentUserID)
		if actor != nil && post.UserID != "" {
			recipient, err := h.userRepository.GetUserByFirebaseUID(post.UserID)
			if err == nil && recipient.ID != currentUserID {
				notif := &models.Notification{
					Type:        "comment",
					ActorID:     currentUserID,
					RecipientID: recipient.ID,
					TargetID:    postID,
					TargetType:  "post",
					Message:     actor.DisplayName + " commented on your post",
				}
				h.notificationRepository.CreateNotification(notif)
			}
		}
	}

	// Get author info
	user, _ := h.userRepository.GetUserByID(currentUserID)
	var author models.UserCompact
	if user != nil {
		author = user.ToCompact()
	}

	return c.JSON(http.StatusCreated, echo.Map{
		"success": true,
		"data": echo.Map{
			"comment": echo.Map{
				"id":            comment.ID,
				"post_id":       comment.PostID,
				"author":        author,
				"content":       comment.Content,
				"likes_count":   0,
				"is_liked":      false,
				"replies_count": 0,
				"parent_id":     nil,
				"created_at":    comment.CreatedAt,
			},
		},
	})
}

// GetCommentsByPostID retrieves all comments for a specific post
func (h *CommentHandler) GetCommentsByPostID(c echo.Context) error {
	currentUserID := getUserIDFromContext(c)
	postID := c.Param("post_id")

	_, err := h.postRepository.GetPostByID(c.Request().Context(), postID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Post not found")
	}

	comments, err := h.commentRepository.GetCommentsByPostID(postID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	enriched := make([]echo.Map, len(comments))
	userCache := make(map[uint]models.UserCompact)

	for i, comment := range comments {
		var author models.UserCompact
		if cached, ok := userCache[comment.UserID]; ok {
			author = cached
		} else {
			user, err := h.userRepository.GetUserByID(comment.UserID)
			if err == nil {
				author = user.ToCompact()
				userCache[comment.UserID] = author
			}
		}

		isLiked := false
		if currentUserID > 0 && h.commentLikeRepository != nil {
			isLiked, _ = h.commentLikeRepository.HasUserLikedComment(comment.ID, currentUserID)
		}

		likesCount := int64(0)
		if h.commentLikeRepository != nil {
			likesCount, _ = h.commentLikeRepository.GetLikesCount(comment.ID)
		}

		enriched[i] = echo.Map{
			"id":            comment.ID,
			"post_id":       comment.PostID,
			"author":        author,
			"content":       comment.Content,
			"likes_count":   likesCount,
			"is_liked":      isLiked,
			"replies_count": 0,
			"parent_id":     nil,
			"created_at":    comment.CreatedAt,
		}
	}

	return c.JSON(http.StatusOK, echo.Map{
		"success": true,
		"data":    echo.Map{"comments": enriched},
	})
}

// UpdateComment updates an existing comment
func (h *CommentHandler) UpdateComment(c echo.Context) error {
	currentUserID := getUserIDFromContext(c)
	if currentUserID == 0 {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

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

	comment, err := h.commentRepository.GetCommentByID(uint(commentID))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "Comment not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if comment.UserID != currentUserID {
		return echo.NewHTTPError(http.StatusForbidden, "You are not authorized to update this comment")
	}

	comment.Content = req.Content
	if err := h.commentRepository.UpdateComment(comment); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, echo.Map{"success": true, "data": echo.Map{"comment": comment}})
}

// DeleteComment deletes a comment
func (h *CommentHandler) DeleteComment(c echo.Context) error {
	currentUserID := getUserIDFromContext(c)
	if currentUserID == 0 {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	commentID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid comment ID")
	}

	comment, err := h.commentRepository.GetCommentByID(uint(commentID))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "Comment not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if comment.UserID != currentUserID {
		return echo.NewHTTPError(http.StatusForbidden, "You are not authorized to delete this comment")
	}

	if err := h.commentRepository.DeleteComment(uint(commentID)); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	go h.postRepository.DecrementCommentsCount(context.Background(), comment.PostID)

	return c.JSON(http.StatusOK, echo.Map{"success": true, "data": echo.Map{"deleted": true}})
}

// LikeComment likes a comment
func (h *CommentHandler) LikeComment(c echo.Context) error {
	currentUserID := getUserIDFromContext(c)
	if currentUserID == 0 {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	commentID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid comment ID")
	}

	hasLiked, _ := h.commentLikeRepository.HasUserLikedComment(uint(commentID), currentUserID)
	if hasLiked {
		return echo.NewHTTPError(http.StatusConflict, "Comment already liked")
	}

	like := &models.CommentLike{
		CommentID: uint(commentID),
		UserID:    currentUserID,
	}

	if err := h.commentLikeRepository.CreateCommentLike(like); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, echo.Map{"success": true, "data": echo.Map{"liked": true}})
}

// UnlikeComment unlikes a comment
func (h *CommentHandler) UnlikeComment(c echo.Context) error {
	currentUserID := getUserIDFromContext(c)
	if currentUserID == 0 {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	commentID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid comment ID")
	}

	if err := h.commentLikeRepository.DeleteCommentLike(uint(commentID), currentUserID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, echo.Map{"success": true, "data": echo.Map{"liked": false}})
}
