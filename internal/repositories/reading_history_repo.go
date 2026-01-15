package repositories

import (
	"nekozanedex/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ReadingHistoryRepository interface {
	Upsert(history *models.ReadingHistory) error
	GetByUser(userID uuid.UUID, page, limit int) ([]models.ReadingHistory, int64, error)
	GetContinueReading(userID uuid.UUID, limit int) ([]models.ReadingHistory, error)
	GetByUserAndStory(userID, storyID uuid.UUID) (*models.ReadingHistory, error)
	DeleteByStory(userID, storyID uuid.UUID) error
	DeleteAll(userID uuid.UUID) error
}

type readingHistoryRepository struct {
	db *gorm.DB
}

func NewReadingHistoryRepository(db *gorm.DB) ReadingHistoryRepository {
	return &readingHistoryRepository{db: db}
}

func (r *readingHistoryRepository) Upsert(history *models.ReadingHistory) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "story_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"chapter_id", "last_read_at", "scroll_position"}),
	}).Create(history).Error
}

func (r *readingHistoryRepository) GetByUser(userID uuid.UUID, page, limit int) ([]models.ReadingHistory, int64, error) {
	var histories []models.ReadingHistory
	var total int64

	r.db.Model(&models.ReadingHistory{}).Where("user_id = ?", userID).Count(&total)

	offset := (page - 1) * limit
	err := r.db.
		Where("user_id = ?", userID).
		Preload("Story").
		Preload("Chapter").
		Order("last_read_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&histories).Error

	if err != nil {
		return nil, 0, err
	}
	return histories, total, nil
}

func (r *readingHistoryRepository) GetContinueReading(userID uuid.UUID, limit int) ([]models.ReadingHistory, error) {
	var histories []models.ReadingHistory
	err := r.db.
		Where("user_id = ?", userID).
		Preload("Story").
		Preload("Chapter").
		Order("last_read_at DESC").
		Limit(limit).
		Find(&histories).Error

	return histories, err
}

func (r *readingHistoryRepository) GetByUserAndStory(userID, storyID uuid.UUID) (*models.ReadingHistory, error) {
	var history models.ReadingHistory
	err := r.db.
		Where("user_id = ? AND story_id = ?", userID, storyID).
		Preload("Chapter").
		First(&history).Error

	if err != nil {
		return nil, err
	}
	return &history, nil
}

func (r *readingHistoryRepository) DeleteByStory(userID, storyID uuid.UUID) error {
	return r.db.
		Where("user_id = ? AND story_id = ?", userID, storyID).
		Delete(&models.ReadingHistory{}).Error
}

func (r *readingHistoryRepository) DeleteAll(userID uuid.UUID) error {
	return r.db.
		Where("user_id = ?", userID).
		Delete(&models.ReadingHistory{}).Error
}
