package models

import (
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model  `json:"-"`
	ID          uint   `json:"id" gorm:"primaryKey"`
	Name        string `json:"name"`
	Email       string `json:"email" gorm:"uniqueIndex"` // Ensure email is unique across all users
	Age         int    `json:"age"`
	Password    string `json:"-"`                                         // Store hashed password, ignore for JSON serialization
	FirebaseUID string `json:"firebase_uid,omitempty" gorm:"uniqueIndex"` // Link to Firebase User UID
}

type CreateUserRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=50"`
	Email       string `json:"email" validate:"required,email"`
	Age         int    `json:"age" validate:"required,min=0,max=150"`
	FirebaseUID string `json:"firebase_uid" validate:"required"` // Firebase UID will be provided by the client after Firebase Auth
}

type CreateLocalUserRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Age      int    `json:"age" validate:"required,min=0,max=150"`
	Password string `json:"password" validate:"required,min=8"`
}

type UpdateUserRequest struct {
	Name  string `json:"name,omitempty" validate:"omitempty,min=2,max=50"`
	Email string `json:"email,omitempty" validate:"omitempty,email"`
	Age   int    `json:"age,omitempty" validate:"min=0,max=150"`
}

// JwtCustomClaims are custom claims extending standard jwt.RegisteredClaims
type JwtCustomClaims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}
