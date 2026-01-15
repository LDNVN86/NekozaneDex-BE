package services

import (
	"nekozanedex/internal/models"
	"nekozanedex/internal/repositories"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserSettingsService interface {
	GetSettings(userID uuid.UUID) (*models.UserSettings, error)
	UpdateSettings(userID uuid.UUID, settings *models.UserSettings) (*models.UserSettings, error)
}

type userSettingsService struct {
	settingsRepo repositories.UserSettingsRepository
}

func NewUserSettingsService(settingsRepo repositories.UserSettingsRepository) UserSettingsService {
	return &userSettingsService{settingsRepo: settingsRepo}
}

// GetSettings - Lấy settings của user, tạo mới nếu chưa có
func (s *userSettingsService) GetSettings(userID uuid.UUID) (*models.UserSettings, error) {
	settings, err := s.settingsRepo.FindByUserID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Tạo settings mặc định
			newSettings := &models.UserSettings{
				UserID:          userID,
				Theme:           "system",
				FontSize:        16,
				FontFamily:      "system",
				LineHeight:      1.8,
				ReadingBg:       "white",
				AutoScrollSpeed: 0,
			}
			if err := s.settingsRepo.Upsert(newSettings); err != nil {
				return nil, err
			}
			return newSettings, nil
		}
		return nil, err
	}
	return settings, nil
}

// UpdateSettings - Cập nhật settings
func (s *userSettingsService) UpdateSettings(userID uuid.UUID, updates *models.UserSettings) (*models.UserSettings, error) {
	settings, err := s.GetSettings(userID)
	if err != nil {
		return nil, err
	}

	// Cập nhật các trường
	if updates.Theme != "" {
		settings.Theme = updates.Theme
	}
	if updates.FontSize > 0 {
		settings.FontSize = updates.FontSize
	}
	if updates.FontFamily != "" {
		settings.FontFamily = updates.FontFamily
	}
	if updates.LineHeight > 0 {
		settings.LineHeight = updates.LineHeight
	}
	if updates.ReadingBg != "" {
		settings.ReadingBg = updates.ReadingBg
	}
	if updates.AutoScrollSpeed >= 0 {
		settings.AutoScrollSpeed = updates.AutoScrollSpeed
	}

	if err := s.settingsRepo.Upsert(settings); err != nil {
		return nil, err
	}

	return settings, nil
}
