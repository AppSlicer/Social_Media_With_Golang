package repositories

import (
	"fmt"

	"github.com/anonto42/nano-midea/backend/internal/models"
	"gorm.io/gorm"
)

// FollowRepository defines the interface for follow data operations
type FollowRepository interface {
	CreateFollow(follow *models.Follow) error
	DeleteFollow(followerID, followingID uint) error
	IsFollowing(followerID, followingID uint) (bool, error)
	GetFollowers(userID uint) ([]models.User, error)
	GetFollowing(userID uint) ([]models.User, error)
	GetFollowersCount(userID uint) (int64, error)
	GetFollowingCount(userID uint) (int64, error)
	GetFollowingIDs(userID uint) ([]uint, error)
}

// PostgresFollowRepository implements FollowRepository for PostgreSQL
type PostgresFollowRepository struct {
	db *gorm.DB
}

// NewPostgresFollowRepository creates a new PostgresFollowRepository
func NewPostgresFollowRepository(db *gorm.DB) *PostgresFollowRepository {
	return &PostgresFollowRepository{db: db}
}

func (r *PostgresFollowRepository) CreateFollow(follow *models.Follow) error {
	return r.db.Create(follow).Error
}

func (r *PostgresFollowRepository) DeleteFollow(followerID, followingID uint) error {
	res := r.db.Where("follower_id = ? AND following_id = ?", followerID, followingID).Delete(&models.Follow{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("follow relationship not found")
	}
	return nil
}

func (r *PostgresFollowRepository) IsFollowing(followerID, followingID uint) (bool, error) {
	var count int64
	if err := r.db.Model(&models.Follow{}).Where("follower_id = ? AND following_id = ?", followerID, followingID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *PostgresFollowRepository) GetFollowers(userID uint) ([]models.User, error) {
	var users []models.User
	err := r.db.Where("id IN (?)",
		r.db.Table("follows").Select("follower_id").Where("following_id = ?", userID),
	).Find(&users).Error
	return users, err
}

func (r *PostgresFollowRepository) GetFollowing(userID uint) ([]models.User, error) {
	var users []models.User
	err := r.db.Where("id IN (?)",
		r.db.Table("follows").Select("following_id").Where("follower_id = ?", userID),
	).Find(&users).Error
	return users, err
}

func (r *PostgresFollowRepository) GetFollowersCount(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.Follow{}).Where("following_id = ?", userID).Count(&count).Error
	return count, err
}

func (r *PostgresFollowRepository) GetFollowingCount(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.Follow{}).Where("follower_id = ?", userID).Count(&count).Error
	return count, err
}

func (r *PostgresFollowRepository) GetFollowingIDs(userID uint) ([]uint, error) {
	var ids []uint
	err := r.db.Model(&models.Follow{}).Where("follower_id = ?", userID).Pluck("following_id", &ids).Error
	return ids, err
}
