package services

import (
	"log"

	"nekozanedex/internal/centrifugo"
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
	NotifyCommentReply(userID uuid.UUID, commenterName, storySlug string) error
	NotifyMention(userID uuid.UUID, mentionerName, storySlug string) error
}

type notificationService struct {
	notificationRepo repositories.NotificationRepository
	centrifugoClient *centrifugo.Client
}

func NewNotificationService(
	notificationRepo repositories.NotificationRepository,
	centrifugoClient *centrifugo.Client,
) NotificationService {
	return &notificationService{
		notificationRepo: notificationRepo,
		centrifugoClient: centrifugoClient,
	}
}

// CreateNotification - Tạo notification và push realtime
func (s *notificationService) CreateNotification(userID uuid.UUID, notifType, title string, content, link *string) error {
	notification := &models.Notification{
		UserID:  userID,
		Type:    notifType,
		Title:   title,
		Message: content,
		Link:    link,
		IsRead:  false,
	}
	if err := s.notificationRepo.CreateNotification(notification); err != nil {
		return err
	}

	// Push realtime notification via Centrifugo to user's personal channel
	if s.centrifugoClient != nil {
		go func() {
			channel := "user:" + userID.String()
			if err := s.centrifugoClient.Publish(channel, map[string]interface{}{
				"type":         "new_notification",
				"notification": notification,
			}); err != nil {
				log.Printf("[Centrifugo] Failed to push notification to %s: %v", channel, err)
			} else {
				log.Printf("[Centrifugo] Pushed notification to %s", channel)
			}
		}()
	} else {
		log.Printf("[Centrifugo] Client is nil, cannot push notification")
	}

	return nil
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
	link := "/client/stories/" + storySlug

	return s.CreateNotification(userID, "new_chapter", title, &content, &link)
}

// NotifyCommentReply - Thông báo có reply comment
func (s *notificationService) NotifyCommentReply(userID uuid.UUID, commenterName, storySlug string) error {
	title := "💬 Có người trả lời bình luận của bạn"
	content := commenterName + " đã trả lời bình luận của bạn"
	link := "/client/stories/" + storySlug

	return s.CreateNotification(userID, "reply", title, &content, &link)
}

// NotifyMention - Thông báo có người mention
func (s *notificationService) NotifyMention(userID uuid.UUID, mentionerName, storySlug string) error {
	title := "📢 Có người nhắc đến bạn"
	content := mentionerName + " đã nhắc đến bạn trong bình luận"
	link := "/client/stories/" + storySlug

	return s.CreateNotification(userID, "mention", title, &content, &link)
}
