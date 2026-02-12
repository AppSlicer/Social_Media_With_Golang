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

// UserHandler handles HTTP requests related to users
type UserHandler struct {
	userRepository repositories.UserRepository
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(userRepo repositories.UserRepository) *UserHandler {
	return &UserHandler{userRepository: userRepo}
}

// RegisterProfileRoutes registers user profile-related routes
func (h *UserHandler) RegisterProfileRoutes(g *echo.Group) {
	g.GET("/profile", h.GetProfile)
	g.PUT("/profile", h.UpdateProfile)
	g.GET("/users/suggested", h.GetSuggestedUsers)
	g.GET("/users/:id", h.GetUser)
	g.DELETE("/profile", h.DeleteUser)
}

func (h *UserHandler) GetUser(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID")
	}
	user, err := h.userRepository.GetUserByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "User profile not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"success": true, "data": echo.Map{"user": user}})
}

// GetProfile retrieves the authenticated user's profile
func (h *UserHandler) GetProfile(c echo.Context) error {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	user, err := h.userRepository.GetUserByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "User profile not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, echo.Map{"success": true, "data": echo.Map{"user": user}})
}

// UpdateProfile updates the authenticated user's profile
func (h *UserHandler) UpdateProfile(c echo.Context) error {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	var req models.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	user, err := h.userRepository.GetUserByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "User profile not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if req.DisplayName != "" {
		user.DisplayName = req.DisplayName
	}
	if req.Username != "" {
		user.Username = req.Username
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Bio != "" {
		user.Bio = req.Bio
	}
	if req.AvatarURL != "" {
		user.AvatarURL = req.AvatarURL
	}
	if req.IsPrivate != nil {
		user.IsPrivate = *req.IsPrivate
	}

	if err := h.userRepository.UpdateUser(user); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, echo.Map{"success": true, "data": echo.Map{"user": user}})
}

// DeleteUser deletes the authenticated user's profile
func (h *UserHandler) DeleteUser(c echo.Context) error {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	if err := h.userRepository.DeleteUser(userID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "User profile not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

// SearchUsers searches for users by a query string
func (h *UserHandler) SearchUsers(c echo.Context) error {
	query := c.QueryParam("q")
	if query == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Search query 'q' is required")
	}

	users, err := h.userRepository.SearchUsers(query)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	compact := make([]models.UserCompact, len(users))
	for i, u := range users {
		compact[i] = u.ToCompact()
	}

	return c.JSON(http.StatusOK, echo.Map{"success": true, "data": echo.Map{"users": compact}})
}

// GetSuggestedUsers returns suggested users to follow
func (h *UserHandler) GetSuggestedUsers(c echo.Context) error {
	users, err := h.userRepository.GetUsers()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	limit := 10
	if len(users) < limit {
		limit = len(users)
	}

	compact := make([]models.UserCompact, limit)
	for i := 0; i < limit; i++ {
		compact[i] = users[i].ToCompact()
	}

	return c.JSON(http.StatusOK, echo.Map{"success": true, "data": echo.Map{"users": compact}})
}

// getUserIDFromContext extracts user ID from JWT context
func getUserIDFromContext(c echo.Context) uint {
	if claims, ok := c.Get("user_claims").(*models.JwtCustomClaims); ok {
		return claims.UserID
	}
	if uid, ok := c.Get("user_id").(uint); ok {
		return uid
	}
	return 0
}
