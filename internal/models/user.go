package models

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"gorm.io/gorm"
)

type User struct {
	ID           uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email        string         `json:"email" gorm:"uniqueIndex;not null;size:255"`
	Username     string         `json:"username" gorm:"uniqueIndex;not null;size:50"`
	TagName      string         `json:"tag_name" gorm:"uniqueIndex;size:50"`
	PasswordHash string         `json:"-" gorm:"not null"`
	AvatarURL    *string        `json:"avatar_url"`
	Role         string         `json:"role" gorm:"default:reader;size:20;not null"`
	IsActive     bool           `json:"is_active"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at"`

	//Relations
	Bookmarks      []BookMark       `json:"bookmarks,omitempty" gorm:"foreignKey:UserID"`
	ReadingHistory []ReadingHistory `json:"reading_history,omitempty" gorm:"foreignKey:UserID"`
	Comments       []Comment        `json:"comments,omitempty" gorm:"foreignKey:UserID"`
	Settings       *UserSettings    `json:"settings,omitempty" gorm:"foreignKey:UserID"`
}

// Table name - custom table name
func (User) TableName() string {
	return "users"
}

// GenerateTagName creates a tag_name from username
// Removes diacritics, converts to lowercase, removes spaces and special chars
func GenerateTagName(username string) string {
	// Normalize and remove diacritics
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(t, username)

	// Convert to lowercase
	result = strings.ToLower(result)

	// Replace spaces with empty string and remove special characters
	// Keep only letters and numbers
	reg := regexp.MustCompile(`[^a-z0-9]`)
	result = reg.ReplaceAllString(result, "")

	return result
}

// GenerateUniqueTagName generates a unique tag_name by appending numbers if collision
func GenerateUniqueTagName(tx *gorm.DB, baseTagName string, excludeUserID uuid.UUID) string {
	tagName := baseTagName
	suffix := 1

	for {
		var count int64
		query := tx.Model(&User{}).Where("tag_name = ?", tagName)
		// Exclude current user when updating
		if excludeUserID != uuid.Nil {
			query = query.Where("id != ?", excludeUserID)
		}
		query.Count(&count)

		if count == 0 {
			return tagName
		}

		// Collision detected, append suffix
		tagName = fmt.Sprintf("%s%d", baseTagName, suffix)
		suffix++

		// Safety limit
		if suffix > 1000 {
			return fmt.Sprintf("%s%d", baseTagName, time.Now().UnixNano()%10000)
		}
	}
}

// BeforeCreate - hook before create
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	// Auto-generate unique tag_name if empty
	if u.TagName == "" && u.Username != "" {
		baseTagName := GenerateTagName(u.Username)
		u.TagName = GenerateUniqueTagName(tx, baseTagName, uuid.Nil)
	}
	return nil
}

// BeforeUpdate - hook before update to keep tag_name in sync
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	// If username changed, regenerate unique tag_name
	if tx.Statement.Changed("Username") {
		baseTagName := GenerateTagName(u.Username)
		u.TagName = GenerateUniqueTagName(tx, baseTagName, u.ID)
	}
	return nil
}