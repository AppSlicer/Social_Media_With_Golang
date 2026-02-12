package repositories

import (
	"fmt"

	"github.com/anonto42/nano-midea/backend/internal/models"
	"gorm.io/gorm"
)

// CommentLikeRepository defines the interface for comment like operations
type CommentLikeRepository interface {
	CreateCommentLike(like *models.CommentLike) error
	DeleteCommentLike(commentID, userID uint) error
	HasUserLikedComment(commentID, userID uint) (bool, error)
	GetLikesCount(commentID uint) (int64, error)
}

type postgresCommentLikeRepository struct {
	db *gorm.DB
}

func NewPostgresCommentLikeRepository(db *gorm.DB) CommentLikeRepository {
	return &postgresCommentLikeRepository{db: db}
}

func (r *postgresCommentLikeRepository) CreateCommentLike(like *models.CommentLike) error {
	return r.db.Create(like).Error
}

func (r *postgresCommentLikeRepository) DeleteCommentLike(commentID, userID uint) error {
	res := r.db.Where("comment_id = ? AND user_id = ?", commentID, userID).Delete(&models.CommentLike{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("comment like not found")
	}
	return nil
}

func (r *postgresCommentLikeRepository) HasUserLikedComment(commentID, userID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.CommentLike{}).Where("comment_id = ? AND user_id = ?", commentID, userID).Count(&count).Error
	return count > 0, err
}

func (r *postgresCommentLikeRepository) GetLikesCount(commentID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.CommentLike{}).Where("comment_id = ?", commentID).Count(&count).Error
	return count, err
}
