package repositories

import (
	"nekozanedex/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NotificationRepository interface {
	CreateNotification(notification *models.Notification) error
	FindNotificationByID(id uuid.UUID) (*models.Notification, error)
	GetNotificationsByUser(userID uuid.UUID, page, limit int) ([]models.Notification, int64, error)
	MarkNotificationAsRead(id uuid.UUID) error
	MarkAllNotificationsAsRead(userID uuid.UUID) error
	GetUnreadNotificationCount(userID uuid.UUID) int64
	DeleteNotification(id uuid.UUID) error
}

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

// CreateNotification - Tạo Notification
func (r *notificationRepository) CreateNotification(notification *models.Notification) error {
	return r.db.Create(notification).Error
}

// FindNotificationByID - Tìm Notification theo ID
func (r *notificationRepository) FindNotificationByID(id uuid.UUID) (*models.Notification, error) {
	var notification models.Notification
	err := r.db.First(&notification, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

// GetNotificationsByUser - Lấy Notifications theo User
func (r *notificationRepository) GetNotificationsByUser(userID uuid.UUID, page, limit int) ([]models.Notification, int64, error) {
	var notifications []models.Notification
	var total int64

	r.db.Model(&models.Notification{}).Where("user_id = ?", userID).Count(&total)

	offset := (page - 1) * limit
	err := r.db.Where("user_id = ?", userID).
		Offset(offset).Limit(limit).
		Order("created_at DESC").
		Find(&notifications).Error

	return notifications, total, err
}

// MarkNotificationAsRead - Đánh dấu đã đọc
func (r *notificationRepository) MarkNotificationAsRead(id uuid.UUID) error {
	return r.db.Model(&models.Notification{}).Where("id = ?", id).
		Update("is_read", true).Error
}

// MarkAllNotificationsAsRead - Đánh dấu tất cả đã đọc
func (r *notificationRepository) MarkAllNotificationsAsRead(userID uuid.UUID) error {
	return r.db.Model(&models.Notification{}).Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true).Error
}

// GetUnreadNotificationCount - Đếm số thông báo chưa đọc
func (r *notificationRepository) GetUnreadNotificationCount(userID uuid.UUID) int64 {
	var count int64
	r.db.Model(&models.Notification{}).Where("user_id = ? AND is_read = ?", userID, false).Count(&count)
	return count
}

// DeleteNotification - Xóa Notification
func (r *notificationRepository) DeleteNotification(id uuid.UUID) error {
	return r.db.Delete(&models.Notification{}, "id = ?", id).Error
}
