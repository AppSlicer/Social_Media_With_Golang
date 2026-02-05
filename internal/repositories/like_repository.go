package repositories

import (
	"fmt"

	"github.com/anonto42/nano-midea/backend/internal/models"
	"gorm.io/gorm"
)

// LikeRepository defines the interface for like data operations
type LikeRepository interface {
	CreateLike(like *models.Like) error
	DeleteLike(postID string, userID uint) error
	GetLike(postID string, userID uint) (*models.Like, error)
	GetLikesByPostID(postID string) ([]models.Like, error)
	GetLikesCountByPostID(postID string) (int64, error)
	HasUserLikedPost(postID string, userID uint) (bool, error)
}

// PostgresLikeRepository implements LikeRepository for PostgreSQL
type PostgresLikeRepository struct {
	db *gorm.DB
}

// NewPostgresLikeRepository creates a new PostgresLikeRepository
func NewPostgresLikeRepository(db *gorm.DB) *PostgresLikeRepository {
	return &PostgresLikeRepository{db: db}
}

// CreateLike creates a new like in PostgreSQL
func (r *PostgresLikeRepository) CreateLike(like *models.Like) error {
	return r.db.Create(like).Error
}

// DeleteLike deletes a like from PostgreSQL
func (r *PostgresLikeRepository) DeleteLike(postID string, userID uint) error {
	res := r.db.Where("post_id = ? AND user_id = ?", postID, userID).Delete(&models.Like{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("like not found")
	}
	return nil
}

// GetLike retrieves a specific like by postID and userID
func (r *PostgresLikeRepository) GetLike(postID string, userID uint) (*models.Like, error) {
	var like models.Like
	if err := r.db.Where("post_id = ? AND user_id = ?", postID, userID).First(&like).Error; err != nil {
		return nil, err
	}
	return &like, nil
}

// GetLikesByPostID retrieves all likes for a specific post from PostgreSQL
func (r *PostgresLikeRepository) GetLikesByPostID(postID string) ([]models.Like, error) {
	var likes []models.Like
	if err := r.db.Where("post_id = ?", postID).Find(&likes).Error; err != nil {
		return nil, err
	}
	return likes, nil
}

// GetLikesCountByPostID retrieves the count of likes for a specific post from PostgreSQL
func (r *PostgresLikeRepository) GetLikesCountByPostID(postID string) (int64, error) {
	var count int64
	if err := r.db.Model(&models.Like{}).Where("post_id = ?", postID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// HasUserLikedPost checks if a user has liked a specific post
func (r *PostgresLikeRepository) HasUserLikedPost(postID string, userID uint) (bool, error) {
	var count int64
	if err := r.db.Model(&models.Like{}).Where("post_id = ? AND user_id = ?", postID, userID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
