package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Story struct {
	ID	        	uuid.UUID		`json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Title			string			`json:"title" gorm:"not null;size:255"`
	Slug			string			`json:"slug" gorm:"uniqueIndex;not null;size:255"`
	Description 	*string			`json:"description"`
	CoverImageURL	*string			`json:"cover_image_url"`
	AuthorName		*string			`json:"author_name" gorm:"size:100"`
	Status        	string          `json:"status" gorm:"default:ongoing;size:20"` // ongoing, completed, hiatus
	IsPublished   	bool            `json:"is_published" gorm:"default:false"`
	ViewCount     	int64           `json:"view_count" gorm:"default:0"`
	TotalChapters 	int             `json:"total_chapters" gorm:"default:0"`
	CreatedAt     	time.Time       `json:"created_at"`
	UpdatedAt     	time.Time       `json:"updated_at"`
	DeletedAt     	gorm.DeletedAt  `json:"-" gorm:"index"`

	// Relations
	Chapters []Chapter `json:"chapters,omitempty" gorm:"foreignKey:StoryID"`
	Genres   []Genre   `json:"genres,omitempty" gorm:"many2many:story_genres"`
}

//Table name - custom table name
func (Story) TableName() string {
	return "stories"
}


func (s *Story) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}