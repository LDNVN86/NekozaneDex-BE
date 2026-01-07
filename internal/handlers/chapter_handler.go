package handlers

import (
	"encoding/json"
	"strconv"
	"time"

	"nekozanedex/internal/models"
	"nekozanedex/internal/services"
	"nekozanedex/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ChapterHandler struct {
	chapterService services.ChapterService
}

func NewChapterHandler(chapterService services.ChapterService) *ChapterHandler {
	return &ChapterHandler{chapterService: chapterService}
}

// CreateChapterRequest - Request body tạo chapter (Manga)
type CreateChapterRequest struct {
	Title  string   `json:"title" binding:"required"`
	Images []string `json:"images" binding:"required"` // URLs of manga pages
}

// ScheduleChapterRequest - Request body hẹn giờ đăng
type ScheduleChapterRequest struct {
	ScheduledAt string `json:"scheduled_at" binding:"required"` // RFC3339 format
}

// BulkImportRequest - Request body import nhiều chapters
type BulkImportRequest struct {
	Chapters []struct {
		Title  string   `json:"title" binding:"required"`
		Images []string `json:"images" binding:"required"`
	} `json:"chapters" binding:"required,min=1"`
}

// ============ PUBLIC ENDPOINTS ============

// GetChapterByNumber godoc
// @Summary Lấy chapter theo số
// @Tags Chapters
// @Produce json
// @Param slug path string true "Story Slug"
// @Param number path int true "Chapter Number"
// @Success 200 {object} response.Response
// @Router /api/stories/{slug}/chapters/{number} [get]
func (h *ChapterHandler) GetChapterByNumber(c *gin.Context) {
	storySlug := c.Param("slug")
	chapterNumber, err := strconv.Atoi(c.Param("number"))
	if err != nil {
		response.BadRequest(c, "Số chapter không hợp lệ")
		return
	}

	chapter, err := h.chapterService.GetChapterByNumber(storySlug, chapterNumber)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Oke(c, chapter)
}

// GetChaptersByStory godoc
// @Summary Lấy danh sách chapters của truyện
// @Tags Chapters
// @Produce json
// @Param slug path string true "Story Slug"
// @Success 200 {object} response.Response
// @Router /api/stories/{slug}/chapters [get]
func (h *ChapterHandler) GetChaptersByStory(c *gin.Context) {
	storySlug := c.Param("slug")

	chapters, err := h.chapterService.GetChaptersByStory(storySlug)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Oke(c, chapters)
}

// ============ ADMIN ENDPOINTS ============

// CreateChapter godoc
// @Summary Tạo chapter mới (Admin)
// @Tags Admin - Chapters
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param storyId path string true "Story ID"
// @Param body body CreateChapterRequest true "Chapter Info"
// @Success 201 {object} response.Response
// @Router /api/admin/stories/{storyId}/chapters [post]
func (h *ChapterHandler) CreateChapter(c *gin.Context) {
	storyIDStr := c.Param("storyId")
	storyID, err := uuid.Parse(storyIDStr)
	if err != nil {
		response.BadRequest(c, "Story ID không hợp lệ")
		return
	}

	var req CreateChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Dữ liệu không hợp lệ")
		return
	}

	// Convert images to JSON
	imagesJSON, err := json.Marshal(req.Images)
	if err != nil {
		response.BadRequest(c, "Không thể xử lý danh sách ảnh")
		return
	}

	chapter := &models.Chapter{
		Title:     req.Title,
		Images:    imagesJSON,
		PageCount: len(req.Images),
	}

	if err := h.chapterService.CreateChapter(storyID, chapter); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Created(c, chapter)
}

// UpdateChapter godoc
// @Summary Cập nhật chapter (Admin)
// @Tags Admin - Chapters
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Chapter ID"
// @Param body body CreateChapterRequest true "Chapter Info"
// @Success 200 {object} response.Response
// @Router /api/admin/chapters/{id} [put]
func (h *ChapterHandler) UpdateChapter(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "ID không hợp lệ")
		return
	}

	var req CreateChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Dữ liệu không hợp lệ")
		return
	}

	// Convert images to JSON
	imagesJSON, err := json.Marshal(req.Images)
	if err != nil {
		response.BadRequest(c, "Không thể xử lý danh sách ảnh")
		return
	}

	chapter := &models.Chapter{
		Title:     req.Title,
		Images:    imagesJSON,
		PageCount: len(req.Images),
	}

	if err := h.chapterService.UpdateChapter(id, chapter); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Oke(c, gin.H{"message": "Cập nhật thành công"})
}

// DeleteChapter godoc
// @Summary Xóa chapter (Admin)
// @Tags Admin - Chapters
// @Security BearerAuth
// @Produce json
// @Param id path string true "Chapter ID"
// @Success 200 {object} response.Response
// @Router /api/admin/chapters/{id} [delete]
func (h *ChapterHandler) DeleteChapter(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "ID không hợp lệ")
		return
	}

	if err := h.chapterService.DeleteChapter(id); err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Oke(c, gin.H{"message": "Xóa thành công"})
}

// PublishChapter godoc
// @Summary Xuất bản chapter (Admin)
// @Tags Admin - Chapters
// @Security BearerAuth
// @Produce json
// @Param id path string true "Chapter ID"
// @Success 200 {object} response.Response
// @Router /api/admin/chapters/{id}/publish [post]
func (h *ChapterHandler) PublishChapter(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "ID không hợp lệ")
		return
	}

	if err := h.chapterService.PublishChapter(id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Oke(c, gin.H{"message": "Xuất bản thành công"})
}

// ScheduleChapter godoc
// @Summary Hẹn giờ xuất bản chapter (Admin)
// @Tags Admin - Chapters
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Chapter ID"
// @Param body body ScheduleChapterRequest true "Schedule Info"
// @Success 200 {object} response.Response
// @Router /api/admin/chapters/{id}/schedule [post]
func (h *ChapterHandler) ScheduleChapter(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "ID không hợp lệ")
		return
	}

	var req ScheduleChapterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Dữ liệu không hợp lệ")
		return
	}

	scheduledAt, err := time.Parse(time.RFC3339, req.ScheduledAt)
	if err != nil {
		response.BadRequest(c, "Định dạng thời gian không hợp lệ (sử dụng RFC3339)")
		return
	}

	if err := h.chapterService.ScheduleChapter(id, scheduledAt); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Oke(c, gin.H{"message": "Đã hẹn giờ xuất bản"})
}

// BulkImportChapters godoc
// @Summary Import nhiều chapters cùng lúc (Admin)
// @Tags Admin - Chapters
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param storyId path string true "Story ID"
// @Param body body BulkImportRequest true "Chapters Info"
// @Success 201 {object} response.Response
// @Router /api/admin/stories/{storyId}/chapters/bulk [post]
func (h *ChapterHandler) BulkImportChapters(c *gin.Context) {
	storyIDStr := c.Param("storyId")
	storyID, err := uuid.Parse(storyIDStr)
	if err != nil {
		response.BadRequest(c, "Story ID không hợp lệ")
		return
	}

	var req BulkImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Dữ liệu không hợp lệ")
		return
	}

	chapters := make([]models.Chapter, len(req.Chapters))
	for i, ch := range req.Chapters {
		imagesJSON, _ := json.Marshal(ch.Images)
		chapters[i] = models.Chapter{
			Title:     ch.Title,
			Images:    imagesJSON,
			PageCount: len(ch.Images),
		}
	}

	if err := h.chapterService.BulkImportChapters(storyID, chapters); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Created(c, gin.H{
		"message": "Import thành công",
		"count":   len(chapters),
	})
}

// GetChapterByID godoc
// @Summary Lấy chapter theo ID (Admin)
// @Tags Admin - Chapters
// @Security BearerAuth
// @Produce json
// @Param id path string true "Chapter ID"
// @Success 200 {object} response.Response
// @Router /api/admin/chapters/{id} [get]
func (h *ChapterHandler) GetChapterByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "ID không hợp lệ")
		return
	}

	chapter, err := h.chapterService.GetChapterByID(id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Oke(c, chapter)
}
