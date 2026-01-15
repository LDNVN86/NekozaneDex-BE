package handlers

import (
	"log"
	"regexp"
	"strconv"
	"strings"

	"nekozanedex/internal/centrifugo"
	"nekozanedex/internal/models"
	"nekozanedex/internal/repositories"
	"nekozanedex/internal/services"
	"nekozanedex/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CommentHandler struct {
	commentService      services.CommentService
	notificationService services.NotificationService
	userRepo            repositories.UserRepository
	storyRepo           repositories.StoryRepository
	commentLikeRepo     repositories.CommentLikeRepository
	centrifugoClient    *centrifugo.Client
	reportService       services.CommentReportService
}

func NewCommentHandler(
	commentService services.CommentService,
	notificationService services.NotificationService,
	userRepo repositories.UserRepository,
	storyRepo repositories.StoryRepository,
	commentLikeRepo repositories.CommentLikeRepository,
	centrifugoClient *centrifugo.Client,
	reportService services.CommentReportService,
) *CommentHandler {
	return &CommentHandler{
		commentService:      commentService,
		notificationService: notificationService,
		userRepo:            userRepo,
		storyRepo:           storyRepo,
		commentLikeRepo:     commentLikeRepo,
		centrifugoClient:    centrifugoClient,
		reportService:       reportService,
	}
}

type CreateCommentRequest struct {
	Content   string  `json:"content" binding:"required,max=2000"`
	ChapterID *string `json:"chapter_id"`
}

type ReplyCommentRequest struct {
	Content string `json:"content" binding:"required,max=2000"`
}

type UpdateCommentRequest struct {
	Content string `json:"content" binding:"required,max=2000"`
}

type ReportCommentRequest struct {
	Reason string `json:"reason" binding:"required,max=500"`
}

type ResolveReportRequest struct {
	Status string `json:"status" binding:"required,oneof=resolved dismissed"`
}

func parseTagNames(content string) []string {
	re := regexp.MustCompile(`@([a-zA-Z0-9]+)`)
	matches := re.FindAllStringSubmatch(content, -1)

	tagNames := make([]string, 0, len(matches))
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) > 1 {
			tagName := strings.ToLower(match[1])
			if !seen[tagName] {
				tagNames = append(tagNames, tagName)
				seen[tagName] = true
			}
		}
	}
	return tagNames
}

// GetCommentsByStory godoc
// @Summary Lấy comments của truyện
// @Tags Comments
// @Produce json
// @Param storyId path string true "Story ID"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} response.Pagination
// @Router /api/stories/{storyId}/comments [get]
func (h *CommentHandler) GetCommentsByStory(c *gin.Context) {
	storyIDStr := c.Param("storyId")
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

	enrichedComments := h.enrichCommentsWithLikeStatus(c, comments)
	response.PaginatedResponse(c, enrichedComments, page, limit, total)
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

	enrichedComments := h.enrichCommentsWithLikeStatus(c, comments)
	response.PaginatedResponse(c, enrichedComments, page, limit, total)
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

	storySlug := ""
	if story, err := h.storyRepo.FindStoryByID(storyID); err == nil {
		storySlug = story.Slug
	}

	if storySlug != "" {
		go h.processMentions(req.Content, comment.User.Username, storySlug, userID.(uuid.UUID), nil)
	}
	if h.centrifugoClient != nil {
		go func() {
			if err := h.centrifugoClient.Publish("story:"+storyID.String(), centrifugo.CommentEvent{
				Type:    "new_comment",
				Comment: comment,
			}); err != nil {
				log.Printf("[Centrifugo] Publish error: %v", err)
			}
		}()
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

	storySlug := ""
	if story, err := h.storyRepo.FindStoryByID(reply.StoryID); err == nil {
		storySlug = story.Slug
	}

	var notifiedUserIDs []uuid.UUID

	if reply.Parent != nil && reply.Parent.UserID != userID.(uuid.UUID) && storySlug != "" {
		notifiedUserIDs = append(notifiedUserIDs, reply.Parent.UserID)
		go func() {
			if err := h.notificationService.NotifyCommentReply(
				reply.Parent.UserID,
				reply.User.Username,
				storySlug,
			); err != nil {
				log.Printf("[Notification] Reply notification error: %v", err)
			}
		}()
	}

	if storySlug != "" {
		go h.processMentions(req.Content, reply.User.Username, storySlug, userID.(uuid.UUID), notifiedUserIDs)
	}
	if h.centrifugoClient != nil {
		go func() {
			if err := h.centrifugoClient.Publish("story:"+reply.StoryID.String(), centrifugo.CommentEvent{
				Type:    "reply_comment",
				Comment: reply,
			}); err != nil {
				log.Printf("[Centrifugo] Publish error: %v", err)
			}
		}()
	}

	response.Created(c, reply)
}

func (h *CommentHandler) processMentions(content, mentionerName, storySlug string, excludeUserID uuid.UUID, skipUserIDs []uuid.UUID) {
	tagNames := parseTagNames(content)
	if len(tagNames) == 0 {
		return
	}

	users, err := h.userRepo.FindUsersByTagNames(tagNames)
	if err != nil {
		log.Printf("[Mention] Find users error: %v", err)
		return
	}

	skipMap := make(map[uuid.UUID]bool)
	for _, id := range skipUserIDs {
		skipMap[id] = true
	}
	for _, user := range users {
		if user.ID == excludeUserID || skipMap[user.ID] {
			continue
		}
		if err := h.notificationService.NotifyMention(user.ID, mentionerName, storySlug); err != nil {
			log.Printf("[Mention] Notification error for user %s: %v", user.Username, err)
		}
	}
}

type CommentWithLikeStatus struct {
	models.Comment
	UserHasLiked bool `json:"user_has_liked"`
}

func (h *CommentHandler) enrichCommentsWithLikeStatus(c *gin.Context, comments []models.Comment) []CommentWithLikeStatus {
	result := make([]CommentWithLikeStatus, len(comments))
	
	var currentUserID uuid.UUID
	if userID, exists := c.Get("user_id"); exists {
		currentUserID = userID.(uuid.UUID)
	}

	var allCommentIDs []uuid.UUID
	for _, comment := range comments {
		allCommentIDs = append(allCommentIDs, comment.ID)
		for _, reply := range comment.Replies {
			allCommentIDs = append(allCommentIDs, reply.ID)
		}
	}

	var likedIDs []uuid.UUID
	if currentUserID != uuid.Nil {
		likedIDs = h.commentLikeRepo.GetUserLikedCommentIDs(currentUserID, allCommentIDs)
	}

	likedMap := make(map[uuid.UUID]bool)
	for _, id := range likedIDs {
		likedMap[id] = true
	}
	for i, comment := range comments {
		enrichedReplies := make([]CommentWithLikeStatus, len(comment.Replies))
		for j, reply := range comment.Replies {
			enrichedReplies[j] = CommentWithLikeStatus{
				Comment:      reply,
				UserHasLiked: likedMap[reply.ID],
			}
		}
		
		result[i] = CommentWithLikeStatus{
			Comment:      comment,
			UserHasLiked: likedMap[comment.ID],
		}
	}

	return result
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

// UpdateComment godoc
// @Summary Cập nhật comment
// @Tags Comments
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param commentId path string true "Comment ID"
// @Param body body UpdateCommentRequest true "Updated content"
// @Success 200 {object} response.Response
// @Router /api/comments/{commentId} [put]
func (h *CommentHandler) UpdateComment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "Chưa đăng nhập")
		return
	}

	commentIDStr := c.Param("commentId")
	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		response.BadRequest(c, "Comment ID không hợp lệ")
		return
	}

	var req UpdateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Dữ liệu không hợp lệ")
		return
	}

	// Get old comment to identify previous mentions
	oldComment, err := h.commentService.FindCommentByID(commentID)
	if err != nil {
		response.NotFound(c, "Bình luận không tồn tại")
		return
	}

	comment, err := h.commentService.UpdateComment(userID.(uuid.UUID), commentID, req.Content)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Get story slug for notification links
	storySlug := ""
	if story, err := h.storyRepo.FindStoryByID(comment.StoryID); err == nil {
		storySlug = story.Slug
	}

	if storySlug != "" {
		// Identify new mentions: mentioned in content but not in oldContent
		oldMentions := parseTagNames(oldComment.Content)
		newMentions := parseTagNames(comment.Content)

		oldMentionMap := make(map[string]bool)
		for _, tag := range oldMentions {
			oldMentionMap[tag] = true
		}

		// Find truly new tagnames
		var trulyNewTags []string
		for _, tag := range newMentions {
			if !oldMentionMap[tag] {
				trulyNewTags = append(trulyNewTags, tag)
			}
		}

		if len(trulyNewTags) > 0 {
			// Find mentioned users by tag_name
			users, err := h.userRepo.FindUsersByTagNames(trulyNewTags)
			if err == nil {
				for _, user := range users {
					if user.ID != userID.(uuid.UUID) {
						if err := h.notificationService.NotifyMention(user.ID, comment.User.Username, storySlug); err != nil {
							log.Printf("[Mention] Update notification error for user %s: %v", user.Username, err)
						}
					}
				}
			}
		}
	}

	response.Oke(c, comment)
}

// ToggleLike godoc
// @Summary Toggle like on a comment
// @Tags Comments
// @Security BearerAuth
// @Produce json
// @Param commentId path string true "Comment ID"
// @Success 200 {object} response.Response
// @Router /api/comments/{commentId}/like [post]
func (h *CommentHandler) ToggleLike(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "Chưa đăng nhập")
		return
	}

	commentIDStr := c.Param("commentId")
	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		response.BadRequest(c, "Comment ID không hợp lệ")
		return
	}

	// Check if already liked
	hasLiked := h.commentLikeRepo.HasUserLiked(commentID, userID.(uuid.UUID))

	if hasLiked {
		// Unlike
		if err := h.commentLikeRepo.DeleteLike(commentID, userID.(uuid.UUID)); err != nil {
			response.InternalServerError(c, "Không thể bỏ thích")
			return
		}
	} else {
		like := &models.CommentLike{
			CommentID: commentID,
			UserID:    userID.(uuid.UUID),
		}
		if err := h.commentLikeRepo.CreateLike(like); err != nil {
			response.InternalServerError(c, "Không thể thích")
			return
		}
	}

	likeCount := h.commentLikeRepo.GetLikeCount(commentID)
	h.commentService.UpdateLikeCount(commentID, int(likeCount))

	response.Oke(c, gin.H{
		"liked":      !hasLiked,
		"like_count": likeCount,
	})
}

// TogglePin godoc
// @Summary Ghim/Bỏ ghim bình luận (Admin only)
// @Tags Comments
// @Security BearerAuth
// @Produce json
// @Param commentId path string true "Comment ID"
// @Success 200 {object} response.Response
// @Router /api/comments/{commentId}/pin [post]
func (h *CommentHandler) TogglePin(c *gin.Context) {
	// Check admin permission
	role, exists := c.Get("role")
	if !exists || role.(string) != "admin" {
		response.Forbidden(c, "Bạn không có quyền thực hiện hành động này")
		return
	}

	commentIDStr := c.Param("commentId")
	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		response.BadRequest(c, "Comment ID không hợp lệ")
		return
	}

	comment, err := h.commentService.FindCommentByID(commentID)
	if err != nil {
		response.NotFound(c, "Bình luận không tồn tại")
		return
	}

	newPinnedStatus := !comment.IsPinned
	if err := h.commentService.TogglePin(commentID, newPinnedStatus); err != nil {
		response.InternalServerError(c, "Không thể ghim bình luận")
		return
	}

	response.Oke(c, gin.H{
		"is_pinned": newPinnedStatus,
		"message":   "Cập nhật trạng thái ghim thành công",
	})
}

// ReportComment godoc
// @Summary Báo cáo bình luận
// @Tags Comments
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param commentId path string true "Comment ID"
// @Param body body ReportCommentRequest true "Report Reason"
// @Success 201 {object} response.Response
// @Router /api/comments/{commentId}/report [post]
func (h *CommentHandler) ReportComment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "Chưa đăng nhập")
		return
	}

	commentIDStr := c.Param("commentId")
	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		response.BadRequest(c, "Comment ID không hợp lệ")
		return
	}

	var req ReportCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Dữ liệu không hợp lệ")
		return
	}

	report, err := h.reportService.ReportComment(userID.(uuid.UUID), commentID, req.Reason)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Created(c, report)
}

// GetReports godoc
// @Summary Lấy danh sách báo cáo (Admin only)
// @Tags Admin Comments
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param status query string false "Report status (pending, resolved, dismissed)"
// @Success 200 {object} response.Pagination
// @Router /api/admin/comments/reports [get]
func (h *CommentHandler) GetReports(c *gin.Context) {
	// Check admin role
	role, _ := c.Get("role")
	if role != "admin" {
		response.Forbidden(c, "Bạn không có quyền thực hiện hành động này")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")

	reports, total, err := h.reportService.GetReports(page, limit, status)
	if err != nil {
		response.InternalServerError(c, "Không thể lấy danh sách báo cáo")
		return
	}

	response.PaginatedResponse(c, reports, page, limit, total)
}

// ResolveReport godoc
// @Summary Xử lý báo cáo (Admin only)
// @Tags Admin Comments
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param reportId path string true "Report ID"
// @Param body body ResolveReportRequest true "Resolve Action"
// @Success 200 {object} response.Response
// @Router /api/admin/comments/reports/{reportId} [put]
func (h *CommentHandler) ResolveReport(c *gin.Context) {
	// Check admin role
	role, _ := c.Get("role")
	if role != "admin" {
		response.Forbidden(c, "Bạn không có quyền thực hiện hành động này")
		return
	}

	reportIDStr := c.Param("reportId")
	reportID, err := uuid.Parse(reportIDStr)
	if err != nil {
		response.BadRequest(c, "Report ID không hợp lệ")
		return
	}

	var req ResolveReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Dữ liệu không hợp lệ")
		return
	}

	if err := h.reportService.UpdateReportStatus(reportID, req.Status); err != nil {
		response.InternalServerError(c, "Không thể xử lý báo cáo")
		return
	}

	response.Oke(c, gin.H{"message": "Đã xử lý báo cáo thành công"})
}

