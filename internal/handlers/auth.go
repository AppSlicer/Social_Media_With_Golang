package handlers

import (
	"net/http"

	"firebase.google.com/go/v4/auth"
	"github.com/anonto42/nano-midea/backend/internal/models"
	"github.com/anonto42/nano-midea/backend/internal/repositories"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	userRepository repositories.UserRepository
	firebaseAuth   *auth.Client
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(userRepo repositories.UserRepository, firebaseAuthClient *auth.Client) *AuthHandler {
	return &AuthHandler{
		userRepository: userRepo,
		firebaseAuth:   firebaseAuthClient,
	}
}

// RegisterAuthRoutes registers authentication-related routes
func (h *AuthHandler) RegisterAuthRoutes(g *echo.Group) {
	g.POST("/register", h.Register)
	// Login is typically handled client-side with Firebase SDK,
	// or if needed, a custom token endpoint can be added here.
	// g.POST("/login", h.Login)
}

// Register handles user registration
func (h *AuthHandler) Register(c echo.Context) error {
	var req models.CreateUserRequest

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Verify the Firebase UID
	// This assumes the client has already authenticated with Firebase and sent the ID token,
	// from which the FirebaseUID is extracted and included in the request.
	// Alternatively, the client sends the ID token, and the backend verifies it to get the UID.
	// For this flow, we assume the FirebaseUID is provided directly in the request for simplicity,
	// but a more robust solution would involve verifying the ID token sent in the Authorization header
	// during registration itself (if not using the general FirebaseAuthMiddleware for this specific endpoint).

	// Check if user with this Firebase UID already exists in our DB
	_, err := h.userRepository.GetUserByFirebaseUID(req.FirebaseUID)
	if err == nil {
		return echo.NewHTTPError(http.StatusConflict, "User with this Firebase UID already registered")
	}

	user := &models.User{
		Name:        req.Name,
		Email:       req.Email,
		Age:         req.Age,
		FirebaseUID: req.FirebaseUID,
	}

	if err := h.userRepository.CreateUser(user); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, user)
}
