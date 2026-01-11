package repositories

import (
	"time"

	"nekozanedex/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StoryViewRepository interface {
	HasViewedRecently(storyID uuid.UUID, userID *uuid.UUID, ipAddress string, duration time.Duration) (bool, error)
	RecordView(view *models.StoryView) error
	GetViewStats(storyID uuid.UUID) (int64, error)
}

type storyViewRepository struct {
	db *gorm.DB
}

func NewStoryViewRepository(db *gorm.DB) StoryViewRepository {
	return &storyViewRepository{db: db}
}

// HasViewedRecently checks if user/IP has viewed this story within duration
// Priority: UserID > IPAddress
func (r *storyViewRepository) HasViewedRecently(storyID uuid.UUID, userID *uuid.UUID, ipAddress string, duration time.Duration) (bool, error) {
	var count int64
	cutoff := time.Now().Add(-duration)

	query := r.db.Model(&models.StoryView{}).Where("story_id = ? AND viewed_at > ?", storyID, cutoff)

	// If user is logged in, check by user_id
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	} else {
		// Anonymous user - check by IP
		query = query.Where("user_id IS NULL AND ip_address = ?", ipAddress)
	}

	if err := query.Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *storyViewRepository) RecordView(view *models.StoryView) error {
	view.ViewedAt = time.Now()
	return r.db.Create(view).Error
}

func (r *storyViewRepository) GetViewStats(storyID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.Model(&models.StoryView{}).Where("story_id = ?", storyID).Count(&count).Error
	return count, err
}
