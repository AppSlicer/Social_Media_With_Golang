package repositories

import (
	"time"

	"github.com/anonto42/nano-midea/backend/internal/models"
	"gorm.io/gorm"
)

// NotificationRepository defines the interface for notification operations
type NotificationRepository interface {
	CreateNotification(notification *models.Notification) error
	GetByRecipientID(recipientID uint, page, limit int) ([]models.Notification, int64, error)
	GetGrouped(recipientID uint) ([]models.Notification, []models.Notification, []models.Notification, []models.Notification, error)
	GetUnreadCount(recipientID uint) (int64, error)
	MarkAsRead(notificationID uint) error
	MarkAllAsRead(recipientID uint) error
}

type postgresNotificationRepository struct {
	db *gorm.DB
}

func NewPostgresNotificationRepository(db *gorm.DB) NotificationRepository {
	return &postgresNotificationRepository{db: db}
}

func (r *postgresNotificationRepository) CreateNotification(notification *models.Notification) error {
	return r.db.Create(notification).Error
}

func (r *postgresNotificationRepository) GetByRecipientID(recipientID uint, page, limit int) ([]models.Notification, int64, error) {
	var notifications []models.Notification
	var total int64

	r.db.Model(&models.Notification{}).Where("recipient_id = ?", recipientID).Count(&total)

	offset := (page - 1) * limit
	err := r.db.Where("recipient_id = ?", recipientID).
		Order("created_at DESC").
		Offset(offset).Limit(limit).
		Find(&notifications).Error

	return notifications, total, err
}

func (r *postgresNotificationRepository) GetGrouped(recipientID uint) (today, yesterday, thisWeek, older []models.Notification, retErr error) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	yesterdayStart := todayStart.AddDate(0, 0, -1)
	weekStart := todayStart.AddDate(0, 0, -7)

	// Today
	if err := r.db.Where("recipient_id = ? AND created_at >= ?", recipientID, todayStart).
		Order("created_at DESC").Find(&today).Error; err != nil {
		return nil, nil, nil, nil, err
	}

	// Yesterday
	if err := r.db.Where("recipient_id = ? AND created_at >= ? AND created_at < ?", recipientID, yesterdayStart, todayStart).
		Order("created_at DESC").Find(&yesterday).Error; err != nil {
		return nil, nil, nil, nil, err
	}

	// This week (excluding today and yesterday)
	if err := r.db.Where("recipient_id = ? AND created_at >= ? AND created_at < ?", recipientID, weekStart, yesterdayStart).
		Order("created_at DESC").Find(&thisWeek).Error; err != nil {
		return nil, nil, nil, nil, err
	}

	// Older
	if err := r.db.Where("recipient_id = ? AND created_at < ?", recipientID, weekStart).
		Order("created_at DESC").Limit(50).Find(&older).Error; err != nil {
		return nil, nil, nil, nil, err
	}

	return today, yesterday, thisWeek, older, nil
}

func (r *postgresNotificationRepository) GetUnreadCount(recipientID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.Notification{}).Where("recipient_id = ? AND is_read = false", recipientID).Count(&count).Error
	return count, err
}

func (r *postgresNotificationRepository) MarkAsRead(notificationID uint) error {
	return r.db.Model(&models.Notification{}).Where("id = ?", notificationID).Update("is_read", true).Error
}

func (r *postgresNotificationRepository) MarkAllAsRead(recipientID uint) error {
	return r.db.Model(&models.Notification{}).Where("recipient_id = ? AND is_read = false", recipientID).Update("is_read", true).Error
}
