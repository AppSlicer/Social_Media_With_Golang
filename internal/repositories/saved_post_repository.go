package repositories

import (
	"fmt"

	"github.com/anonto42/nano-midea/backend/internal/models"
	"gorm.io/gorm"
)

// SavedPostRepository defines the interface for saved post operations
type SavedPostRepository interface {
	SavePost(savedPost *models.SavedPost) error
	UnsavePost(userID uint, postID string) error
	IsPostSaved(userID uint, postID string) (bool, error)
	GetSavedPostsByUser(userID uint) ([]models.SavedPost, error)
	GetSavedPostIDs(userID uint, postIDs []string) (map[string]bool, error)
}

// PostgresSavedPostRepository implements SavedPostRepository
type PostgresSavedPostRepository struct {
	db *gorm.DB
}

func NewPostgresSavedPostRepository(db *gorm.DB) *PostgresSavedPostRepository {
	return &PostgresSavedPostRepository{db: db}
}

func (r *PostgresSavedPostRepository) SavePost(savedPost *models.SavedPost) error {
	return r.db.Create(savedPost).Error
}

func (r *PostgresSavedPostRepository) UnsavePost(userID uint, postID string) error {
	res := r.db.Where("user_id = ? AND post_id = ?", userID, postID).Delete(&models.SavedPost{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("saved post not found")
	}
	return nil
}

func (r *PostgresSavedPostRepository) IsPostSaved(userID uint, postID string) (bool, error) {
	var count int64
	err := r.db.Model(&models.SavedPost{}).Where("user_id = ? AND post_id = ?", userID, postID).Count(&count).Error
	return count > 0, err
}

func (r *PostgresSavedPostRepository) GetSavedPostsByUser(userID uint) ([]models.SavedPost, error) {
	var saved []models.SavedPost
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&saved).Error
	return saved, err
}

func (r *PostgresSavedPostRepository) GetSavedPostIDs(userID uint, postIDs []string) (map[string]bool, error) {
	result := make(map[string]bool)
	if len(postIDs) == 0 {
		return result, nil
	}
	var saved []models.SavedPost
	err := r.db.Where("user_id = ? AND post_id IN ?", userID, postIDs).Find(&saved).Error
	if err != nil {
		return nil, err
	}
	for _, s := range saved {
		result[s.PostID] = true
	}
	return result, nil
}
