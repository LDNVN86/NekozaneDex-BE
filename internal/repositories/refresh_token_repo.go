package repositories

import (
	"time"

	"nekozanedex/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RefreshTokenRepository interface {
	Create(token *models.RefreshToken) error
	FindByHash(tokenHash string) (*models.RefreshToken, error)
	Revoke(id uuid.UUID) error
	RevokeByHash(tokenHash string) error
	RevokeAllByUser(userID uuid.UUID) error
	DeleteExpired() error
	GetActiveByUser(userID uuid.UUID) ([]models.RefreshToken, error)
}

type refreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

// Create - Tạo refresh token mới
func (r *refreshTokenRepository) Create(token *models.RefreshToken) error {
	return r.db.Create(token).Error
}

// FindByHash - Tìm token theo hash
func (r *refreshTokenRepository) FindByHash(tokenHash string) (*models.RefreshToken, error) {
	var token models.RefreshToken
	err := r.db.First(&token, "token_hash = ?", tokenHash).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// Revoke - Revoke token theo ID
func (r *refreshTokenRepository) Revoke(id uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&models.RefreshToken{}).
		Where("id = ?", id).
		Update("revoked_at", now).Error
}

// RevokeByHash - Revoke token theo hash
func (r *refreshTokenRepository) RevokeByHash(tokenHash string) error {
	now := time.Now()
	return r.db.Model(&models.RefreshToken{}).
		Where("token_hash = ?", tokenHash).
		Update("revoked_at", now).Error
}

// RevokeAllByUser - Revoke tất cả tokens của user (dùng khi đổi password, logout all)
func (r *refreshTokenRepository) RevokeAllByUser(userID uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&models.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", now).Error
}

// DeleteExpired - Xóa các tokens đã hết hạn (cleanup job)
func (r *refreshTokenRepository) DeleteExpired() error {
	return r.db.Where("expires_at < ?", time.Now()).
		Delete(&models.RefreshToken{}).Error
}

// GetActiveByUser - Lấy tất cả active tokens của user (hiển thị sessions)
func (r *refreshTokenRepository) GetActiveByUser(userID uuid.UUID) ([]models.RefreshToken, error) {
	var tokens []models.RefreshToken
	err := r.db.Where("user_id = ? AND revoked_at IS NULL AND expires_at > ?", userID, time.Now()).
		Order("created_at DESC").
		Find(&tokens).Error
	return tokens, err
}
