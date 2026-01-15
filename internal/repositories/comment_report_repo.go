package repositories

import (
	"nekozanedex/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CommentReportRepository interface {
	CreateReport(report *models.CommentReport) error
	FindReportByID(id uuid.UUID) (*models.CommentReport, error)
	GetReports(page, limit int, status string) ([]models.CommentReport, int64, error)
	UpdateReportStatus(id uuid.UUID, status string) error
	HasUserReported(commentID, userID uuid.UUID) (bool, error)
}

type commentReportRepository struct {
	db *gorm.DB
}

func NewCommentReportRepository(db *gorm.DB) CommentReportRepository {
	return &commentReportRepository{db: db}
}

func (r *commentReportRepository) CreateReport(report *models.CommentReport) error {
	return r.db.Create(report).Error
}

func (r *commentReportRepository) FindReportByID(id uuid.UUID) (*models.CommentReport, error) {
	var report models.CommentReport
	err := r.db.Preload("Comment").Preload("User").First(&report, "id = ?", id).Error
	return &report, err
}

func (r *commentReportRepository) GetReports(page, limit int, status string) ([]models.CommentReport, int64, error) {
	var reports []models.CommentReport
	var total int64
	offset := (page - 1) * limit

	query := r.db.Model(&models.CommentReport{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Preload("Comment").Preload("User").Preload("Comment.User").
		Offset(offset).Limit(limit).
		Order("created_at DESC").
		Find(&reports).Error

	return reports, total, err
}

func (r *commentReportRepository) UpdateReportStatus(id uuid.UUID, status string) error {
	return r.db.Model(&models.CommentReport{}).Where("id = ?", id).Update("status", status).Error
}

func (r *commentReportRepository) HasUserReported(commentID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&models.CommentReport{}).
		Where("comment_id = ? AND user_id = ? AND status = 'pending'", commentID, userID).
		Count(&count).Error
	return count > 0, err
}
