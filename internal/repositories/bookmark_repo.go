package repositories

import (
	"nekozanedex/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BookmarkRepository interface {
	CreateBookmark(bookmark *models.BookMark) error
	DeleteBookmark(userID, storyID uuid.UUID) error
	FindBookmarkByUserAndStory(userID, storyID uuid.UUID) (*models.BookMark, error)
	GetBookmarksByUser(userID uuid.UUID, page, limit int) ([]models.BookMark, int64, error)
	IsBookmarked(userID, storyID uuid.UUID) bool
}

type bookmarkRepository struct {
	db *gorm.DB
}

func NewBookmarkRepository(db *gorm.DB) BookmarkRepository {
	return &bookmarkRepository{db: db}
}

// CreateBookmark - Tạo Bookmark
func (r *bookmarkRepository) CreateBookmark(bookmark *models.BookMark) error {
	return r.db.Create(bookmark).Error
}

// DeleteBookmark - Xóa Bookmark
func (r *bookmarkRepository) DeleteBookmark(userID, storyID uuid.UUID) error {
	return r.db.Where("user_id = ? AND story_id = ?", userID, storyID).
		Delete(&models.BookMark{}).Error
}

// FindBookmarkByUserAndStory - Tìm Bookmark theo User và Story
func (r *bookmarkRepository) FindBookmarkByUserAndStory(userID, storyID uuid.UUID) (*models.BookMark, error) {
	var bookmark models.BookMark
	err := r.db.First(&bookmark, "user_id = ? AND story_id = ?", userID, storyID).Error
	if err != nil {
		return nil, err
	}
	return &bookmark, nil
}

// GetBookmarksByUser - Lấy Bookmark theo User
func (r *bookmarkRepository) GetBookmarksByUser(userID uuid.UUID, page, limit int) ([]models.BookMark, int64, error) {
	var bookmarks []models.BookMark
	var total int64

	r.db.Model(&models.BookMark{}).Where("user_id = ?", userID).Count(&total)

	offset := (page - 1) * limit
	err := r.db.Preload("Story").Preload("Story.Genres").
		Where("user_id = ?", userID).
		Offset(offset).Limit(limit).
		Order("created_at DESC").
		Find(&bookmarks).Error

	return bookmarks, total, err
}

// IsBookmarked - Kiểm tra đã Bookmark chưa
func (r *bookmarkRepository) IsBookmarked(userID, storyID uuid.UUID) bool {
	var count int64
	r.db.Model(&models.BookMark{}).Where("user_id = ? AND story_id = ?", userID, storyID).Count(&count)
	return count > 0
}
