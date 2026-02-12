package repositories

import (
	"github.com/anonto42/nano-midea/backend/internal/models"
	"gorm.io/gorm"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	CreateUser(user *models.User) error
	GetUserByID(id uint) (*models.User, error)
	GetUserByFirebaseUID(firebaseUID string) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	GetUsers() ([]models.User, error)
	GetUsersByIDs(ids []uint) ([]models.User, error)
	UpdateUser(user *models.User) error
	DeleteUser(id uint) error
	SearchUsers(query string) ([]models.User, error)
	IncrementFollowersCount(userID uint)
	DecrementFollowersCount(userID uint)
	IncrementFollowingCount(userID uint)
	DecrementFollowingCount(userID uint)
	IncrementPostsCount(userID uint)
	DecrementPostsCount(userID uint)
}

// PostgresUserRepository implements UserRepository for PostgreSQL
type PostgresUserRepository struct {
	db *gorm.DB
}

// NewPostgresUserRepository creates a new PostgresUserRepository
func NewPostgresUserRepository(db *gorm.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *PostgresUserRepository) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *PostgresUserRepository) GetUserByFirebaseUID(firebaseUID string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("firebase_uid = ?", firebaseUID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *PostgresUserRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *PostgresUserRepository) GetUsers() ([]models.User, error) {
	var users []models.User
	if err := r.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *PostgresUserRepository) GetUsersByIDs(ids []uint) ([]models.User, error) {
	var users []models.User
	if len(ids) == 0 {
		return users, nil
	}
	if err := r.db.Where("id IN ?", ids).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *PostgresUserRepository) UpdateUser(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *PostgresUserRepository) DeleteUser(id uint) error {
	return r.db.Delete(&models.User{}, id).Error
}

func (r *PostgresUserRepository) SearchUsers(query string) ([]models.User, error) {
	var users []models.User
	if err := r.db.Where("LOWER(display_name) LIKE LOWER(?) OR LOWER(username) LIKE LOWER(?) OR LOWER(email) LIKE LOWER(?)",
		"%"+query+"%", "%"+query+"%", "%"+query+"%").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (r *PostgresUserRepository) IncrementFollowersCount(userID uint) {
	r.db.Model(&models.User{}).Where("id = ?", userID).UpdateColumn("followers_count", gorm.Expr("followers_count + 1"))
}

func (r *PostgresUserRepository) DecrementFollowersCount(userID uint) {
	r.db.Model(&models.User{}).Where("id = ? AND followers_count > 0", userID).UpdateColumn("followers_count", gorm.Expr("followers_count - 1"))
}

func (r *PostgresUserRepository) IncrementFollowingCount(userID uint) {
	r.db.Model(&models.User{}).Where("id = ?", userID).UpdateColumn("following_count", gorm.Expr("following_count + 1"))
}

func (r *PostgresUserRepository) DecrementFollowingCount(userID uint) {
	r.db.Model(&models.User{}).Where("id = ? AND following_count > 0", userID).UpdateColumn("following_count", gorm.Expr("following_count - 1"))
}

func (r *PostgresUserRepository) IncrementPostsCount(userID uint) {
	r.db.Model(&models.User{}).Where("id = ?", userID).UpdateColumn("posts_count", gorm.Expr("posts_count + 1"))
}

func (r *PostgresUserRepository) DecrementPostsCount(userID uint) {
	r.db.Model(&models.User{}).Where("id = ? AND posts_count > 0", userID).UpdateColumn("posts_count", gorm.Expr("posts_count - 1"))
}
