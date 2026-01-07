package services

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"nekozanedex/internal/models"
	"nekozanedex/internal/repositories"

	"github.com/google/uuid"
	"github.com/gosimple/slug"
)

type StoryService interface {
	// Public methods
	GetStoryBySlug(slug string) (*models.Story, error)
	GetAllStories(page, limit int) ([]models.Story, int64, error)
	GetStoriesByGenre(genreSlug string, page, limit int) ([]models.Story, int64, error)
	GetLatestStories(limit int) ([]models.Story, error)
	GetHotStories(limit int) ([]models.Story, error)
	SearchStories(query string, page, limit int) ([]models.Story, int64, error)
	GetRandomStory() (*models.Story, error)

	// Admin methods
	CreateStory(story *models.Story) error
	UpdateStory(id uuid.UUID, story *models.Story) error
	DeleteStory(id uuid.UUID) error
	GetStoryByID(id uuid.UUID) (*models.Story, error)
	GetAllStoriesAdmin(page, limit int) ([]models.Story, int64, error)
}

type storyService struct {
	storyRepo repositories.StoryRepository
	genreRepo repositories.GenreRepository
}

func NewStoryService(
	storyRepo repositories.StoryRepository,
	genreRepo repositories.GenreRepository,
) StoryService {
	return &storyService{
		storyRepo: storyRepo,
		genreRepo: genreRepo,
	}
}

// CreateStory - Tạo truyện mới (Admin)
func (s *storyService) CreateStory(story *models.Story) error {
	// Validate title
	if strings.TrimSpace(story.Title) == "" {
		return errors.New("tiêu đề không được để trống")
	}

	// Generate slug từ title
	story.Slug = s.generateUniqueSlug(story.Title)

	// Set defaults
	story.ViewCount = 0
	story.TotalChapters = 0
	story.CreatedAt = time.Now()
	story.UpdatedAt = time.Now()

	return s.storyRepo.CreateStory(story)
}

// UpdateStory - Cập nhật truyện (Admin)
func (s *storyService) UpdateStory(id uuid.UUID, updatedStory *models.Story) error {
	existingStory, err := s.storyRepo.FindStoryByID(id)
	if err != nil {
		return errors.New("truyện không tồn tại")
	}

	// Update fields
	if updatedStory.Title != "" {
		existingStory.Title = updatedStory.Title
		// Regenerate slug if title changed
		existingStory.Slug = s.generateUniqueSlug(updatedStory.Title)
	}
	if updatedStory.Description != nil {
		existingStory.Description = updatedStory.Description
	}
	if updatedStory.CoverImageURL != nil {
		existingStory.CoverImageURL = updatedStory.CoverImageURL
	}
	if updatedStory.Status != "" {
		existingStory.Status = updatedStory.Status
	}
	existingStory.IsPublished = updatedStory.IsPublished
	existingStory.UpdatedAt = time.Now()

	return s.storyRepo.UpdateStory(existingStory)
}

// DeleteStory - Xóa truyện (Admin)
func (s *storyService) DeleteStory(id uuid.UUID) error {
	_, err := s.storyRepo.FindStoryByID(id)
	if err != nil {
		return errors.New("truyện không tồn tại")
	}
	return s.storyRepo.DeleteStory(id)
}

// GetStoryByID - Lấy truyện theo ID (Admin)
func (s *storyService) GetStoryByID(id uuid.UUID) (*models.Story, error) {
	story, err := s.storyRepo.FindStoryByID(id)
	if err != nil {
		return nil, errors.New("truyện không tồn tại")
	}
	return story, nil
}

// GetStoryBySlug - Lấy truyện theo slug (Public)
func (s *storyService) GetStoryBySlug(storySlug string) (*models.Story, error) {
	story, err := s.storyRepo.FindStoryBySlug(storySlug)
	if err != nil {
		return nil, errors.New("truyện không tồn tại")
	}

	// Increment view count
	_ = s.storyRepo.IncrementViewCountStory(story.ID)

	return story, nil
}

// GetAllStories - Lấy tất cả truyện đã publish (Public)
func (s *storyService) GetAllStories(page, limit int) ([]models.Story, int64, error) {
	return s.storyRepo.GetAllStories(page, limit, true)
}

// GetAllStoriesAdmin - Lấy tất cả truyện (Admin)
func (s *storyService) GetAllStoriesAdmin(page, limit int) ([]models.Story, int64, error) {
	return s.storyRepo.GetAllStories(page, limit, false)
}

// GetStoriesByGenre - Lấy truyện theo thể loại (Public)
func (s *storyService) GetStoriesByGenre(genreSlug string, page, limit int) ([]models.Story, int64, error) {
	genre, err := s.genreRepo.FindGenreBySlug(genreSlug)
	if err != nil {
		return nil, 0, errors.New("thể loại không tồn tại")
	}
	return s.storyRepo.GetStoriesByGenre(genre.ID, page, limit)
}

// GetLatestStories - Lấy truyện mới cập nhật (Public)
func (s *storyService) GetLatestStories(limit int) ([]models.Story, error) {
	return s.storyRepo.GetStoriesLatest(limit)
}

// GetHotStories - Lấy truyện hot (Public)
func (s *storyService) GetHotStories(limit int) ([]models.Story, error) {
	return s.storyRepo.GetStoriesHot(limit)
}

// SearchStories - Tìm kiếm truyện (Public)
func (s *storyService) SearchStories(query string, page, limit int) ([]models.Story, int64, error) {
	if strings.TrimSpace(query) == "" {
		return nil, 0, errors.New("từ khóa tìm kiếm không được để trống")
	}
	return s.storyRepo.SearchStories(query, page, limit)
}

// GetRandomStory - Lấy truyện ngẫu nhiên (Public)
func (s *storyService) GetRandomStory() (*models.Story, error) {
	stories, _, err := s.storyRepo.GetAllStories(1, 100, true)
	if err != nil || len(stories) == 0 {
		return nil, errors.New("không có truyện nào")
	}

	// Random pick
	randomIndex := time.Now().UnixNano() % int64(len(stories))
	return &stories[randomIndex], nil
}

// Helper: Generate unique slug
func (s *storyService) generateUniqueSlug(title string) string {
	baseSlug := slug.Make(title)

	// Remove special characters
	reg := regexp.MustCompile("[^a-zA-Z0-9-]+")
	baseSlug = reg.ReplaceAllString(baseSlug, "")

	// Check if slug exists
	_, err := s.storyRepo.FindStoryBySlug(baseSlug)
	if err != nil {
		// Slug doesn't exist, use it
		return baseSlug
	}

	// Slug exists, append timestamp
	return baseSlug + "-" + uuid.New().String()[:8]
}
