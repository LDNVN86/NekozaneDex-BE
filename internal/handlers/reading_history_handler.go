package handlers

import (
	"nekozanedex/internal/models"
	"nekozanedex/internal/repositories"
	"nekozanedex/pkg/response"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ReadingHistoryHandler struct {
	historyRepo repositories.ReadingHistoryRepository
}

func NewReadingHistoryHandler(historyRepo repositories.ReadingHistoryRepository) *ReadingHistoryHandler {
	return &ReadingHistoryHandler{historyRepo: historyRepo}
}

type SaveProgressRequest struct {
	StoryID        string `json:"story_id" binding:"required,uuid"`
	ChapterID      string `json:"chapter_id" binding:"required,uuid"`
	ScrollPosition int    `json:"scroll_position"`
}

// SaveProgress godoc
// @Summary Save or update reading progress
// @Tags Reading History
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body SaveProgressRequest true "Reading progress"
// @Success 200 {object} response.Response
// @Router /api/reading-history [post]
func (h *ReadingHistoryHandler) SaveProgress(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "Chưa đăng nhập")
		return
	}

	var req SaveProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Dữ liệu không hợp lệ")
		return
	}

	storyUUID, _ := uuid.Parse(req.StoryID)
	chapterUUID, _ := uuid.Parse(req.ChapterID)

	history := &models.ReadingHistory{
		UserID:         userID.(uuid.UUID),
		StoryID:        storyUUID,
		ChapterID:      chapterUUID,
		LastReadAt:     time.Now(),
		ScrollPosition: req.ScrollPosition,
	}

	if err := h.historyRepo.Upsert(history); err != nil {
		response.InternalServerError(c, "Không thể lưu tiến độ đọc")
		return
	}

	response.Oke(c, gin.H{
		"message": "Đã lưu tiến độ đọc",
	})
}

// GetHistory godoc
// @Summary Get user's reading history with pagination
// @Tags Reading History
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} response.Response
// @Router /api/reading-history [get]
func (h *ReadingHistoryHandler) GetHistory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "Chưa đăng nhập")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 20
	}

	histories, total, err := h.historyRepo.GetByUser(userID.(uuid.UUID), page, limit)
	if err != nil {
		response.InternalServerError(c, "Không thể lấy lịch sử đọc")
		return
	}

	response.PaginatedResponse(c, histories, page, limit, total)
}

// GetContinueReading godoc
// @Summary Get "Continue Reading" stories for home page
// @Tags Reading History
// @Security BearerAuth
// @Produce json
// @Param limit query int false "Number of items" default(6)
// @Success 200 {object} response.Response
// @Router /api/reading-history/continue [get]
func (h *ReadingHistoryHandler) GetContinueReading(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "Chưa đăng nhập")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "6"))
	if limit < 1 || limit > 12 {
		limit = 6
	}

	histories, err := h.historyRepo.GetContinueReading(userID.(uuid.UUID), limit)
	if err != nil {
		response.InternalServerError(c, "Không thể lấy danh sách tiếp tục đọc")
		return
	}

	response.Oke(c, histories)
}

// GetProgressByStory godoc
// @Summary Get reading progress for a specific story
// @Tags Reading History
// @Security BearerAuth
// @Produce json
// @Param storyId path string true "Story ID"
// @Success 200 {object} response.Response
// @Router /api/reading-history/story/{storyId} [get]
func (h *ReadingHistoryHandler) GetProgressByStory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "Chưa đăng nhập")
		return
	}

	storyID, err := uuid.Parse(c.Param("storyId"))
	if err != nil {
		response.BadRequest(c, "Story ID không hợp lệ")
		return
	}

	history, err := h.historyRepo.GetByUserAndStory(userID.(uuid.UUID), storyID)
	if err != nil {
		// Not found is okay - just return null
		response.Oke(c, nil)
		return
	}

	response.Oke(c, history)
}

// DeleteByStory godoc
// @Summary Remove a story from reading history
// @Tags Reading History
// @Security BearerAuth
// @Produce json
// @Param storyId path string true "Story ID"
// @Success 200 {object} response.Response
// @Router /api/reading-history/{storyId} [delete]
func (h *ReadingHistoryHandler) DeleteByStory(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "Chưa đăng nhập")
		return
	}

	storyID, err := uuid.Parse(c.Param("storyId"))
	if err != nil {
		response.BadRequest(c, "Story ID không hợp lệ")
		return
	}

	if err := h.historyRepo.DeleteByStory(userID.(uuid.UUID), storyID); err != nil {
		response.InternalServerError(c, "Không thể xóa lịch sử")
		return
	}

	response.Oke(c, gin.H{
		"message": "Đã xóa khỏi lịch sử đọc",
	})
}

// ClearAll godoc
// @Summary Clear all reading history
// @Tags Reading History
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/reading-history [delete]
func (h *ReadingHistoryHandler) ClearAll(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "Chưa đăng nhập")
		return
	}

	if err := h.historyRepo.DeleteAll(userID.(uuid.UUID)); err != nil {
		response.InternalServerError(c, "Không thể xóa lịch sử")
		return
	}

	response.Oke(c, gin.H{
		"message": "Đã xóa toàn bộ lịch sử đọc",
	})
}
