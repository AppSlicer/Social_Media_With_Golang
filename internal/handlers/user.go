package handlers

import (
	"net/http"
	"strconv" // Added this import

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
	g.GET("/profile", h.GetProfile)    // Get own profile
	g.PUT("/profile", h.UpdateProfile) // Update own profile
	g.GET("/users/:id", h.GetUser)     // Get other user's profile by ID
	g.DELETE("/profile", h.DeleteUser) // Delete own user profile
}

func (h *UserHandler) GetUser(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32) // Changed to ParseUint
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID")
	}
	user, err := h.userRepository.GetUserByID(uint(id)) // Cast to uint
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "User profile not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, user)
}

// GetProfile retrieves the authenticated user's profile
func (h *UserHandler) GetProfile(c echo.Context) error {
	firebaseUID := c.Get("firebaseUID").(string) // Get Firebase UID from middleware
	
	user, err := h.userRepository.GetUserByFirebaseUID(firebaseUID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "User profile not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, user)
}

// UpdateProfile updates the authenticated user's profile
func (h *UserHandler) UpdateProfile(c echo.Context) error {
	firebaseUID := c.Get("firebaseUID").(string) // Get Firebase UID from middleware

	var req models.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	user, err := h.userRepository.GetUserByFirebaseUID(firebaseUID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "User profile not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Age != 0 {
		user.Age = req.Age
	}

	if err := h.userRepository.UpdateUser(user); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, user)
}

// DeleteUser deletes the authenticated user's profile
func (h *UserHandler) DeleteUser(c echo.Context) error {
	firebaseUID := c.Get("firebaseUID").(string) // Get Firebase UID from middleware

	user, err := h.userRepository.GetUserByFirebaseUID(firebaseUID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "User profile not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if err := h.userRepository.DeleteUser(user.ID); err != nil { // user.ID is already uint
		if err == gorm.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "User profile not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusNoContent)
}

// SearchUsers searches for users by a query string (email or name)
func (h *UserHandler) SearchUsers(c echo.Context) error {
	query := c.QueryParam("q")
	if query == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Search query 'q' is required")
	}

	users, err := h.userRepository.SearchUsers(query)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, users)
}