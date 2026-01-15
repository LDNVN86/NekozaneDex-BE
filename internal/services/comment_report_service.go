package services

import (
	"errors"
	"nekozanedex/internal/models"
	"nekozanedex/internal/repositories"

	"github.com/google/uuid"
)

type CommentReportService interface {
	ReportComment(userID, commentID uuid.UUID, reason string) (*models.CommentReport, error)
	GetReports(page, limit int, status string) ([]models.CommentReport, int64, error)
	UpdateReportStatus(id uuid.UUID, status string) error
}

type commentReportService struct {
	reportRepo repositories.CommentReportRepository
}

func NewCommentReportService(reportRepo repositories.CommentReportRepository) CommentReportService {
	return &commentReportService{reportRepo: reportRepo}
}

func (s *commentReportService) ReportComment(userID, commentID uuid.UUID, reason string) (*models.CommentReport, error) {
	// Check if user already reported this comment and it's still pending
	exists, err := s.reportRepo.HasUserReported(commentID, userID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("bạn đã báo cáo bình luận này rồi")
	}

	report := &models.CommentReport{
		ID:        uuid.New(),
		CommentID: commentID,
		UserID:    userID,
		Reason:    reason,
		Status:    "pending",
	}

	if err := s.reportRepo.CreateReport(report); err != nil {
		return nil, err
	}

	return report, nil
}

func (s *commentReportService) GetReports(page, limit int, status string) ([]models.CommentReport, int64, error) {
	return s.reportRepo.GetReports(page, limit, status)
}

func (s *commentReportService) UpdateReportStatus(id uuid.UUID, status string) error {
	return s.reportRepo.UpdateReportStatus(id, status)
}
