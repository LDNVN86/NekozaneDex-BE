package handlers

import (
	"strconv"

	"nekozanedex/internal/models"
	"nekozanedex/internal/services"
	"nekozanedex/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type StoryHandler struct {
	storyService services.StoryService
}

func NewStoryHandler(storyService services.StoryService) *StoryHandler {
	return &StoryHandler{storyService: storyService}
}

// CreateStoryRequest - Request body tạo truyện
type CreateStoryRequest struct {
	Title         string   `json:"title" binding:"required"`
	Description   *string  `json:"description"`
	CoverImageURL *string  `json:"cover_image_url"`
	AuthorName    *string  `json:"author_name"`
	Status        string   `json:"status"`
	IsPublished   bool     `json:"is_published"`
	GenreIDs      []string `json:"genre_ids"`
}

// ============ PUBLIC ENDPOINTS ============

// GetStories godoc
// @Summary Lấy danh sách truyện
// @Tags Stories
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} response.Pagination
// @Router /api/stories [get]
func (h *StoryHandler) GetStories(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	stories, total, err := h.storyService.GetAllStories(page, limit)
	if err != nil {
		response.InternalServerError(c, "Không thể lấy danh sách truyện")
		return
	}

	response.PaginatedResponse(c, stories, page, limit, total)
}

// GetStoryBySlug godoc
// @Summary Lấy chi tiết truyện theo slug
// @Tags Stories
// @Produce json
// @Param slug path string true "Story Slug"
// @Success 200 {object} response.Response
// @Router /api/stories/{slug} [get]
func (h *StoryHandler) GetStoryBySlug(c *gin.Context) {
	slug := c.Param("slug")

	story, err := h.storyService.GetStoryBySlug(slug)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Oke(c, story)
}

// GetLatestStories godoc
// @Summary Lấy truyện mới cập nhật
// @Tags Stories
// @Produce json
// @Param limit query int false "Number of stories" default(10)
// @Success 200 {object} response.Response
// @Router /api/stories/latest [get]
func (h *StoryHandler) GetLatestStories(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 || limit > 50 {
		limit = 10
	}

	stories, err := h.storyService.GetLatestStories(limit)
	if err != nil {
		response.InternalServerError(c, "Không thể lấy truyện mới")
		return
	}

	response.Oke(c, stories)
}

// GetHotStories godoc
// @Summary Lấy truyện hot
// @Tags Stories
// @Produce json
// @Param limit query int false "Number of stories" default(10)
// @Success 200 {object} response.Response
// @Router /api/stories/hot [get]
func (h *StoryHandler) GetHotStories(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 || limit > 50 {
		limit = 10
	}

	stories, err := h.storyService.GetHotStories(limit)
	if err != nil {
		response.InternalServerError(c, "Không thể lấy truyện hot")
		return
	}

	response.Oke(c, stories)
}

// GetRandomStory godoc
// @Summary Lấy truyện ngẫu nhiên
// @Tags Stories
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/stories/random [get]
func (h *StoryHandler) GetRandomStory(c *gin.Context) {
	story, err := h.storyService.GetRandomStory()
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Oke(c, story)
}

// SearchStories godoc
// @Summary Tìm kiếm truyện
// @Tags Stories
// @Produce json
// @Param q query string true "Search query"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} response.Pagination
// @Router /api/stories/search [get]
func (h *StoryHandler) SearchStories(c *gin.Context) {
	query := c.Query("q")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if query == "" {
		response.BadRequest(c, "Từ khóa tìm kiếm không được để trống")
		return
	}

	stories, total, err := h.storyService.SearchStories(query, page, limit)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.PaginatedResponse(c, stories, page, limit, total)
}

// GetStoriesByGenre godoc
// @Summary Lấy truyện theo thể loại
// @Tags Stories
// @Produce json
// @Param genre path string true "Genre Slug"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} response.Pagination
// @Router /api/genres/{genre}/stories [get]
func (h *StoryHandler) GetStoriesByGenre(c *gin.Context) {
	genreSlug := c.Param("genre")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	stories, total, err := h.storyService.GetStoriesByGenre(genreSlug, page, limit)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.PaginatedResponse(c, stories, page, limit, total)
}

// ============ ADMIN ENDPOINTS ============

// CreateStory godoc
// @Summary Tạo truyện mới (Admin)
// @Tags Admin - Stories
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body CreateStoryRequest true "Story Info"
// @Success 201 {object} response.Response
// @Router /api/admin/stories [post]
func (h *StoryHandler) CreateStory(c *gin.Context) {
	var req CreateStoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Dữ liệu không hợp lệ")
		return
	}

	story := &models.Story{
		Title:         req.Title,
		Description:   req.Description,
		CoverImageURL: req.CoverImageURL,
		AuthorName:    req.AuthorName,
		Status:        req.Status,
		IsPublished:   req.IsPublished,
	}

	if err := h.storyService.CreateStory(story); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Created(c, story)
}

// UpdateStory godoc
// @Summary Cập nhật truyện (Admin)
// @Tags Admin - Stories
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Story ID"
// @Param body body CreateStoryRequest true "Story Info"
// @Success 200 {object} response.Response
// @Router /api/admin/stories/{id} [put]
func (h *StoryHandler) UpdateStory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "ID không hợp lệ")
		return
	}

	var req CreateStoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Dữ liệu không hợp lệ")
		return
	}

	story := &models.Story{
		Title:         req.Title,
		Description:   req.Description,
		CoverImageURL: req.CoverImageURL,
		AuthorName:    req.AuthorName,
		Status:        req.Status,
		IsPublished:   req.IsPublished,
	}

	if err := h.storyService.UpdateStory(id, story); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Oke(c, gin.H{"message": "Cập nhật thành công"})
}

// DeleteStory godoc
// @Summary Xóa truyện (Admin)
// @Tags Admin - Stories
// @Security BearerAuth
// @Produce json
// @Param id path string true "Story ID"
// @Success 200 {object} response.Response
// @Router /api/admin/stories/{id} [delete]
func (h *StoryHandler) DeleteStory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "ID không hợp lệ")
		return
	}

	if err := h.storyService.DeleteStory(id); err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Oke(c, gin.H{"message": "Xóa thành công"})
}

// GetStoryByID godoc
// @Summary Lấy truyện theo ID (Admin)
// @Tags Admin - Stories
// @Security BearerAuth
// @Produce json
// @Param id path string true "Story ID"
// @Success 200 {object} response.Response
// @Router /api/admin/stories/{id} [get]
func (h *StoryHandler) GetStoryByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "ID không hợp lệ")
		return
	}

	story, err := h.storyService.GetStoryByID(id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Oke(c, story)
}

// GetAllStoriesAdmin godoc
// @Summary Lấy tất cả truyện (Admin)
// @Tags Admin - Stories
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} response.Pagination
// @Router /api/admin/stories [get]
func (h *StoryHandler) GetAllStoriesAdmin(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	stories, total, err := h.storyService.GetAllStoriesAdmin(page, limit)
	if err != nil {
		response.InternalServerError(c, "Không thể lấy danh sách truyện")
		return
	}

	response.PaginatedResponse(c, stories, page, limit, total)
}
