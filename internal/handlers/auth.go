package handlers

import (
	"context"
	"net/http"
	"os"
	"time"

	"firebase.google.com/go/v4/auth"
	"github.com/anonto42/nano-midea/backend/internal/models"
	"github.com/anonto42/nano-midea/backend/internal/repositories"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	userRepository repositories.UserRepository
	firebaseAuth   *auth.Client
	jwtSecret      string
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(userRepo repositories.UserRepository, firebaseAuthClient *auth.Client) *AuthHandler {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "supersecretjwtkey"
	}
	return &AuthHandler{
		userRepository: userRepo,
		firebaseAuth:   firebaseAuthClient,
		jwtSecret:      jwtSecret,
	}
}

// RegisterAuthRoutes registers authentication-related routes
func (h *AuthHandler) RegisterAuthRoutes(g *echo.Group) {
	g.POST("/register", h.Register)
	g.POST("/signup", h.Signup)
	g.POST("/signin", h.SignIn)
	g.POST("/firebase-login", h.FirebaseLogin)
}

// Register handles user registration with Firebase UID
func (h *AuthHandler) Register(c echo.Context) error {
	var req models.CreateUserRequest

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	_, err := h.userRepository.GetUserByFirebaseUID(req.FirebaseUID)
	if err == nil {
		return echo.NewHTTPError(http.StatusConflict, "User with this Firebase UID already registered")
	}

	user := &models.User{
		DisplayName: req.Name,
		Email:       req.Email,
		Age:         req.Age,
		FirebaseUID: req.FirebaseUID,
	}

	if err := h.userRepository.CreateUser(user); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, echo.Map{"user": user})
}

// Signup handles local user registration with email and password
func (h *AuthHandler) Signup(c echo.Context) error {
	var req models.CreateLocalUserRequest

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	_, err := h.userRepository.GetUserByEmail(req.Email)
	if err == nil {
		return echo.NewHTTPError(http.StatusConflict, "User with this email already registered")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to hash password")
	}

	user := &models.User{
		DisplayName: req.Name,
		Username:    req.Username,
		Email:       req.Email,
		Password:    string(hashedPassword),
	}

	if err := h.userRepository.CreateUser(user); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	token, err := h.generateJWT(user)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate token after signup")
	}

	return c.JSON(http.StatusCreated, echo.Map{"token": token, "user": user})
}

// SignIn handles local user authentication with email and password
func (h *AuthHandler) SignIn(c echo.Context) error {
	var req struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	user, err := h.userRepository.GetUserByEmail(req.Email)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not found with email: "+req.Email)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid password")
	}

	token, err := h.generateJWT(user)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate token")
	}

	return c.JSON(http.StatusOK, echo.Map{"token": token, "user": user})
}

// FirebaseLoginRequest defines the request body for Firebase login
type FirebaseLoginRequest struct {
	IDToken string `json:"idToken" validate:"required"`
}

// FirebaseLogin handles Firebase ID token verification and issues a local JWT
func (h *AuthHandler) FirebaseLogin(c echo.Context) error {
	var req FirebaseLoginRequest

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	token, err := h.firebaseAuth.VerifyIDToken(context.Background(), req.IDToken)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid Firebase ID token")
	}

	firebaseUID := token.UID
	email := token.Claims["email"].(string)
	name := ""
	if displayName, ok := token.Claims["name"].(string); ok {
		name = displayName
	}

	user, err := h.userRepository.GetUserByFirebaseUID(firebaseUID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			user, err = h.userRepository.GetUserByEmail(email)
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					newUser := &models.User{
						DisplayName: name,
						Username:    firebaseUID,
						Email:       email,
						FirebaseUID: firebaseUID,
					}
					if err := h.userRepository.CreateUser(newUser); err != nil {
						return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create user")
					}
					user = newUser
				} else {
					return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
				}
			} else {
				user.FirebaseUID = firebaseUID
				if err := h.userRepository.UpdateUser(user); err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update user with Firebase UID")
				}
			}
		} else {
			return echo.NewHTTPError(http.StatusInternalServerError, "Database error")
		}
	} else {
		user.Email = email
		if name != "" {
			user.DisplayName = name
		}
		if err := h.userRepository.UpdateUser(user); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update user details")
		}
	}

	localJWT, err := h.generateJWT(user)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate local JWT")
	}

	return c.JSON(http.StatusOK, echo.Map{"token": localJWT, "user": user})
}

// generateJWT generates a JWT token for a given user
func (h *AuthHandler) generateJWT(user *models.User) (string, error) {
	claims := &models.JwtCustomClaims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 72)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		return "", err
	}
	return t, nil
}
