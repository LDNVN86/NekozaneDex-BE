package repositories

import (
	"nekozanedex/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CommentRepository interface {
	CreateComment(comment *models.Comment) error
	FindCommentByID(id uuid.UUID) (*models.Comment, error)
	UpdateComment(comment *models.Comment) error
	DeleteComment(id uuid.UUID) error
	GetCommentsByStory(storyID uuid.UUID, page, limit int) ([]models.Comment, int64, error)
	GetCommentsByChapter(chapterID uuid.UUID, page, limit int) ([]models.Comment, int64, error)
	GetCommentReplies(parentID uuid.UUID) ([]models.Comment, error)
	UpdateLikeCount(commentID uuid.UUID, count int) error
	TogglePin(commentID uuid.UUID, isPinned bool) error
}

type commentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) CommentRepository {
	return &commentRepository{db: db}
}

// CreateComment - Tạo Comment
func (r *commentRepository) CreateComment(comment *models.Comment) error {
	return r.db.Create(comment).Error
}

// FindCommentByID - Tìm Comment theo ID
func (r *commentRepository) FindCommentByID(id uuid.UUID) (*models.Comment, error) {
	var comment models.Comment
	err := r.db.Preload("User").Preload("Replies").Preload("Parent.User").First(&comment, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

// UpdateComment - Cập nhật Comment
func (r *commentRepository) UpdateComment(comment *models.Comment) error {
	return r.db.Save(comment).Error
}

// DeleteComment - Xóa Comment
func (r *commentRepository) DeleteComment(id uuid.UUID) error {
	return r.db.Delete(&models.Comment{}, "id = ?", id).Error
}

// GetCommentsByStory - Lấy Comments theo Story (top-level only)
func (r *commentRepository) GetCommentsByStory(storyID uuid.UUID, page, limit int) ([]models.Comment, int64, error) {
	var comments []models.Comment
	var total int64

	query := r.db.Model(&models.Comment{}).
		Where("story_id = ? AND chapter_id IS NULL AND parent_id IS NULL AND is_approved = ?", storyID, true)
	query.Count(&total)

	offset := (page - 1) * limit
	err := r.db.Preload("User").Preload("Replies.User").
		Where("story_id = ? AND chapter_id IS NULL AND parent_id IS NULL AND is_approved = ?", storyID, true).
		Offset(offset).Limit(limit).
		Order("is_pinned DESC, created_at DESC").
		Find(&comments).Error

	return comments, total, err
}

// GetCommentsByChapter - Lấy Comments theo Chapter
func (r *commentRepository) GetCommentsByChapter(chapterID uuid.UUID, page, limit int) ([]models.Comment, int64, error) {
	var comments []models.Comment
	var total int64

	query := r.db.Model(&models.Comment{}).
		Where("chapter_id = ? AND parent_id IS NULL AND is_approved = ?", chapterID, true)
	query.Count(&total)

	offset := (page - 1) * limit
	err := r.db.Preload("User").Preload("Replies.User").
		Where("chapter_id = ? AND parent_id IS NULL AND is_approved = ?", chapterID, true).
		Offset(offset).Limit(limit).
		Order("is_pinned DESC, created_at DESC").
		Find(&comments).Error

	return comments, total, err
}

// GetCommentReplies - Lấy Replies của Comment
func (r *commentRepository) GetCommentReplies(parentID uuid.UUID) ([]models.Comment, error) {
	var replies []models.Comment
	err := r.db.Preload("User").
		Where("parent_id = ? AND is_approved = ?", parentID, true).
		Order("created_at ASC").
		Find(&replies).Error
	return replies, err
}

// UpdateLikeCount - Update cached like count for a comment
func (r *commentRepository) UpdateLikeCount(commentID uuid.UUID, count int) error {
	return r.db.Model(&models.Comment{}).Where("id = ?", commentID).Update("like_count", count).Error
}

// TogglePin - Ghim/Bỏ ghim bình luận
func (r *commentRepository) TogglePin(commentID uuid.UUID, isPinned bool) error {
	return r.db.Model(&models.Comment{}).Where("id = ?", commentID).Update("is_pinned", isPinned).Error
}

