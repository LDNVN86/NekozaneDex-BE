package repositories

import (
	"nekozanedex/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GenreRepository interface {
	CreateGenre(genre *models.Genre) error
	FindGenreByID(id uuid.UUID) (*models.Genre, error)
	FindGenreBySlug(slug string) (*models.Genre, error)
	UpdateGenre(genre *models.Genre) error
	DeleteGenre(id uuid.UUID) error
	GetAllGenres() ([]models.Genre, error)
}

type genreRepository struct {
	db *gorm.DB
}

func NewGenreRepository(db *gorm.DB) GenreRepository {
	return &genreRepository{db: db}
}

// CreateGenre - Tạo Genre
func (r *genreRepository) CreateGenre(genre *models.Genre) error {
	return r.db.Create(genre).Error
}

// FindGenreByID - Tìm Genre theo ID
func (r *genreRepository) FindGenreByID(id uuid.UUID) (*models.Genre, error) {
	var genre models.Genre
	err := r.db.First(&genre, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &genre, nil
}

// FindGenreBySlug - Tìm Genre theo Slug
func (r *genreRepository) FindGenreBySlug(slug string) (*models.Genre, error) {
	var genre models.Genre
	err := r.db.First(&genre, "slug = ?", slug).Error
	if err != nil {
		return nil, err
	}
	return &genre, nil
}

// UpdateGenre - Cập nhật Genre
func (r *genreRepository) UpdateGenre(genre *models.Genre) error {
	return r.db.Save(genre).Error
}

// DeleteGenre - Xóa Genre
func (r *genreRepository) DeleteGenre(id uuid.UUID) error {
	return r.db.Delete(&models.Genre{}, "id = ?", id).Error
}

// GetAllGenres - Lấy tất cả Genre
func (r *genreRepository) GetAllGenres() ([]models.Genre, error) {
	var genres []models.Genre
	err := r.db.Order("name ASC").Find(&genres).Error
	return genres, err
}
