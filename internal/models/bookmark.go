package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BookMark struct {
	ID			uuid.UUID			`json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID		uuid.UUID			`json:"user_id" gorm:"type:uuid;not null;index"`
	StoryID		uuid.UUID			`json:"story_id" gorm:"type:uuid;not null;index"`
	CreatedAt	time.Time			`json:"created_at"`

	//Relations
	User		User				`json:"user,omitempty" gorm:"foreignKey:UserID"`
	Story		Story				`json:"story,omitempty" gorm:"foreignKey:StoryID"`
}

func (BookMark) TableName() string {
	return "bookmarks"
}

func (b *BookMark) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}
