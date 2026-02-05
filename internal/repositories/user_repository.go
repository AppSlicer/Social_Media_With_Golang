package repositories

import (
	"github.com/anonto42/nano-midea/backend/internal/models"
	"gorm.io/gorm"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	CreateUser(user *models.User) error
	GetUserByID(id uint) (*models.User, error) // Changed from int to uint
	GetUserByFirebaseUID(firebaseUID string) (*models.User, error)
	GetUsers() ([]models.User, error)
	UpdateUser(user *models.User) error
	DeleteUser(id uint) error
	SearchUsers(query string) ([]models.User, error)
}

// PostgresUserRepository implements UserRepository for PostgreSQL
type PostgresUserRepository struct {
	db *gorm.DB
}

// NewPostgresUserRepository creates a new PostgresUserRepository
func NewPostgresUserRepository(db *gorm.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

// CreateUser creates a new user in PostgreSQL
func (r *PostgresUserRepository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

// GetUserByID retrieves a user by ID from PostgreSQL
func (r *PostgresUserRepository) GetUserByID(id uint) (*models.User, error) { // Changed from int to uint
	var user models.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByFirebaseUID retrieves a user by Firebase UID from PostgreSQL
func (r *PostgresUserRepository) GetUserByFirebaseUID(firebaseUID string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("firebase_uid = ?", firebaseUID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUsers retrieves all users from PostgreSQL
func (r *PostgresUserRepository) GetUsers() ([]models.User, error) {
	var users []models.User
	if err := r.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// UpdateUser updates an existing user in PostgreSQL
func (r *PostgresUserRepository) UpdateUser(user *models.User) error {
	return r.db.Save(user).Error
}

// DeleteUser deletes a user by ID from PostgreSQL
func (r *PostgresUserRepository) DeleteUser(id uint) error { // Changed from int to uint
	return r.db.Delete(&models.User{}, id).Error
}

// SearchUsers searches for users by name or email
func (r *PostgresUserRepository) SearchUsers(query string) ([]models.User, error) {
	var users []models.User
	// Search by name or email (case-insensitive)
	if err := r.db.Where("LOWER(name) LIKE LOWER(?) OR LOWER(email) LIKE LOWER(?)", "%"+query+"%", "%"+query+"%").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
