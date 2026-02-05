package firebase

import (
	"context"
	"fmt"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

// App holds the initialized Firebase app and auth client
type App struct {
	FirebaseApp  *firebase.App
	AuthClient *auth.Client
}

// InitFirebase initializes the Firebase application and authentication client
func InitFirebase(ctx context.Context, credentialsPath string) (*App, error) {
	if credentialsPath == "" {
		return nil, fmt.Errorf("Firebase credentials path not provided")
	}

	// Check if the credentials file exists
	if _, err := os.Stat(credentialsPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("Firebase credentials file not found at %s", credentialsPath)
	}

	opt := option.WithCredentialsFile(credentialsPath)
	
	firebaseApp, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing firebase app: %w", err)
	}

	authClient, err := firebaseApp.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting firebase auth client: %w", err)
	}

	log.Println("Firebase app and auth client initialized successfully!")
	return &App{FirebaseApp: firebaseApp, AuthClient: authClient}, nil
}
