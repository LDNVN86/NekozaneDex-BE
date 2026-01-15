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
	RecordStoryView(storyID uuid.UUID, userID *uuid.UUID, ipAddress string) error // Fair view counting
	GetAllStories(page, limit int) ([]models.Story, int64, error)
	GetStoriesByGenre(genreSlug string, page, limit int) ([]models.Story, int64, error)
	GetLatestStories(limit int) ([]models.Story, error)
	GetHotStories(limit int) ([]models.Story, error)
	SearchStories(query string, page, limit int) ([]models.Story, int64, error)
	GetRandomStory() (*models.Story, error)
	GetAllGenres() ([]models.Genre, error)

	// Admin methods
	CreateStory(story *models.Story) error
	UpdateStory(id uuid.UUID, story *models.Story) error
	UpdateStoryGenres(storyID uuid.UUID, genreIDs []string) error
	DeleteStory(id uuid.UUID) error
	GetStoryByID(id uuid.UUID) (*models.Story, error)
	GetAllStoriesAdmin(page, limit int) ([]models.Story, int64, error)
	SearchStoriesAdmin(query string, page, limit int) ([]models.Story, int64, error)
}

type storyService struct {
	storyRepo     repositories.StoryRepository
	genreRepo     repositories.GenreRepository
	storyViewRepo repositories.StoryViewRepository
	uploadService UploadService
}

func NewStoryService(
	storyRepo repositories.StoryRepository,
	genreRepo repositories.GenreRepository,
	storyViewRepo repositories.StoryViewRepository,
	uploadService UploadService,
) StoryService {
	return &storyService{
		storyRepo:     storyRepo,
		genreRepo:     genreRepo,
		storyViewRepo: storyViewRepo,
		uploadService: uploadService,
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
		if updatedStory.Title != existingStory.Title {
			existingStory.Title = updatedStory.Title
			existingStory.Slug = s.generateUniqueSlug(updatedStory.Title)
		}
	}
	
	// New metadata fields
	existingStory.OriginalTitle = updatedStory.OriginalTitle
	existingStory.AltTitles = updatedStory.AltTitles
	existingStory.AuthorName = updatedStory.AuthorName
	existingStory.ArtistName = updatedStory.ArtistName
	existingStory.Translator = updatedStory.Translator
	existingStory.SourceURL = updatedStory.SourceURL
	existingStory.SourceName = updatedStory.SourceName
	existingStory.Country = updatedStory.Country
	existingStory.ReleaseYear = updatedStory.ReleaseYear
	existingStory.EndYear = updatedStory.EndYear
	
	if updatedStory.Description != nil {
		existingStory.Description = updatedStory.Description
	}
	if updatedStory.CoverImageURL != nil {
		// Delete old cover image if it exists and is different from the new one
		if existingStory.CoverImageURL != nil &&
			*existingStory.CoverImageURL != "" &&
			*existingStory.CoverImageURL != *updatedStory.CoverImageURL {
			
			// Debug: Log old and new URLs
			println("[StoryService] Old cover URL:", *existingStory.CoverImageURL)
			println("[StoryService] New cover URL:", *updatedStory.CoverImageURL)
			
			oldPublicID := extractCloudinaryPublicID(*existingStory.CoverImageURL)
			println("[StoryService] Extracted public_id:", oldPublicID)
			
			if oldPublicID != "" && s.uploadService != nil {
				// Delete asynchronously, don't block update if delete fails
				go func(publicID string) {
					println("[StoryService] Attempting to delete:", publicID)
					if err := s.uploadService.DeleteImage(publicID); err != nil {
						println("[StoryService] Failed to delete old cover:", err.Error())
					} else {
						println("[StoryService] Successfully deleted old cover:", publicID)
					}
				}(oldPublicID)
			} else {
				if oldPublicID == "" {
					println("[StoryService] WARNING: Could not extract public_id from URL")
				}
				if s.uploadService == nil {
					println("[StoryService] WARNING: uploadService is nil")
				}
			}
		}
		existingStory.CoverImageURL = updatedStory.CoverImageURL
	}
	if updatedStory.Status != "" {
		existingStory.Status = updatedStory.Status
	}
	existingStory.IsPublished = updatedStory.IsPublished
	existingStory.UpdatedAt = time.Now()

	return s.storyRepo.UpdateStory(existingStory)
}

// UpdateStoryGenres - Cập nhật thể loại của truyện (Admin)
func (s *storyService) UpdateStoryGenres(storyID uuid.UUID, genreIDStrings []string) error {
	// Convert string IDs to UUIDs
	genreIDs := make([]uuid.UUID, 0, len(genreIDStrings))
	for _, idStr := range genreIDStrings {
		id, err := uuid.Parse(idStr)
		if err != nil {
			continue // Skip invalid UUIDs
		}
		genreIDs = append(genreIDs, id)
	}
	
	return s.storyRepo.UpdateStoryGenres(storyID, genreIDs)
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
// Note: Does NOT increment view count - call RecordStoryView separately
func (s *storyService) GetStoryBySlug(storySlug string) (*models.Story, error) {
	story, err := s.storyRepo.FindStoryBySlug(storySlug)
	if err != nil {
		return nil, errors.New("truyện không tồn tại")
	}
	return story, nil
}

// RecordStoryView - Fair view counting (1 view per user/IP per 24h)
func (s *storyService) RecordStoryView(storyID uuid.UUID, userID *uuid.UUID, ipAddress string) error {
	// Check if already viewed in last 24 hours
	hasViewed, err := s.storyViewRepo.HasViewedRecently(storyID, userID, ipAddress, 24*time.Hour)
	if err != nil {
		return err
	}

	if hasViewed {
		// Already counted, skip
		return nil
	}

	// Record the view
	view := &models.StoryView{
		StoryID:   storyID,
		UserID:    userID,
		IPAddress: ipAddress,
	}
	if err := s.storyViewRepo.RecordView(view); err != nil {
		return err
	}

	// Increment the story's view count
	return s.storyRepo.IncrementViewCountStory(storyID)
}

// GetAllStories - Lấy tất cả truyện đã publish (Public)
func (s *storyService) GetAllStories(page, limit int) ([]models.Story, int64, error) {
	return s.storyRepo.GetAllStories(page, limit, true)
}

// GetAllStoriesAdmin - Lấy tất cả truyện (Admin)
func (s *storyService) GetAllStoriesAdmin(page, limit int) ([]models.Story, int64, error) {
	return s.storyRepo.GetAllStories(page, limit, false)
}

// SearchStoriesAdmin - Tìm kiếm truyện (Admin - includes drafts)
func (s *storyService) SearchStoriesAdmin(query string, page, limit int) ([]models.Story, int64, error) {
	if strings.TrimSpace(query) == "" {
		return s.storyRepo.GetAllStories(page, limit, false)
	}
	return s.storyRepo.SearchStoriesAdmin(query, page, limit)
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

// GetAllGenres - Lấy tất cả thể loại (Public)
func (s *storyService) GetAllGenres() ([]models.Genre, error) {
	return s.genreRepo.GetAllGenres()
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

// Helper: Extract Cloudinary public ID from URL
// Cloudinary URLs format: https://res.cloudinary.com/{cloud}/image/upload/{transformations}/{version}/{folder}/{filename}
// public_id = folder/filename (without extension)
func extractCloudinaryPublicID(url string) string {
	if url == "" {
		return ""
	}

	// Find "upload/" in URL
	uploadIdx := strings.Index(url, "/upload/")
	if uploadIdx == -1 {
		return ""
	}

	// Get everything after "upload/"
	afterUpload := url[uploadIdx+8:]

	// Split by "/"
	parts := strings.Split(afterUpload, "/")
	if len(parts) < 2 {
		return ""
	}

	// Skip version (starts with "v") and transformations (contains "_" like "f_webp")
	var publicIDParts []string
	for _, part := range parts {
		// Skip version numbers (v123456789)
		if len(part) > 1 && part[0] == 'v' {
			isVersion := true
			for _, c := range part[1:] {
				if c < '0' || c > '9' {
					isVersion = false
					break
				}
			}
			if isVersion {
				continue
			}
		}
		// Skip transformations (contain specific patterns)
		if strings.Contains(part, "_") && (strings.HasPrefix(part, "f_") ||
			strings.HasPrefix(part, "q_") ||
			strings.HasPrefix(part, "w_") ||
			strings.HasPrefix(part, "h_")) {
			continue
		}
		publicIDParts = append(publicIDParts, part)
	}

	if len(publicIDParts) == 0 {
		return ""
	}

	// Join and remove file extension
	publicID := strings.Join(publicIDParts, "/")
	if lastDot := strings.LastIndex(publicID, "."); lastDot != -1 {
		publicID = publicID[:lastDot]
	}

	return publicID
}

