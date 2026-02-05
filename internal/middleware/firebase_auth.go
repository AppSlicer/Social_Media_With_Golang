package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"firebase.google.com/go/v4/auth"
	"github.com/labstack/echo/v4"
)

// FirebaseAuthMiddleware creates an Echo middleware to verify Firebase ID tokens
func FirebaseAuthMiddleware(authClient *auth.Client) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Authorization header is missing")
			}

			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Authorization header must be in Bearer format")
			}

			idToken := tokenParts[1]
			
			// Verify the ID token
			token, err := authClient.VerifyIDToken(context.Background(), idToken)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, fmt.Sprintf("Invalid or expired ID token: %v", err))
			}

			// Store the Firebase UID in the context for later use
			c.Set("firebaseUID", token.UID)
			c.Set("firebaseToken", token) // You might want to store the whole token as well

			return next(c)
		}
	}
}
