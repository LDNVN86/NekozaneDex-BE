package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Comment struct {
	ID         uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID     uuid.UUID      `json:"user_id" gorm:"type:uuid;not null;index"`
	StoryID    uuid.UUID      `json:"story_id" gorm:"type:uuid;not null;index"`
	ChapterID  *uuid.UUID     `json:"chapter_id" gorm:"type:uuid;index"` // NULL = comment on story
	ParentID   *uuid.UUID     `json:"parent_id" gorm:"type:uuid;index"`  // Reply to comment
	Content    string         `json:"content" gorm:"type:text;not null"`
	IsApproved bool           `json:"is_approved" gorm:"default:true"`
	IsPinned   bool           `json:"is_pinned" gorm:"default:false"` // Admin can pin
	LikeCount  int            `json:"like_count" gorm:"default:0"`    // Cached like count
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	User    User          `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Story   Story         `json:"story,omitempty" gorm:"foreignKey:StoryID"`
	Chapter *Chapter      `json:"chapter,omitempty" gorm:"foreignKey:ChapterID"`
	Parent  *Comment      `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Replies []Comment     `json:"replies,omitempty" gorm:"foreignKey:ParentID"`
	Likes   []CommentLike `json:"-" gorm:"foreignKey:CommentID"` // Hidden from JSON
}

func (Comment) TableName() string {
	return "comments"
}

func (c *Comment) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}