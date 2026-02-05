package repositories

import (
	"github.com/anonto42/nano-midea/backend/internal/models"
	"gorm.io/gorm"
)

// CommentRepository defines the interface for comment data operations
type CommentRepository interface {
	CreateComment(comment *models.Comment) error
	GetCommentByID(id uint) (*models.Comment, error)
	GetCommentsByPostID(postID string) ([]models.Comment, error)
	UpdateComment(comment *models.Comment) error
	DeleteComment(id uint) error
}

// PostgresCommentRepository implements CommentRepository for PostgreSQL
type PostgresCommentRepository struct {
	db *gorm.DB
}

// NewPostgresCommentRepository creates a new PostgresCommentRepository
func NewPostgresCommentRepository(db *gorm.DB) *PostgresCommentRepository {
	return &PostgresCommentRepository{db: db}
}

// CreateComment creates a new comment in PostgreSQL
func (r *PostgresCommentRepository) CreateComment(comment *models.Comment) error {
	return r.db.Create(comment).Error
}

// GetCommentByID retrieves a comment by ID from PostgreSQL
func (r *PostgresCommentRepository) GetCommentByID(id uint) (*models.Comment, error) {
	var comment models.Comment
	if err := r.db.First(&comment, id).Error; err != nil {
		return nil, err
	}
	return &comment, nil
}

// GetCommentsByPostID retrieves all comments for a specific post from PostgreSQL
func (r *PostgresCommentRepository) GetCommentsByPostID(postID string) ([]models.Comment, error) {
	var comments []models.Comment
	if err := r.db.Where("post_id = ?", postID).Find(&comments).Error; err != nil {
		return nil, err
	}
	return comments, nil
}

// UpdateComment updates an existing comment in PostgreSQL
func (r *PostgresCommentRepository) UpdateComment(comment *models.Comment) error {
	return r.db.Save(comment).Error
}

// DeleteComment deletes a comment by ID from PostgreSQL
func (r *PostgresCommentRepository) DeleteComment(id uint) error {
	return r.db.Delete(&models.Comment{}, id).Error
}
