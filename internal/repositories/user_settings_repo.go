package repositories

import (
	"nekozanedex/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserSettingsRepository interface {
	FindByUserID(userID uuid.UUID) (*models.UserSettings, error)
	Upsert(settings *models.UserSettings) error
}

type userSettingsRepository struct {
	db *gorm.DB
}

func NewUserSettingsRepository(db *gorm.DB) UserSettingsRepository {
	return &userSettingsRepository{db: db}
}

// FindByUserID - Lấy settings theo UserID
func (r *userSettingsRepository) FindByUserID(userID uuid.UUID) (*models.UserSettings, error) {
	var settings models.UserSettings
	err := r.db.First(&settings, "user_id = ?", userID).Error
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

// Upsert - Tạo hoặc cập nhật settings
func (r *userSettingsRepository) Upsert(settings *models.UserSettings) error {
	return r.db.Save(settings).Error
}
