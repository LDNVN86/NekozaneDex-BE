package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TypoReport struct {
	ID            uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID        *uuid.UUID `json:"user_id" gorm:"type:uuid;index"` // NULL = anonymous - Không xác định người dùng
	ChapterID     uuid.UUID  `json:"chapter_id" gorm:"type:uuid;not null;index"`
	OriginalText  string     `json:"original_text" gorm:"type:text;not null"`
	SuggestedText *string    `json:"suggested_text" gorm:"type:text"`
	PositionHint  *string    `json:"position_hint"` // Vị trí gợi ý
	Status        string     `json:"status" gorm:"size:20;default:pending;index"` // pending, fixed, rejected
	CreatedAt     time.Time  `json:"created_at"`

	// Relations
	User    *User   `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Chapter Chapter `json:"chapter,omitempty" gorm:"foreignKey:ChapterID"`
}

func (TypoReport) TableName() string {
	return "typo_reports"
}

func (t *TypoReport) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}