package handlers

import (
	"net/http"

	"github.com/anonto42/nano-midea/backend/internal/models"
	"github.com/anonto42/nano-midea/backend/internal/repositories"
	"github.com/labstack/echo/v4"
)

// SavedPostHandler handles saved post HTTP requests
type SavedPostHandler struct {
	savedPostRepository repositories.SavedPostRepository
	postRepository      repositories.PostRepository
}

// NewSavedPostHandler creates a new SavedPostHandler
func NewSavedPostHandler(savedPostRepo repositories.SavedPostRepository, postRepo repositories.PostRepository) *SavedPostHandler {
	return &SavedPostHandler{
		savedPostRepository: savedPostRepo,
		postRepository:      postRepo,
	}
}

// RegisterSavedPostRoutes registers saved post routes
func (h *SavedPostHandler) RegisterSavedPostRoutes(g *echo.Group) {
	g.POST("/posts/:id/save", h.SavePost)
	g.DELETE("/posts/:id/save", h.UnsavePost)
}

// SavePost saves/bookmarks a post
func (h *SavedPostHandler) SavePost(c echo.Context) error {
	currentUserID := getUserIDFromContext(c)
	if currentUserID == 0 {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	postID := c.Param("id")

	// Verify post exists
	_, err := h.postRepository.GetPostByID(c.Request().Context(), postID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Post not found")
	}

	// Check if already saved
	isSaved, _ := h.savedPostRepository.IsPostSaved(currentUserID, postID)
	if isSaved {
		return echo.NewHTTPError(http.StatusConflict, "Post already saved")
	}

	savedPost := &models.SavedPost{
		UserID: currentUserID,
		PostID: postID,
	}

	if err := h.savedPostRepository.SavePost(savedPost); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, echo.Map{"success": true, "data": echo.Map{"saved": true}})
}

// UnsavePost removes a post from saved
func (h *SavedPostHandler) UnsavePost(c echo.Context) error {
	currentUserID := getUserIDFromContext(c)
	if currentUserID == 0 {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	postID := c.Param("id")

	if err := h.savedPostRepository.UnsavePost(currentUserID, postID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, echo.Map{"success": true, "data": echo.Map{"saved": false}})
}
