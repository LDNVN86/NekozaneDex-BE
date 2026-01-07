package services

import (
	"nekozanedex/internal/models"
	"nekozanedex/internal/repositories"

	"github.com/google/uuid"
)

type NotificationService interface {
	CreateNotification(userID uuid.UUID, notifType, title string, content, link *string) error
	GetUserNotifications(userID uuid.UUID, page, limit int) ([]models.Notification, int64, error)
	MarkAsRead(notificationID uuid.UUID) error
	MarkAllAsRead(userID uuid.UUID) error
	GetUnreadCount(userID uuid.UUID) int64

	// Notification helpers
	NotifyNewChapter(userID uuid.UUID, storyTitle string, chapterNumber int, storySlug string) error
	NotifyCommentReply(userID uuid.UUID, commenterName string, storySlug string) error
}

type notificationService struct {
	notificationRepo repositories.NotificationRepository
}

func NewNotificationService(notificationRepo repositories.NotificationRepository) NotificationService {
	return &notificationService{
		notificationRepo: notificationRepo,
	}
}

// CreateNotification - Tạo notification
func (s *notificationService) CreateNotification(userID uuid.UUID, notifType, title string, content, link *string) error {
	notification := &models.Notification{
		UserID:  userID,
		Type:    notifType,
		Title:   title,
		Content: content,
		Link:    link,
		IsRead:  false,
	}
	return s.notificationRepo.CreateNotification(notification)
}

// GetUserNotifications - Lấy danh sách notifications của user
func (s *notificationService) GetUserNotifications(userID uuid.UUID, page, limit int) ([]models.Notification, int64, error) {
	return s.notificationRepo.GetNotificationsByUser(userID, page, limit)
}

// MarkAsRead - Đánh dấu đã đọc
func (s *notificationService) MarkAsRead(notificationID uuid.UUID) error {
	return s.notificationRepo.MarkNotificationAsRead(notificationID)
}

// MarkAllAsRead - Đánh dấu tất cả đã đọc
func (s *notificationService) MarkAllAsRead(userID uuid.UUID) error {
	return s.notificationRepo.MarkAllNotificationsAsRead(userID)
}

// GetUnreadCount - Lấy số thông báo chưa đọc
func (s *notificationService) GetUnreadCount(userID uuid.UUID) int64 {
	return s.notificationRepo.GetUnreadNotificationCount(userID)
}

// NotifyNewChapter - Thông báo chapter mới
func (s *notificationService) NotifyNewChapter(userID uuid.UUID, storyTitle string, chapterNumber int, storySlug string) error {
	title := "📖 Chapter mới!"
	content := storyTitle + " vừa cập nhật chapter " + string(rune(chapterNumber))
	link := "/stories/" + storySlug

	return s.CreateNotification(userID, "new_chapter", title, &content, &link)
}

// NotifyCommentReply - Thông báo có reply comment
func (s *notificationService) NotifyCommentReply(userID uuid.UUID, commenterName string, storySlug string) error {
	title := "💬 Có người trả lời bình luận của bạn"
	content := commenterName + " đã trả lời bình luận của bạn"
	link := "/stories/" + storySlug

	return s.CreateNotification(userID, "reply", title, &content, &link)
}
