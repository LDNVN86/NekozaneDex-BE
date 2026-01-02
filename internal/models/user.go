package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)


type User struct {
	ID				uuid.UUID			`json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email       	string				`json:"email" gorm:"uniqueIndex;not null;size:255"`
	Username    	string				`json:"username" gorm:"uniqueIndex;not null;size:50"`
	PasswordHash 	string				`json:"-" gorm:"not null"`
	AvatarURL 		*string 			`json:"avatar_url"`
	Role			string				`json:"role" gorm:"default:reader;size:20;not null"`
	IsActive		bool 				`json:"is_active"`
	CreatedAt		time.Time			`json:"created_at"`
	UpdatedAt		time.Time			`json:"updated_at"`
	DeletedAt		gorm.DeletedAt		`json:"deleted_at"`

	//Relations
	Bookmarks		[]BookMark	  		`json:"bookmarks,omitempty" gorm:"foreignKey:UserID"`
	ReadingHistory  []ReadingHistory 	`json:"reading_history,omitempty" gorm:"foreignKey:UserID"`
	Comments        []Comment		  	`json:"comments,omitempty" gorm:"foreignKey:UserID"`
	Settings		*UserSettings 	  	`json:"settings,omitempty" gorm:"foreignKey:UserID"`
	

}

//Table name - custom tabel name
func (User) TableName() string{
		return "users"
	}

//Before create - hook trc khi create
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}