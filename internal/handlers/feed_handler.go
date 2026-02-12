package handlers

import (
	"math"
	"net/http"
	"strconv"

	"github.com/anonto42/nano-midea/backend/internal/models"
	"github.com/anonto42/nano-midea/backend/internal/repositories"
	"github.com/labstack/echo/v4"
)

// FeedHandler handles feed-related HTTP requests
type FeedHandler struct {
	postRepository      repositories.PostRepository
	userRepository      repositories.UserRepository
	followRepository    repositories.FollowRepository
	likeRepository      repositories.LikeRepository
	savedPostRepository repositories.SavedPostRepository
}

// NewFeedHandler creates a new FeedHandler
func NewFeedHandler(
	postRepo repositories.PostRepository,
	userRepo repositories.UserRepository,
	followRepo repositories.FollowRepository,
	likeRepo repositories.LikeRepository,
	savedPostRepo repositories.SavedPostRepository,
) *FeedHandler {
	return &FeedHandler{
		postRepository:      postRepo,
		userRepository:      userRepo,
		followRepository:    followRepo,
		likeRepository:      likeRepo,
		savedPostRepository: savedPostRepo,
	}
}

// RegisterFeedRoutes registers feed-related routes
func (h *FeedHandler) RegisterFeedRoutes(g *echo.Group) {
	g.GET("/feed", h.GetFeed)
}

// EnrichedPost is a post with author info and user-specific flags
type EnrichedPost struct {
	models.Post
	Author  models.UserCompact `json:"author"`
	IsLiked bool               `json:"is_liked"`
	IsSaved bool               `json:"is_saved"`
}

// GetFeed returns enriched feed posts for the current user
func (h *FeedHandler) GetFeed(c echo.Context) error {
	currentUserID := getUserIDFromContext(c)

	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}

	skip := int64((page - 1) * limit)

	// Get all posts (in a real app, filter by followed users + own)
	posts, err := h.postRepository.GetAllPosts(c.Request().Context(), skip, int64(limit))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// Get total count for pagination
	allPosts, err := h.postRepository.GetAllPosts(c.Request().Context(), 0, 10000)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	totalItems := len(allPosts)

	// Collect unique user IDs from posts (these are Firebase UIDs stored as strings)
	userFirebaseUIDs := make(map[string]bool)
	postIDs := make([]string, len(posts))
	for i, p := range posts {
		userFirebaseUIDs[p.UserID] = true
		postIDs[i] = p.ID.Hex()
	}

	// Build user map by Firebase UID
	userMap := make(map[string]models.UserCompact)
	for uid := range userFirebaseUIDs {
		// Try to find user by Firebase UID
		user, err := h.userRepository.GetUserByFirebaseUID(uid)
		if err == nil {
			userMap[uid] = user.ToCompact()
		} else {
			// Try parsing as uint ID
			if id, parseErr := strconv.ParseUint(uid, 10, 32); parseErr == nil {
				user, err := h.userRepository.GetUserByID(uint(id))
				if err == nil {
					userMap[uid] = user.ToCompact()
				}
			}
		}
	}

	// Check liked status for current user
	likedMap := make(map[string]bool)
	savedMap := make(map[string]bool)
	if currentUserID > 0 {
		for _, pid := range postIDs {
			liked, _ := h.likeRepository.HasUserLikedPost(pid, currentUserID)
			likedMap[pid] = liked
		}
		savedMap, _ = h.savedPostRepository.GetSavedPostIDs(currentUserID, postIDs)
	}

	// Build enriched posts
	enrichedPosts := make([]EnrichedPost, len(posts))
	for i, p := range posts {
		pid := p.ID.Hex()
		enrichedPosts[i] = EnrichedPost{
			Post:    p,
			Author:  userMap[p.UserID],
			IsLiked: likedMap[pid],
			IsSaved: savedMap[pid],
		}
	}

	totalPages := int(math.Ceil(float64(totalItems) / float64(limit)))

	return c.JSON(http.StatusOK, echo.Map{
		"success": true,
		"data": echo.Map{
			"posts": enrichedPosts,
		},
		"meta": echo.Map{
			"currentPage":     page,
			"totalPages":      totalPages,
			"totalItems":      totalItems,
			"itemsPerPage":    limit,
			"hasNextPage":     page < totalPages,
			"hasPreviousPage": page > 1,
		},
	})
}
