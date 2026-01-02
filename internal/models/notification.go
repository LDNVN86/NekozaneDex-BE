package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Notification struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null;index"`
	Type      string    `json:"type" gorm:"size:50;not null"` // new_chapter, reply, system
	Title     string    `json:"title" gorm:"size:255;not null"`
	Content   *string   `json:"content"`
	Link      *string   `json:"link"` // URL để navigate
	IsRead    bool      `json:"is_read" gorm:"default:false;index"`
	CreatedAt time.Time `json:"created_at"`

	// Relations
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (Notification) TableName() string {
	return "notifications"
}

func (n *Notification) BeforeCreate(tx *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	return nil
}