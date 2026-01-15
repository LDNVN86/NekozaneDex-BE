package services

import (
	"errors"
	"strings"

	"nekozanedex/internal/models"
	"nekozanedex/internal/repositories"

	"github.com/google/uuid"
)

type CommentService interface {
	CreateComment(userID, storyID uuid.UUID, chapterID *uuid.UUID, content string) (*models.Comment, error)
	ReplyComment(userID, parentID uuid.UUID, content string) (*models.Comment, error)
	UpdateComment(userID, commentID uuid.UUID, content string) (*models.Comment, error)
	DeleteComment(userID, commentID uuid.UUID, isAdmin bool) error
	GetCommentsByStory(storyID uuid.UUID, page, limit int) ([]models.Comment, int64, error)
	GetCommentsByChapter(chapterID uuid.UUID, page, limit int) ([]models.Comment, int64, error)
	UpdateLikeCount(commentID uuid.UUID, count int) error
	TogglePin(commentID uuid.UUID, isPinned bool) error
	FindCommentByID(id uuid.UUID) (*models.Comment, error)
}

type commentService struct {
	commentRepo repositories.CommentRepository
	storyRepo   repositories.StoryRepository
	chapterRepo repositories.ChapterRepository
}

func NewCommentService(
	commentRepo repositories.CommentRepository,
	storyRepo repositories.StoryRepository,
	chapterRepo repositories.ChapterRepository,
) CommentService {
	return &commentService{
		commentRepo: commentRepo,
		storyRepo:   storyRepo,
		chapterRepo: chapterRepo,
	}
}

// CreateComment - Tạo comment mới
func (s *commentService) CreateComment(userID, storyID uuid.UUID, chapterID *uuid.UUID, content string) (*models.Comment, error) {
	// Validate content
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, errors.New("nội dung comment không được để trống")
	}

	if len(content) > 2000 {
		return nil, errors.New("nội dung comment quá dài (tối đa 2000 ký tự)")
	}

	// Check story exists
	_, err := s.storyRepo.FindStoryByID(storyID)
	if err != nil {
		return nil, errors.New("truyện không tồn tại")
	}

	// Check chapter exists (if provided)
	if chapterID != nil {
		_, err := s.chapterRepo.FindByID(*chapterID)
		if err != nil {
			return nil, errors.New("chapter không tồn tại")
		}
	}

	comment := &models.Comment{
		UserID:     userID,
		StoryID:    storyID,
		ChapterID:  chapterID,
		Content:    content,
		IsApproved: true, // Auto approve, có thể đổi thành false nếu muốn kiểm duyệt
	}

	if err := s.commentRepo.CreateComment(comment); err != nil {
		return nil, err
	}

	// Fetch lại comment với User preloaded
	createdComment, err := s.commentRepo.FindCommentByID(comment.ID)
	if err != nil {
		return comment, nil // Fallback: trả về comment không có user
	}

	return createdComment, nil
}

// ReplyComment - Trả lời comment
func (s *commentService) ReplyComment(userID, parentID uuid.UUID, content string) (*models.Comment, error) {
	// Validate content
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, errors.New("nội dung reply không được để trống")
	}

	// Get parent comment
	parentComment, err := s.commentRepo.FindCommentByID(parentID)
	if err != nil {
		return nil, errors.New("comment gốc không tồn tại")
	}

	// Don't allow nested replies (only 1 level)
	if parentComment.ParentID != nil {
		return nil, errors.New("không thể reply một reply")
	}

	reply := &models.Comment{
		UserID:     userID,
		StoryID:    parentComment.StoryID,
		ChapterID:  parentComment.ChapterID,
		ParentID:   &parentID,
		Content:    content,
		IsApproved: true,
	}

	if err := s.commentRepo.CreateComment(reply); err != nil {
		return nil, err
	}

	// Fetch lại reply với User preloaded
	createdReply, err := s.commentRepo.FindCommentByID(reply.ID)
	if err != nil {
		return reply, nil // Fallback
	}

	return createdReply, nil
}

// UpdateComment - Chỉnh sửa comment (chỉ owner mới được)
func (s *commentService) UpdateComment(userID, commentID uuid.UUID, content string) (*models.Comment, error) {
	// Validate content
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, errors.New("nội dung comment không được để trống")
	}

	if len(content) > 2000 {
		return nil, errors.New("nội dung comment quá dài (tối đa 2000 ký tự)")
	}

	// Get existing comment
	comment, err := s.commentRepo.FindCommentByID(commentID)
	if err != nil {
		return nil, errors.New("comment không tồn tại")
	}

	// Only owner can edit (not even admin, to prevent abuse)
	if comment.UserID != userID {
		return nil, errors.New("bạn không có quyền chỉnh sửa comment này")
	}

	// Update content
	comment.Content = content
	if err := s.commentRepo.UpdateComment(comment); err != nil {
		return nil, err
	}

	// Fetch lại với User preloaded
	updatedComment, err := s.commentRepo.FindCommentByID(commentID)
	if err != nil {
		return comment, nil
	}

	return updatedComment, nil
}

// DeleteComment - Xóa comment
func (s *commentService) DeleteComment(userID, commentID uuid.UUID, isAdmin bool) error {
	comment, err := s.commentRepo.FindCommentByID(commentID)
	if err != nil {
		return errors.New("comment không tồn tại")
	}

	// Only owner or admin can delete
	if comment.UserID != userID && !isAdmin {
		return errors.New("bạn không có quyền xóa comment này")
	}

	return s.commentRepo.DeleteComment(commentID)
}

// GetCommentsByStory - Lấy comments của truyện
func (s *commentService) GetCommentsByStory(storyID uuid.UUID, page, limit int) ([]models.Comment, int64, error) {
	return s.commentRepo.GetCommentsByStory(storyID, page, limit)
}

// GetCommentsByChapter - Lấy comments của chapter
func (s *commentService) GetCommentsByChapter(chapterID uuid.UUID, page, limit int) ([]models.Comment, int64, error) {
	return s.commentRepo.GetCommentsByChapter(chapterID, page, limit)
}

// UpdateLikeCount - Update cached like count for a comment
func (s *commentService) UpdateLikeCount(commentID uuid.UUID, count int) error {
	return s.commentRepo.UpdateLikeCount(commentID, count)
}

// TogglePin - Ghim/Bỏ ghim bình luận
func (s *commentService) TogglePin(commentID uuid.UUID, isPinned bool) error {
	return s.commentRepo.TogglePin(commentID, isPinned)
}

// FindCommentByID - Tìm bình luận theo ID
func (s *commentService) FindCommentByID(id uuid.UUID) (*models.Comment, error) {
	return s.commentRepo.FindCommentByID(id)
}

