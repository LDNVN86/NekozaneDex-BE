package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserSettings struct {
	ID              uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID          uuid.UUID `json:"user_id" gorm:"type:uuid;uniqueIndex;not null"`
	Theme           string    `json:"theme" gorm:"size:20;default:light"`           // light, dark, system
	FontSize        int       `json:"font_size" gorm:"default:16"`
	FontFamily      string    `json:"font_family" gorm:"size:50;default:system"`
	LineHeight      float64   `json:"line_height" gorm:"default:1.8"`
	ReadingBg       string    `json:"reading_bg" gorm:"size:20;default:white"`      // white, sepia, dark
	AutoScrollSpeed int       `json:"auto_scroll_speed" gorm:"default:0"`           // 0 = off
	UpdatedAt       time.Time `json:"updated_at"`

	// Relations
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (UserSettings) TableName() string {
	return "user_settings"
}

func (us *UserSettings) BeforeCreate(tx *gorm.DB) error {
	if us.ID == uuid.Nil {
		us.ID = uuid.New()
	}
	return nil
}