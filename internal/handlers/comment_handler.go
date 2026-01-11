package handlers

import (
	"strconv"

	"nekozanedex/internal/services"
	"nekozanedex/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CommentHandler struct {
	commentService services.CommentService
}

func NewCommentHandler(commentService services.CommentService) *CommentHandler {
	return &CommentHandler{commentService: commentService}
}

// CreateCommentRequest - Request body tạo comment
type CreateCommentRequest struct {
	Content   string  `json:"content" binding:"required,max=2000"`
	ChapterID *string `json:"chapter_id"` // Tùy Nhé, nếu comment cho chapter cụ thể
}

type ReplyCommentRequest struct {
	Content string `json:"content" binding:"required,max=2000"`
}

// GetCommentsByStory godoc
// @Summary Lấy comments của truyện
// @Tags Comments
// @Produce json
// @Param slug path string true "Story Slug"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} response.Pagination
// @Router /api/stories/{slug}/comments [get]
func (h *CommentHandler) GetCommentsByStory(c *gin.Context) {
	storyIDStr := c.Query("story_id")
	storyID, err := uuid.Parse(storyIDStr)
	if err != nil {
		response.BadRequest(c, "Story ID không hợp lệ")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	comments, total, err := h.commentService.GetCommentsByStory(storyID, page, limit)
	if err != nil {
		response.InternalServerError(c, "Không thể lấy comments")
		return
	}

	response.PaginatedResponse(c, comments, page, limit, total)
}

// GetCommentsByChapter godoc
// @Summary Lấy comments của chapter
// @Tags Comments
// @Produce json
// @Param chapterId path string true "Chapter ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} response.Pagination
// @Router /api/chapters/{chapterId}/comments [get]
func (h *CommentHandler) GetCommentsByChapter(c *gin.Context) {
	chapterIDStr := c.Param("chapterId")
	chapterID, err := uuid.Parse(chapterIDStr)
	if err != nil {
		response.BadRequest(c, "Chapter ID không hợp lệ")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	comments, total, err := h.commentService.GetCommentsByChapter(chapterID, page, limit)
	if err != nil {
		response.InternalServerError(c, "Không thể lấy comments")
		return
	}

	response.PaginatedResponse(c, comments, page, limit, total)
}

// CreateComment godoc
// @Summary Tạo comment
// @Tags Comments
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param storyId path string true "Story ID"
// @Param body body CreateCommentRequest true "Comment Info"
// @Success 201 {object} response.Response
// @Router /api/stories/{storyId}/comments [post]
func (h *CommentHandler) CreateComment(c *gin.Context) {
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

	var req CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Dữ liệu không hợp lệ")
		return
	}

	var chapterID *uuid.UUID
	if req.ChapterID != nil {
		parsed, err := uuid.Parse(*req.ChapterID)
		if err == nil {
			chapterID = &parsed
		}
	}

	comment, err := h.commentService.CreateComment(userID.(uuid.UUID), storyID, chapterID, req.Content)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Created(c, comment)
}

// ReplyComment godoc
// @Summary Trả lời comment
// @Tags Comments
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param commentId path string true "Parent Comment ID"
// @Param body body ReplyCommentRequest true "Reply Info"
// @Success 201 {object} response.Response
// @Router /api/comments/{commentId}/reply [post]
func (h *CommentHandler) ReplyComment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "Chưa đăng nhập")
		return
	}

	parentIDStr := c.Param("commentId")
	parentID, err := uuid.Parse(parentIDStr)
	if err != nil {
		response.BadRequest(c, "Comment ID không hợp lệ")
		return
	}

	var req ReplyCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Dữ liệu không hợp lệ")
		return
	}

	reply, err := h.commentService.ReplyComment(userID.(uuid.UUID), parentID, req.Content)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Created(c, reply)
}

// DeleteComment godoc
// @Summary Xóa comment
// @Tags Comments
// @Security BearerAuth
// @Produce json
// @Param commentId path string true "Comment ID"
// @Success 200 {object} response.Response
// @Router /api/comments/{commentId} [delete]
func (h *CommentHandler) DeleteComment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "Chưa đăng nhập")
		return
	}

	role, _ := c.Get("role")
	isAdmin := role == "admin"

	commentIDStr := c.Param("commentId")
	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		response.BadRequest(c, "Comment ID không hợp lệ")
		return
	}

	if err := h.commentService.DeleteComment(userID.(uuid.UUID), commentID, isAdmin); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Oke(c, gin.H{"message": "Xóa comment thành công"})
}
