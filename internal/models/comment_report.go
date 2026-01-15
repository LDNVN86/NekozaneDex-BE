package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CommentReport struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CommentID uuid.UUID `gorm:"type:uuid;not null;index" json:"comment_id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Reason    string    `gorm:"type:text;not null" json:"reason"`
	Status    string    `gorm:"type:varchar(20);default:'pending';index" json:"status"` // pending, resolved, dismissed
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relations
	Comment Comment `gorm:"foreignKey:CommentID" json:"comment,omitempty"`
	User    User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (CommentReport) TableName() string {
	return "comment_reports"
}

func (r *CommentReport) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}
