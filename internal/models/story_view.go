package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StoryView struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	StoryID   uuid.UUID  `json:"story_id" gorm:"type:uuid;not null;index"`
	ChapterID *uuid.UUID `json:"chapter_id" gorm:"type:uuid;index"`
	UserID    *uuid.UUID `json:"user_id" gorm:"type:uuid;index"` // NULL = anonymous - Không xác định người dùng
	IPAddress string     `json:"ip_address" gorm:"size:45"`
	ViewedAt  time.Time  `json:"viewed_at" gorm:"type:date;default:CURRENT_DATE;index"`
	ViewCount int        `json:"view_count" gorm:"default:1"`

	// Relations
	Story   Story    `json:"story,omitempty" gorm:"foreignKey:StoryID"`
	Chapter *Chapter `json:"chapter,omitempty" gorm:"foreignKey:ChapterID"`
	User    *User    `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (StoryView) TableName() string {
	return "story_views"
}

func (sv *StoryView) BeforeCreate(tx *gorm.DB) error {
	if sv.ID == uuid.Nil {
		sv.ID = uuid.New()
	}
	return nil
}