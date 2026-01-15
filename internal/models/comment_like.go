package models

import (
	"time"

	"github.com/google/uuid"
)

// CommentLike represents a like on a comment
type CommentLike struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CommentID uuid.UUID `json:"comment_id" gorm:"type:uuid;not null;uniqueIndex:idx_comment_user_like"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null;uniqueIndex:idx_comment_user_like"`
	CreatedAt time.Time `json:"created_at"`
}

func (CommentLike) TableName() string {
	return "comment_likes"
}

func (cl *CommentLike) BeforeCreate(tx interface{}) error {
	if cl.ID == uuid.Nil {
		cl.ID = uuid.New()
	}
	return nil
}
