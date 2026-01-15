package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Story struct {
	ID            uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Title         string         `json:"title" gorm:"not null;size:255"`
	Slug          string         `json:"slug" gorm:"uniqueIndex;not null;size:255"`
	OriginalTitle *string        `json:"original_title" gorm:"size:255"`           // Tên gốc (JP/CN/KR)
	AltTitles     datatypes.JSON `json:"alt_titles" gorm:"type:jsonb"`             // Tên phụ ["Tên 1", "Tên 2"]
	Description   *string        `json:"description"`
	CoverImageURL *string        `json:"cover_image_url"`
	AuthorName    *string        `json:"author_name" gorm:"size:100"`
	ArtistName    *string        `json:"artist_name" gorm:"size:100"`              // Họa sĩ
	Translator    *string        `json:"translator" gorm:"size:100"`               // Dịch giả
	SourceURL     *string        `json:"source_url" gorm:"size:500"`               // Nguồn gốc (link)
	SourceName    *string        `json:"source_name" gorm:"size:100"`              // Tên nguồn gọn
	Country       *string        `json:"country" gorm:"size:20"`                   // JP, CN, KR, VN
	ReleaseYear   *int           `json:"release_year"`                             // Năm ra mắt
	EndYear       *int           `json:"end_year"`                                 // Năm kết thúc (null = đang tiếp diễn)
	Status        string         `json:"status" gorm:"default:ongoing;size:20"`    // ongoing, completed, hiatus
	IsPublished   bool           `json:"is_published" gorm:"default:false"`
	ViewCount     int64          `json:"view_count" gorm:"default:0"`
	TotalChapters int            `json:"total_chapters" gorm:"default:0"`
	Rating        *float64       `json:"rating" gorm:"type:decimal(3,2)"`          // Rating trung bình (0.00 - 5.00)
	RatingCount   int            `json:"rating_count" gorm:"default:0"`            // Số lượt đánh giá
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`

	// Relations
	Chapters []Chapter `json:"chapters,omitempty" gorm:"foreignKey:StoryID"`
	Genres   []Genre   `json:"genres,omitempty" gorm:"many2many:story_genres"`
}

// TableName - custom table name
func (Story) TableName() string {
	return "stories"
}

func (s *Story) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// GetAltTitles - Helper to get alt_titles as []string
func (s *Story) GetAltTitles() []string {
	var titles []string
	if s.AltTitles != nil {
		_ = json.Unmarshal(s.AltTitles, &titles)
	}
	return titles
}

// SetAltTitles - Helper to set alt_titles from []string
func (s *Story) SetAltTitles(titles []string) error {
	data, err := json.Marshal(titles)
	if err != nil {
		return err
	}
	s.AltTitles = data
	return nil
}