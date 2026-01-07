package handlers

import (
	"strconv"

	"nekozanedex/internal/services"
	"nekozanedex/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type BookmarkHandler struct {
	bookmarkService services.BookmarkService
}

func NewBookmarkHandler(bookmarkService services.BookmarkService) *BookmarkHandler {
	return &BookmarkHandler{bookmarkService: bookmarkService}
}

// AddBookmark godoc
// @Summary Thêm bookmark
// @Tags Bookmarks
// @Security BearerAuth
// @Produce json
// @Param storyId path string true "Story ID"
// @Success 201 {object} response.Response
// @Router /api/bookmarks/{storyId} [post]
func (h *BookmarkHandler) AddBookmark(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "Chưa đăng nhập")
		return
	}

	storyIDStr := c.Param("storyId")
	storyID, err := uuid.Parse(storyIDStr)
	if err != nil {
		response.BadRequest(c, "Story ID không hợp lệ")
		return
	}

	if err := h.bookmarkService.AddBookmark(userID.(uuid.UUID), storyID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Created(c, gin.H{"message": "Đã thêm vào bookmark"})
}

// RemoveBookmark godoc
// @Summary Xóa bookmark
// @Tags Bookmarks
// @Security BearerAuth
// @Produce json
// @Param storyId path string true "Story ID"
// @Success 200 {object} response.Response
// @Router /api/bookmarks/{storyId} [delete]
func (h *BookmarkHandler) RemoveBookmark(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "Chưa đăng nhập")
		return
	}

	storyIDStr := c.Param("storyId")
	storyID, err := uuid.Parse(storyIDStr)
	if err != nil {
		response.BadRequest(c, "Story ID không hợp lệ")
		return
	}

	if err := h.bookmarkService.RemoveBookmark(userID.(uuid.UUID), storyID); err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Oke(c, gin.H{"message": "Đã xóa khỏi bookmark"})
}

// GetMyBookmarks godoc
// @Summary Lấy danh sách bookmark của tôi
// @Tags Bookmarks
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} response.Pagination
// @Router /api/bookmarks [get]
func (h *BookmarkHandler) GetMyBookmarks(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "Chưa đăng nhập")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	bookmarks, total, err := h.bookmarkService.GetUserBookmarks(userID.(uuid.UUID), page, limit)
	if err != nil {
		response.InternalServerError(c, "Không thể lấy bookmark")
		return
	}

	response.PaginatedResponse(c, bookmarks, page, limit, total)
}

// CheckBookmark godoc
// @Summary Kiểm tra đã bookmark chưa
// @Tags Bookmarks
// @Security BearerAuth
// @Produce json
// @Param storyId path string true "Story ID"
// @Success 200 {object} response.Response
// @Router /api/bookmarks/{storyId}/check [get]
func (h *BookmarkHandler) CheckBookmark(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "Chưa đăng nhập")
		return
	}

	storyIDStr := c.Param("storyId")
	storyID, err := uuid.Parse(storyIDStr)
	if err != nil {
		response.BadRequest(c, "Story ID không hợp lệ")
		return
	}

	isBookmarked := h.bookmarkService.IsBookmarked(userID.(uuid.UUID), storyID)

	response.Oke(c, gin.H{"is_bookmarked": isBookmarked})
}
