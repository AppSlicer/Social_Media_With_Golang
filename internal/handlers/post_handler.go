package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/anonto42/nano-midea/backend/internal/models"
	"github.com/anonto42/nano-midea/backend/internal/repositories"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
)

// PostHandler handles HTTP requests related to posts
type PostHandler struct {
	postRepository repositories.PostRepository
	userRepository repositories.UserRepository // To fetch user details if needed, e.g., for posts feed
}

// NewPostHandler creates a new PostHandler
func NewPostHandler(postRepo repositories.PostRepository, userRepo repositories.UserRepository) *PostHandler {
	return &PostHandler{
		postRepository: postRepo,
		userRepository: userRepo,
	}
}

// RegisterPostRoutes registers post-related routes
func (h *PostHandler) RegisterPostRoutes(g *echo.Group) {
	g.POST("/posts", h.CreatePost)
	g.GET("/posts/:id", h.GetPost)
	g.GET("/posts", h.GetPosts) // Get all posts or posts by user (with query param)
	g.PUT("/posts/:id", h.UpdatePost)
	g.DELETE("/posts/:id", h.DeletePost)
}

// CreatePost creates a new post
func (h *PostHandler) CreatePost(c echo.Context) error {
	firebaseUID := c.Get("firebaseUID").(string) // Get Firebase UID from middleware

	var req models.CreatePostRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	post := &models.Post{
		UserID:    firebaseUID,
		Content:   req.Content,
		ImageURLs: req.ImageURLs,
		VideoURLs: req.VideoURLs,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.postRepository.CreatePost(c.Request().Context(), post); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, post)
}

// GetPost retrieves a post by ID
func (h *PostHandler) GetPost(c echo.Context) error {
	postID := c.Param("id")

	post, err := h.postRepository.GetPostByID(c.Request().Context(), postID)
	if err != nil {
		if err.Error() == "post not found" || err == mongo.ErrNoDocuments {
			return echo.NewHTTPError(http.StatusNotFound, "Post not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, post)
}

// GetPosts retrieves multiple posts
func (h *PostHandler) GetPosts(c echo.Context) error {
	userID := c.QueryParam("user_id")
	skip, _ := strconv.ParseInt(c.QueryParam("skip"), 10, 64)
	limit, _ := strconv.ParseInt(c.QueryParam("limit"), 10, 64)
	if limit == 0 {
		limit = 10 // Default limit
	}

	var posts []models.Post
	var err error

	if userID != "" {
		posts, err = h.postRepository.GetPostsByUserID(c.Request().Context(), userID, skip, limit)
	} else {
		posts, err = h.postRepository.GetAllPosts(c.Request().Context(), skip, limit)
	}

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, posts)
}

// UpdatePost updates an existing post
func (h *PostHandler) UpdatePost(c echo.Context) error {
	firebaseUID := c.Get("firebaseUID").(string) // Get Firebase UID from middleware
	postID := c.Param("id")

	var req models.UpdatePostRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	existingPost, err := h.postRepository.GetPostByID(c.Request().Context(), postID)
	if err != nil {
		if err.Error() == "post not found" || err == mongo.ErrNoDocuments {
			return echo.NewHTTPError(http.StatusNotFound, "Post not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Ensure the user updating the post is the owner
	if existingPost.UserID != firebaseUID {
		return echo.NewHTTPError(http.StatusForbidden, "You are not authorized to update this post")
	}

	if req.Content != "" {
		existingPost.Content = req.Content
	}
	if req.ImageURLs != nil {
		existingPost.ImageURLs = req.ImageURLs
	}
	if req.VideoURLs != nil {
		existingPost.VideoURLs = req.VideoURLs
	}
	existingPost.UpdatedAt = time.Now()

	if err := h.postRepository.UpdatePost(c.Request().Context(), postID, existingPost); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, existingPost)
}

// DeletePost deletes a post
func (h *PostHandler) DeletePost(c echo.Context) error {
	firebaseUID := c.Get("firebaseUID").(string) // Get Firebase UID from middleware
	postID := c.Param("id")

	existingPost, err := h.postRepository.GetPostByID(c.Request().Context(), postID)
	if err != nil {
		if err.Error() == "post not found" || err == mongo.ErrNoDocuments {
			return echo.NewHTTPError(http.StatusNotFound, "Post not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Ensure the user deleting the post is the owner
	if existingPost.UserID != firebaseUID {
		return echo.NewHTTPError(http.StatusForbidden, "You are not authorized to delete this post")
	}

	if err := h.postRepository.DeletePost(c.Request().Context(), postID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}
