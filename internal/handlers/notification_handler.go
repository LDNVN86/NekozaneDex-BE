package handlers

import (
	"strconv"

	"nekozanedex/internal/services"
	"nekozanedex/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type NotificationHandler struct {
	notificationService services.NotificationService
}

func NewNotificationHandler(notificationService services.NotificationService) *NotificationHandler {
	return &NotificationHandler{notificationService: notificationService}
}

// GetMyNotifications godoc
// @Summary Lấy danh sách thông báo của tôi
// @Tags Notifications
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} response.Pagination
// @Router /api/notifications [get]
func (h *NotificationHandler) GetMyNotifications(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "Chưa đăng nhập")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	notifications, total, err := h.notificationService.GetUserNotifications(userID.(uuid.UUID), page, limit)
	if err != nil {
		response.InternalServerError(c, "Không thể lấy thông báo")
		return
	}

	response.PaginatedResponse(c, notifications, page, limit, total)
}

// GetUnreadCount godoc
// @Summary Lấy số thông báo chưa đọc
// @Tags Notifications
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/notifications/unread-count [get]
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "Chưa đăng nhập")
		return
	}

	count := h.notificationService.GetUnreadCount(userID.(uuid.UUID))

	response.Oke(c, gin.H{"unread_count": count})
}

// MarkAsRead godoc
// @Summary Đánh dấu đã đọc
// @Tags Notifications
// @Security BearerAuth
// @Produce json
// @Param id path string true "Notification ID"
// @Success 200 {object} response.Response
// @Router /api/notifications/{id}/read [post]
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "ID không hợp lệ")
		return
	}

	if err := h.notificationService.MarkAsRead(id); err != nil {
		response.InternalServerError(c, "Không thể đánh dấu đã đọc")
		return
	}

	response.Oke(c, gin.H{"message": "Đã đánh dấu đã đọc"})
}

// MarkAllAsRead godoc
// @Summary Đánh dấu tất cả đã đọc
// @Tags Notifications
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/notifications/read-all [post]
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "Chưa đăng nhập")
		return
	}

	if err := h.notificationService.MarkAllAsRead(userID.(uuid.UUID)); err != nil {
		response.InternalServerError(c, "Không thể đánh dấu đã đọc")
		return
	}

	response.Oke(c, gin.H{"message": "Đã đánh dấu tất cả đã đọc"})
}
