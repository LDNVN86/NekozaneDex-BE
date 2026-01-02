package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReadingHistory struct {
	ID				uuid.UUID			`json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID			uuid.UUID			`json:"user_id" gorm:"type:uuid;not null;index"`
	StoryID			uuid.UUID			`json:"story_id" gorm:"type:uuid;not null;index"`
	ChapterID		uuid.UUID			`json:"chapter_id" gorm:"type:uuid;not null"`
	LastReadAt      time.Time 			`json:"last_read_at" gorm:"default:CURRENT_TIMESTAMP"`
	ScrollPosition  int       			`json:"scroll_position" gorm:"default:0"`

	// Relations
	User    		User    			`json:"user,omitempty" gorm:"foreignKey:UserID"`
	Story   		Story   			`json:"story,omitempty" gorm:"foreignKey:StoryID"`
	Chapter 		Chapter 			`json:"chapter,omitempty" gorm:"foreignKey:ChapterID"`
}

func (ReadingHistory) TableName() string {
	return "reading_history"
}

func (r *ReadingHistory) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}
