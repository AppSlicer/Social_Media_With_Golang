package models

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model     `json:"-"`
	ID             uint      `json:"id" gorm:"primaryKey"`
	Username       string    `json:"username" gorm:"uniqueIndex;size:30"`
	DisplayName    string    `json:"display_name"`
	Email          string    `json:"email" gorm:"uniqueIndex"`
	Bio            string    `json:"bio" gorm:"size:150"`
	AvatarURL      string    `json:"avatar_url"`
	IsVerified     bool      `json:"is_verified" gorm:"default:false"`
	IsPrivate      bool      `json:"is_private" gorm:"default:false"`
	FollowersCount int       `json:"followers_count" gorm:"default:0"`
	FollowingCount int       `json:"following_count" gorm:"default:0"`
	PostsCount     int       `json:"posts_count" gorm:"default:0"`
	Age            int       `json:"age"`
	Password       string    `json:"-"`
	FirebaseUID    string    `json:"firebase_uid,omitempty" gorm:"uniqueIndex"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// UserCompact is a lightweight user representation for lists
type UserCompact struct {
	ID          uint   `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
	IsVerified  bool   `json:"is_verified"`
}

// ToCompact converts a User to UserCompact
func (u *User) ToCompact() UserCompact {
	return UserCompact{
		ID:          u.ID,
		Username:    u.Username,
		DisplayName: u.DisplayName,
		AvatarURL:   u.AvatarURL,
		IsVerified:  u.IsVerified,
	}
}

type CreateUserRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=50"`
	Email       string `json:"email" validate:"required,email"`
	Age         int    `json:"age" validate:"required,min=0,max=150"`
	FirebaseUID string `json:"firebase_uid" validate:"required"`
}

type CreateLocalUserRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=50"`
	Username string `json:"username" validate:"required,min=3,max=30"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type UpdateUserRequest struct {
	DisplayName string `json:"display_name,omitempty" validate:"omitempty,min=2,max=50"`
	Username    string `json:"username,omitempty" validate:"omitempty,min=3,max=30"`
	Email       string `json:"email,omitempty" validate:"omitempty,email"`
	Bio         string `json:"bio,omitempty" validate:"omitempty,max=150"`
	AvatarURL   string `json:"avatar_url,omitempty"`
	IsPrivate   *bool  `json:"is_private,omitempty"`
}

// JwtCustomClaims are custom claims extending standard jwt.RegisteredClaims
type JwtCustomClaims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}
