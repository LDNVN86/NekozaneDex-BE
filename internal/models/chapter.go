package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Chapter struct {
	ID            uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	StoryID       uuid.UUID      `json:"story_id" gorm:"type:uuid;not null;index"`
	ChapterNumber int            `json:"chapter_number" gorm:"not null"`
	Title         string         `json:"title" gorm:"not null;size:255"`
	Content       string         `json:"content" gorm:"type:text;default:''"` // Text content (for novel-style)
	Images        datatypes.JSON `json:"images" gorm:"type:jsonb"`            // []string - URLs of manga pages
	PageCount     int            `json:"page_count" gorm:"default:0"`
	IsPublished   bool           `json:"is_published" gorm:"default:false"`
	PublishedAt   *time.Time     `json:"published_at"`
	ScheduledAt   *time.Time     `json:"scheduled_at"` // Scheduled publishing
	ViewCount     int64          `json:"view_count" gorm:"default:0"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
	// Relations
	Story    Story     `json:"story,omitempty" gorm:"foreignKey:StoryID"`
	Comments []Comment `json:"comments,omitempty" gorm:"foreignKey:ChapterID"`
}

// TableName - custom table name
func (Chapter) TableName() string {
	return "chapters"
}

func (c *Chapter) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// GetImagesSlice - Helper to get images as []string
func (c *Chapter) GetImagesSlice() []string {
	var images []string
	if c.Images != nil {
		_ = json.Unmarshal(c.Images, &images)
	}
	return images
}

// SetImagesFromSlice - Helper to set images from []string
func (c *Chapter) SetImagesFromSlice(images []string) error {
	data, err := json.Marshal(images)
	if err != nil {
		return err
	}
	c.Images = data
	c.PageCount = len(images)
	return nil
}