package services

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"nekozanedex/internal/models"
	"nekozanedex/internal/repositories"

	"github.com/google/uuid"
)

type ChapterService interface {
	// Public methods
	GetChapterByNumber(storySlug string, chapterNumber int) (*models.Chapter, error)
	GetChaptersByStory(storySlug string) ([]models.Chapter, error)
	GetChaptersByStoryPaginated(storySlug string, page, limit int) ([]models.Chapter, int64, error)

	// Admin methods
	CreateChapter(storyID uuid.UUID, chapter *models.Chapter) error
	UpdateChapter(id uuid.UUID, chapter *models.Chapter) error
	DeleteChapter(id uuid.UUID) error
	GetChapterByID(id uuid.UUID) (*models.Chapter, error)
	GetChaptersByStoryAdmin(storyID uuid.UUID) ([]models.Chapter, error) // All chapters including drafts
	PublishChapter(id uuid.UUID) error
	ScheduleChapter(id uuid.UUID, scheduledAt time.Time) error
	BulkImportChapters(storyID uuid.UUID, chapters []models.Chapter) error
	
	// Scheduler methods
	PublishScheduledChapters() (int, error) // Returns count of published chapters
}

type chapterService struct {
	chapterRepo repositories.ChapterRepository
	storyRepo   repositories.StoryRepository
}

func NewChapterService(
	chapterRepo repositories.ChapterRepository,
	storyRepo repositories.StoryRepository,
) ChapterService {
	return &chapterService{
		chapterRepo: chapterRepo,
		storyRepo:   storyRepo,
	}
}

// CreateChapter - Tạo chapter mới (Admin)
func (s *chapterService) CreateChapter(storyID uuid.UUID, chapter *models.Chapter) error {
	// Validate
	if strings.TrimSpace(chapter.Title) == "" {
		return errors.New("tiêu đề chapter không được để trống")
	}

	// Check story exists
	story, err := s.storyRepo.FindStoryByID(storyID)
	if err != nil {
		return errors.New("truyện không tồn tại")
	}

	// Set chapter info
	chapter.StoryID = storyID
	chapter.ChapterNumber = story.TotalChapters + 1
	
	// Calculate page count from images
	chapter.PageCount = countImages(chapter.Images)
	
	chapter.ViewCount = 0
	chapter.CreatedAt = time.Now()
	chapter.UpdatedAt = time.Now()

	if err := s.chapterRepo.Create(chapter); err != nil {
		return err
	}

	// Update story total chapters
	story.TotalChapters++
	story.UpdatedAt = time.Now()
	return s.storyRepo.UpdateStory(story)
}

// UpdateChapter - Cập nhật chapter (Admin)
func (s *chapterService) UpdateChapter(id uuid.UUID, updatedChapter *models.Chapter) error {
	existingChapter, err := s.chapterRepo.FindByID(id)
	if err != nil {
		return errors.New("chapter không tồn tại")
	}

	if updatedChapter.Title != "" {
		existingChapter.Title = updatedChapter.Title
	}
	// Update new fields
	existingChapter.ChapterLabel = updatedChapter.ChapterLabel
	existingChapter.ChapterType = updatedChapter.ChapterType
	existingChapter.Ordering = updatedChapter.Ordering
	
	// Always update content (can be empty for manga chapters)
	existingChapter.Content = updatedChapter.Content
	if updatedChapter.Images != nil {
		existingChapter.Images = updatedChapter.Images
		existingChapter.PageCount = countImages(updatedChapter.Images)
	}
	existingChapter.UpdatedAt = time.Now()

	return s.chapterRepo.Update(existingChapter)
}

// DeleteChapter - Xóa chapter (Admin)
func (s *chapterService) DeleteChapter(id uuid.UUID) error {
	chapter, err := s.chapterRepo.FindByID(id)
	if err != nil {
		return errors.New("chapter không tồn tại")
	}

	if err := s.chapterRepo.Delete(id); err != nil {
		return err
	}

	// Update story total chapters
	story, _ := s.storyRepo.FindStoryByID(chapter.StoryID)
	if story != nil {
		story.TotalChapters--
		story.UpdatedAt = time.Now()
		_ = s.storyRepo.UpdateStory(story)
	}

	return nil
}

// GetChapterByID - Lấy chapter theo ID (Admin)
func (s *chapterService) GetChapterByID(id uuid.UUID) (*models.Chapter, error) {
	chapter, err := s.chapterRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("chapter không tồn tại")
	}
	return chapter, nil
}

// GetChapterByNumber - Lấy chapter theo số (Public)
func (s *chapterService) GetChapterByNumber(storySlug string, chapterNumber int) (*models.Chapter, error) {
	story, err := s.storyRepo.FindStoryBySlug(storySlug)
	if err != nil {
		return nil, errors.New("truyện không tồn tại")
	}

	chapter, err := s.chapterRepo.FindByStoryAndNumber(story.ID, chapterNumber)
	if err != nil {
		return nil, errors.New("chapter không tồn tại")
	}

	// Increment view count
	_ = s.chapterRepo.IncrementViewCount(chapter.ID)

	return chapter, nil
}

// GetChaptersByStory - Lấy danh sách chapters của truyện (Public - only published)
func (s *chapterService) GetChaptersByStory(storySlug string) ([]models.Chapter, error) {
	story, err := s.storyRepo.FindStoryBySlug(storySlug)
	if err != nil {
		return nil, errors.New("truyện không tồn tại")
	}

	return s.chapterRepo.GetByStory(story.ID, true)
}

// GetChaptersByStoryPaginated - Lấy chapters với phân trang (Public)
func (s *chapterService) GetChaptersByStoryPaginated(storySlug string, page, limit int) ([]models.Chapter, int64, error) {
	story, err := s.storyRepo.FindStoryBySlug(storySlug)
	if err != nil {
		return nil, 0, errors.New("truyện không tồn tại")
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit
	return s.chapterRepo.GetByStoryPaginated(story.ID, true, offset, limit)
}

// GetChaptersByStoryAdmin - Lấy tất cả chapters (Admin - including drafts)
func (s *chapterService) GetChaptersByStoryAdmin(storyID uuid.UUID) ([]models.Chapter, error) {
	_, err := s.storyRepo.FindStoryByID(storyID)
	if err != nil {
		return nil, errors.New("truyện không tồn tại")
	}

	return s.chapterRepo.GetByStory(storyID, false) // false = all chapters
}

// PublishChapter - Xuất bản chapter (Admin)
func (s *chapterService) PublishChapter(id uuid.UUID) error {
	chapter, err := s.chapterRepo.FindByID(id)
	if err != nil {
		return errors.New("chapter không tồn tại")
	}

	now := time.Now()
	chapter.IsPublished = true
	chapter.PublishedAt = &now
	chapter.UpdatedAt = now

	return s.chapterRepo.Update(chapter)
}

// ScheduleChapter - Hẹn giờ xuất bản (Admin)
func (s *chapterService) ScheduleChapter(id uuid.UUID, scheduledAt time.Time) error {
	chapter, err := s.chapterRepo.FindByID(id)
	if err != nil {
		return errors.New("chapter không tồn tại")
	}

	if scheduledAt.Before(time.Now()) {
		return errors.New("thời gian hẹn phải trong tương lai")
	}

	chapter.ScheduledAt = &scheduledAt
	chapter.UpdatedAt = time.Now()

	return s.chapterRepo.Update(chapter)
}

// BulkImportChapters - Import nhiều chapters cùng lúc (Admin)
func (s *chapterService) BulkImportChapters(storyID uuid.UUID, chapters []models.Chapter) error {
	story, err := s.storyRepo.FindStoryByID(storyID)
	if err != nil {
		return errors.New("truyện không tồn tại")
	}

	startNumber := story.TotalChapters + 1

	for i := range chapters {
		chapters[i].StoryID = storyID
		chapters[i].ChapterNumber = startNumber + i
		chapters[i].PageCount = countImages(chapters[i].Images)
		chapters[i].CreatedAt = time.Now()
		chapters[i].UpdatedAt = time.Now()

		if err := s.chapterRepo.Create(&chapters[i]); err != nil {
			return err
		}
	}

	// Update story total
	story.TotalChapters += len(chapters)
	story.UpdatedAt = time.Now()
	return s.storyRepo.UpdateStory(story)
}

// PublishScheduledChapters - Auto-publish chapters that have reached their scheduled time
func (s *chapterService) PublishScheduledChapters() (int, error) {
	chapters, err := s.chapterRepo.GetScheduledChapters()
	if err != nil {
		return 0, err
	}

	count := 0
	now := time.Now()
	for _, chapter := range chapters {
		chapter.IsPublished = true
		chapter.PublishedAt = &now
		chapter.ScheduledAt = nil // Clear scheduled time
		chapter.UpdatedAt = now

		if err := s.chapterRepo.Update(&chapter); err != nil {
			continue // Log error but continue with other chapters
		}
		count++
	}

	return count, nil
}

// Helper function to count images from JSON
func countImages(imagesJSON []byte) int {
	if imagesJSON == nil {
		return 0
	}
	var images []string
	if err := json.Unmarshal(imagesJSON, &images); err != nil {
		return 0
	}
	return len(images)
}
