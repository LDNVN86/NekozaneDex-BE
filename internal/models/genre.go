package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Genre struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name        string    `json:"name" gorm:"uniqueIndex;not null;size:50"`
	Slug        string    `json:"slug" gorm:"uniqueIndex;not null;size:50"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	// Relations
	Stories []Story `json:"stories,omitempty" gorm:"many2many:story_genres"`
}

func (Genre) TableName() string {
	return "genres"
}

func (g *Genre) BeforeCreate(tx *gorm.DB) error {
	if g.ID == uuid.Nil {
		g.ID = uuid.New()
	}
	return nil
}
