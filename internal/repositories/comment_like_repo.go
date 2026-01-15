package repositories

import (
	"nekozanedex/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CommentLikeRepository interface {
	CreateLike(like *models.CommentLike) error
	DeleteLike(commentID, userID uuid.UUID) error
	HasUserLiked(commentID, userID uuid.UUID) bool
	GetLikeCount(commentID uuid.UUID) int64
	GetUserLikedCommentIDs(userID uuid.UUID, commentIDs []uuid.UUID) []uuid.UUID
}

type commentLikeRepository struct {
	db *gorm.DB
}

func NewCommentLikeRepository(db *gorm.DB) CommentLikeRepository {
	return &commentLikeRepository{db: db}
}

// CreateLike - Create a new like
func (r *commentLikeRepository) CreateLike(like *models.CommentLike) error {
	return r.db.Create(like).Error
}

// DeleteLike - Remove a like
func (r *commentLikeRepository) DeleteLike(commentID, userID uuid.UUID) error {
	return r.db.Where("comment_id = ? AND user_id = ?", commentID, userID).
		Delete(&models.CommentLike{}).Error
}

// HasUserLiked - Check if user has liked a comment
func (r *commentLikeRepository) HasUserLiked(commentID, userID uuid.UUID) bool {
	var count int64
	r.db.Model(&models.CommentLike{}).
		Where("comment_id = ? AND user_id = ?", commentID, userID).
		Count(&count)
	return count > 0
}

// GetLikeCount - Get total likes for a comment
func (r *commentLikeRepository) GetLikeCount(commentID uuid.UUID) int64 {
	var count int64
	r.db.Model(&models.CommentLike{}).
		Where("comment_id = ?", commentID).
		Count(&count)
	return count
}

// GetUserLikedCommentIDs - Get list of comment IDs that user has liked
// Useful for batch checking when loading comments
func (r *commentLikeRepository) GetUserLikedCommentIDs(userID uuid.UUID, commentIDs []uuid.UUID) []uuid.UUID {
	if len(commentIDs) == 0 {
		return []uuid.UUID{}
	}
	
	var likedIDs []uuid.UUID
	r.db.Model(&models.CommentLike{}).
		Select("comment_id").
		Where("user_id = ? AND comment_id IN ?", userID, commentIDs).
		Pluck("comment_id", &likedIDs)
	return likedIDs
}
