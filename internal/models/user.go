package models

import "gorm.io/gorm"

type User struct {
	gorm.Model `json:"-"` // Ignore gorm.Model fields for JSON serialization
	ID         uint   `json:"id" gorm:"primaryKey"` // Explicitly define ID for JSON output
	Name       string `json:"name"`
	Email      string `json:"email"`
	Age        int    `json:"age"`
	FirebaseUID string `json:"firebase_uid" gorm:"uniqueIndex"` // Link to Firebase User UID
}

type CreateUserRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=50"`
	Email       string `json:"email" validate:"required,email"`
	Age         int    `json:"age" validate:"required,min=0,max=150"`
	FirebaseUID string `json:"firebase_uid" validate:"required"` // Firebase UID will be provided by the client after Firebase Auth
}

type UpdateUserRequest struct {
	Name  string `json:"name,omitempty" validate:"omitempty,min=2,max=50"`
	Email string `json:"email,omitempty" validate:"omitempty,email"`
	Age   int    `json:"age,omitempty" validate:"min=0,max=150"`
}