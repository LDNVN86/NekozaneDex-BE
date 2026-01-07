package services

import (
	"errors"

	"nekozanedex/internal/models"
	"nekozanedex/internal/repositories"

	"github.com/google/uuid"
)

type BookmarkService interface {
	AddBookmark(userID, storyID uuid.UUID) error
	RemoveBookmark(userID, storyID uuid.UUID) error
	GetUserBookmarks(userID uuid.UUID, page, limit int) ([]models.BookMark, int64, error)
	IsBookmarked(userID, storyID uuid.UUID) bool
}

type bookmarkService struct {
	bookmarkRepo repositories.BookmarkRepository
	storyRepo    repositories.StoryRepository
}

func NewBookmarkService(
	bookmarkRepo repositories.BookmarkRepository,
	storyRepo repositories.StoryRepository,
) BookmarkService {
	return &bookmarkService{
		bookmarkRepo: bookmarkRepo,
		storyRepo:    storyRepo,
	}
}

// AddBookmark - Thêm bookmark
func (s *bookmarkService) AddBookmark(userID, storyID uuid.UUID) error {
	// Check story exists
	_, err := s.storyRepo.FindStoryByID(storyID)
	if err != nil {
		return errors.New("truyện không tồn tại")
	}

	// Check if already bookmarked
	if s.bookmarkRepo.IsBookmarked(userID, storyID) {
		return errors.New("truyện đã được bookmark")
	}

	bookmark := &models.BookMark{
		UserID:  userID,
		StoryID: storyID,
	}

	return s.bookmarkRepo.CreateBookmark(bookmark)
}

// RemoveBookmark - Xóa bookmark
func (s *bookmarkService) RemoveBookmark(userID, storyID uuid.UUID) error {
	if !s.bookmarkRepo.IsBookmarked(userID, storyID) {
		return errors.New("bookmark không tồn tại")
	}
	return s.bookmarkRepo.DeleteBookmark(userID, storyID)
}

// GetUserBookmarks - Lấy danh sách bookmark của user
func (s *bookmarkService) GetUserBookmarks(userID uuid.UUID, page, limit int) ([]models.BookMark, int64, error) {
	return s.bookmarkRepo.GetBookmarksByUser(userID, page, limit)
}

// IsBookmarked - Kiểm tra đã bookmark chưa
func (s *bookmarkService) IsBookmarked(userID, storyID uuid.UUID) bool {
	return s.bookmarkRepo.IsBookmarked(userID, storyID)
}
