package main

import (
	"context"
	"log"

	"github.com/anonto42/nano-midea/backend/internal/router"
	"github.com/anonto42/nano-midea/backend/pkg/config"
	"github.com/anonto42/nano-midea/backend/pkg/firebase"
	"github.com/anonto42/nano-midea/backend/validators"
	"github.com/labstack/echo/v4"
)

func main() {
	// Load configuration
	cfg := config.Load()
	
	// Initialize database connections
	db, err := config.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize databases: %v", err)
	}
	defer db.CloseDB() // Ensure database connections are closed when main exits

	// Initialize Firebase
	ctx := context.Background()
	firebaseApp, err := firebase.InitFirebase(ctx, "./firebase_credentials.json")
	if err != nil {
		log.Fatalf("Failed to initialize Firebase: %v", err)
	}

	// Create Echo instance
	e := echo.New()
	
	// Setup global middleware
	router.SetupMiddleware(e)
	
	// Setup routes and dependencies
	router.SetupRoutes(e, db.Postgres, db.Mongo, firebaseApp.AuthClient)

	// Validator
	e.Validator = validators.NewValidator()

	// Start server
	e.Logger.Fatal(e.Start(":" + cfg.Port))
}