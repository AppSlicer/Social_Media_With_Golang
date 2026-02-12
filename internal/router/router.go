package router

import (
	"log"

	"firebase.google.com/go/v4/auth"
	"github.com/anonto42/nano-midea/backend/internal/handlers"
	"github.com/anonto42/nano-midea/backend/internal/middleware"
	"github.com/anonto42/nano-midea/backend/internal/models"
	"github.com/anonto42/nano-midea/backend/internal/repositories"
	"github.com/labstack/echo/v4"
	eMiddleware "github.com/labstack/echo/v4/middleware"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

// SetupMiddleware configures global Echo middleware
func SetupMiddleware(e *echo.Echo) {
	e.Use(eMiddleware.Recover())
	e.Use(eMiddleware.CORS())
	log.Println("Global middleware configured.")
}

// SetupRoutes configures all application routes and injects dependencies
func SetupRoutes(e *echo.Echo, pgdb *gorm.DB, mgClient *mongo.Client, firebaseAuthClient *auth.Client) {
	// AutoMigrate PostgreSQL models
	err := pgdb.AutoMigrate(
		&models.User{},
		&models.FriendRequest{},
		&models.Comment{},
		&models.Like{},
		&models.Follow{},
		&models.SavedPost{},
		&models.StorySeen{},
		&models.StoryReaction{},
		&models.Notification{},
		&models.CommentLike{},
	)
	if err != nil {
		log.Fatalf("Failed to auto migrate models: %v", err)
	}
	log.Println("PostgreSQL auto-migrations completed for all models.")

	// Health check - always accessible
	e.GET("/health", handlers.HealthCheck)
	e.GET("/", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"message": "Hello, World!"})
	})

	// --- Initialize Repositories ---
	userRepo := repositories.NewPostgresUserRepository(pgdb)
	postRepo := repositories.NewMongoPostRepository(mgClient.Database("socialmedia"))
	commentRepo := repositories.NewPostgresCommentRepository(pgdb)
	likeRepo := repositories.NewPostgresLikeRepository(pgdb)
	friendshipRepo := repositories.NewPostgresFriendshipRepository(pgdb)
	followRepo := repositories.NewPostgresFollowRepository(pgdb)
	savedPostRepo := repositories.NewPostgresSavedPostRepository(pgdb)
	storyRepo := repositories.NewStoryRepository(mgClient.Database("socialmedia"), pgdb)
	notificationRepo := repositories.NewPostgresNotificationRepository(pgdb)
	commentLikeRepo := repositories.NewPostgresCommentLikeRepository(pgdb)

	// --- Unprotected routes for authentication ---
	authGroup := e.Group("/api/v1/auth")
	authHandler := handlers.NewAuthHandler(userRepo, firebaseAuthClient)
	authHandler.RegisterAuthRoutes(authGroup)
	log.Println("Auth routes configured.")

	// --- Protected routes (require JWT authentication) ---
	api := e.Group("/api/v1")
	api.Use(middleware.JWTAuthMiddleware())
	log.Println("JWT authentication middleware applied to /api/v1 group.")

	// User profile routes
	userHandler := handlers.NewUserHandler(userRepo)
	userHandler.RegisterProfileRoutes(api)
	api.GET("/users/search", userHandler.SearchUsers)
	log.Println("User profile routes configured.")

	// Post routes
	postHandler := handlers.NewPostHandler(postRepo, userRepo)
	postHandler.RegisterPostRoutes(api)
	log.Println("Post routes configured.")

	// Feed routes
	feedHandler := handlers.NewFeedHandler(postRepo, userRepo, followRepo, likeRepo, savedPostRepo)
	feedHandler.RegisterFeedRoutes(api)
	log.Println("Feed routes configured.")

	// Follow routes
	followHandler := handlers.NewFollowHandler(followRepo, userRepo, notificationRepo)
	followHandler.RegisterFollowRoutes(api)
	log.Println("Follow routes configured.")

	// Friendship routes (legacy)
	friendshipHandler := handlers.NewFriendshipHandler(friendshipRepo, userRepo)
	friendshipHandler.RegisterFriendshipRoutes(api)
	log.Println("Friendship routes configured.")

	// Comment routes
	commentHandler := handlers.NewCommentHandler(commentRepo, postRepo, userRepo, commentLikeRepo, notificationRepo)
	commentHandler.RegisterCommentRoutes(api)
	log.Println("Comment routes configured.")

	// Like routes
	likeHandler := handlers.NewLikeHandler(likeRepo, postRepo, userRepo, notificationRepo)
	likeHandler.RegisterLikeRoutes(api)
	log.Println("Like routes configured.")

	// Saved post routes
	savedPostHandler := handlers.NewSavedPostHandler(savedPostRepo, postRepo)
	savedPostHandler.RegisterSavedPostRoutes(api)
	log.Println("Saved post routes configured.")

	// Story routes
	storyHandler := handlers.NewStoryHandler(storyRepo, userRepo)
	storyHandler.RegisterStoryRoutes(api)
	log.Println("Story routes configured.")

	// Notification routes
	notificationHandler := handlers.NewNotificationHandler(notificationRepo, userRepo)
	notificationHandler.RegisterNotificationRoutes(api)
	log.Println("Notification routes configured.")

	log.Println("All routes configured.")
}
