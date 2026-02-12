package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/anonto42/nano-midea/backend/internal/models"
	"github.com/anonto42/nano-midea/backend/internal/repositories"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// StoryHandler handles story-related HTTP requests
type StoryHandler struct {
	storyRepository repositories.StoryRepository
	userRepository  repositories.UserRepository
}

// NewStoryHandler creates a new StoryHandler
func NewStoryHandler(storyRepo repositories.StoryRepository, userRepo repositories.UserRepository) *StoryHandler {
	return &StoryHandler{
		storyRepository: storyRepo,
		userRepository:  userRepo,
	}
}

// RegisterStoryRoutes registers story-related routes
func (h *StoryHandler) RegisterStoryRoutes(g *echo.Group) {
	g.GET("/stories", h.GetStories)
	g.GET("/stories/:id", h.GetStory)
	g.POST("/stories", h.CreateStory)
	g.POST("/stories/:id/seen", h.MarkAsSeen)
	g.POST("/stories/:id/react", h.ReactToStory)
}

// StoryResponse is the enriched story response
type StoryResponse struct {
	ID             string             `json:"id"`
	Author         models.UserCompact `json:"author"`
	Items          []models.StoryItem `json:"items"`
	HasUnseenItems bool               `json:"has_unseen_items"`
	ExpiresAt      string             `json:"expires_at"`
}

// GetStories returns active stories
func (h *StoryHandler) GetStories(c echo.Context) error {
	currentUserID := getUserIDFromContext(c)

	stories, err := h.storyRepository.GetActiveStories(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Build user map
	userMap := make(map[string]models.UserCompact)
	storyIDs := make([]string, len(stories))
	for i, s := range stories {
		storyIDs[i] = s.ID.Hex()
		if _, ok := userMap[s.UserID]; !ok {
			user, err := h.userRepository.GetUserByFirebaseUID(s.UserID)
			if err == nil {
				userMap[s.UserID] = user.ToCompact()
			} else {
				if id, parseErr := strconv.ParseUint(s.UserID, 10, 32); parseErr == nil {
					user, err := h.userRepository.GetUserByID(uint(id))
					if err == nil {
						userMap[s.UserID] = user.ToCompact()
					}
				}
			}
		}
	}

	// Check seen status
	seenMap := make(map[string]bool)
	if currentUserID > 0 {
		seenMap, _ = h.storyRepository.GetSeenStoryIDs(currentUserID, storyIDs)
	}

	// Build response
	var currentUserStory *StoryResponse
	otherStories := make([]StoryResponse, 0, len(stories))

	for _, s := range stories {
		resp := StoryResponse{
			ID:             s.ID.Hex(),
			Author:         userMap[s.UserID],
			Items:          s.Items,
			HasUnseenItems: !seenMap[s.ID.Hex()],
			ExpiresAt:      s.ExpiresAt.Format(time.RFC3339),
		}

		// Check if this is current user's story
		if currentUserID > 0 {
			author := userMap[s.UserID]
			if author.ID == currentUserID {
				currentUserStory = &resp
				continue
			}
		}
		otherStories = append(otherStories, resp)
	}

	return c.JSON(http.StatusOK, echo.Map{
		"success": true,
		"data": echo.Map{
			"stories":          otherStories,
			"currentUserStory": currentUserStory,
		},
	})
}

// GetStory returns a single story
func (h *StoryHandler) GetStory(c echo.Context) error {
	storyID := c.Param("id")

	story, err := h.storyRepository.GetStoryByID(c.Request().Context(), storyID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Story not found")
	}

	// Get author info
	var author models.UserCompact
	user, err := h.userRepository.GetUserByFirebaseUID(story.UserID)
	if err == nil {
		author = user.ToCompact()
	}

	resp := StoryResponse{
		ID:             story.ID.Hex(),
		Author:         author,
		Items:          story.Items,
		HasUnseenItems: true,
		ExpiresAt:      story.ExpiresAt.Format(time.RFC3339),
	}

	return c.JSON(http.StatusOK, echo.Map{"success": true, "data": echo.Map{"story": resp}})
}

// CreateStory creates a new story
func (h *StoryHandler) CreateStory(c echo.Context) error {
	currentUserID := getUserIDFromContext(c)
	if currentUserID == 0 {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	var req models.CreateStoryRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	story := &models.Story{
		UserID: fmt.Sprintf("%d", currentUserID),
		Items: []models.StoryItem{
			{
				ID:        fmt.Sprintf("item_%d", time.Now().UnixNano()),
				Type:      req.Type,
				URL:       req.MediaURL,
				Duration:  5,
				CreatedAt: time.Now(),
			},
		},
	}

	if err := h.storyRepository.CreateStory(c.Request().Context(), story); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, echo.Map{"success": true, "data": echo.Map{"story": story}})
}

// MarkAsSeen marks a story as seen
func (h *StoryHandler) MarkAsSeen(c echo.Context) error {
	currentUserID := getUserIDFromContext(c)
	if currentUserID == 0 {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	storyID := c.Param("id")

	// Check if already seen
	hasSeen, _ := h.storyRepository.HasSeen(storyID, currentUserID)
	if hasSeen {
		return c.JSON(http.StatusOK, echo.Map{"success": true, "data": echo.Map{"success": true}})
	}

	storySeen := &models.StorySeen{
		StoryID: storyID,
		UserID:  currentUserID,
	}

	if err := h.storyRepository.MarkSeen(storySeen); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, echo.Map{"success": true, "data": echo.Map{"success": true}})
}

// ReactToStory adds a reaction to a story
func (h *StoryHandler) ReactToStory(c echo.Context) error {
	currentUserID := getUserIDFromContext(c)
	if currentUserID == 0 {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	storyID := c.Param("id")

	var req struct {
		Reaction string `json:"reaction" validate:"required"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	reaction := &models.StoryReaction{
		StoryID:  storyID,
		UserID:   currentUserID,
		Reaction: req.Reaction,
	}

	if err := h.storyRepository.AddReaction(reaction); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, echo.Map{"success": true, "data": echo.Map{"success": true}})
}
