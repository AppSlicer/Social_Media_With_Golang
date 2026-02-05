package router

import (
	"log"

	"firebase.google.com/go/v4/auth"
	"github.com/anonto42/nano-midea/backend/internal/handlers"
	"github.com/anonto42/nano-midea/backend/internal/middleware"
	"github.com/anonto42/nano-midea/backend/internal/repositories"
	"github.com/anonto42/nano-midea/backend/internal/models" // Import models for AutoMigrate
	"github.com/labstack/echo/v4"
	eMiddleware "github.com/labstack/echo/v4/middleware"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

// SetupMiddleware configures global Echo middleware
func SetupMiddleware(e *echo.Echo) {
	e.Use(eMiddleware.RequestLogger())
	e.Use(eMiddleware.Recover())
	e.Use(eMiddleware.CORS())
	log.Println("Global middleware configured.")
}

// SetupRoutes configures all application routes and injects dependencies
func SetupRoutes(e *echo.Echo, pgdb *gorm.DB, mgClient *mongo.Client, firebaseAuthClient *auth.Client) {
	// AutoMigrate PostgreSQL models here as it depends on pgdb
	err := pgdb.AutoMigrate(&models.User{}, &models.FriendRequest{}, &models.Comment{}, &models.Like{})
	if err != nil {
		log.Fatalf("Failed to auto migrate models: %v", err)
	}
	log.Println("PostgreSQL auto-migrations completed for all models.")

	// Health check - always accessible
	e.GET("/health", handlers.HealthCheck)

	e.GET("/", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"message": "Hello, World!"})
	})
	
	// --- Unprotected routes for authentication (e.g., /register) ---
	authGroup := e.Group("/api/v1/auth")
	
	// Initialize User Repository once
	userRepo := repositories.NewPostgresUserRepository(pgdb)
	
	authHandler := handlers.NewAuthHandler(userRepo, firebaseAuthClient)
	authHandler.RegisterAuthRoutes(authGroup)
	log.Println("Auth routes configured.")

	// --- Protected routes (require Firebase authentication) ---
	api := e.Group("/api/v1")
	api.Use(middleware.FirebaseAuthMiddleware(firebaseAuthClient))
	log.Println("Firebase authentication middleware applied to /api/v1 group.")
	
	// User profile routes
	userHandler := handlers.NewUserHandler(userRepo)
	userHandler.RegisterProfileRoutes(api)
	api.GET("/users/search", userHandler.SearchUsers) // New search route
	log.Println("User profile routes configured.")

	// Post routes
	postRepo := repositories.NewMongoPostRepository(mgClient.Database("socialmedia")) // Assuming "socialmedia" is your MongoDB database name
	postHandler := handlers.NewPostHandler(postRepo, userRepo)
	postHandler.RegisterPostRoutes(api)
	log.Println("Post routes configured.")

	// Friendship routes
	friendshipRepo := repositories.NewPostgresFriendshipRepository(pgdb)
	friendshipHandler := handlers.NewFriendshipHandler(friendshipRepo, userRepo)
	friendshipHandler.RegisterFriendshipRoutes(api)
	log.Println("Friendship routes configured.")

	// Comment routes
	commentRepo := repositories.NewPostgresCommentRepository(pgdb)
	commentHandler := handlers.NewCommentHandler(commentRepo, postRepo, userRepo)
	commentHandler.RegisterCommentRoutes(api)
	log.Println("Comment routes configured.")

	// Like routes
	likeRepo := repositories.NewPostgresLikeRepository(pgdb)
	likeHandler := handlers.NewLikeHandler(likeRepo, postRepo, userRepo)
	likeHandler.RegisterLikeRoutes(api)
	log.Println("Like routes configured.")

	// --- Protected routes (require JWT authentication) ---
	jwtApi := e.Group("/api/v1/jwt")
	jwtApi.Use(middleware.JWTAuthMiddleware())
	log.Println("JWT authentication middleware applied to /api/v1/jwt group.")

	// Example: Move a route to be protected by JWT
	jwtApi.GET("/users/search", userHandler.SearchUsers) // Now protected by JWTAuthMiddleware

	log.Println("All routes configured.")
}
