package repositories

import (
	"nekozanedex/internal/models"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	CreateUser(user *models.User) error
	FindUserByID(id uuid.UUID) (*models.User, error)
	FindUserByEmail(email string) (*models.User, error)
	FindUserByUsername(username string) (*models.User, error)
	FindUserByTagName(tagName string) (*models.User, error)
	FindUsersByUsernames(usernames []string) ([]models.User, error)
	FindUsersByTagNames(tagNames []string) ([]models.User, error)
	SearchUsersByUsername(query string, limit int) ([]models.User, error)
	UpdateUser(user *models.User) error
	DeleteUser(id uuid.UUID) error
	GetAllUsers(page, limit int) ([]models.User, int64, error)
	SearchUsersAdmin(query string, page, limit int) ([]models.User, int64, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// Create User - Tạo User
func (r *userRepository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) FindUserByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, "email = ?", email).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindUserByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, "username = ?", username).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindUserByTagName - Find user by tag_name (for public profile)
func (r *userRepository) FindUserByTagName(tagName string) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, "tag_name = ? AND is_active = ?", tagName, true).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindUsersByUsernames - Find multiple users by their usernames (for @mention)
func (r *userRepository) FindUsersByUsernames(usernames []string) ([]models.User, error) {
	if len(usernames) == 0 {
		return []models.User{}, nil
	}
	var users []models.User
	err := r.db.Where("username IN ?", usernames).Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

// FindUsersByTagNames - Find multiple users by their tag_names (for @mention notifications)
func (r *userRepository) FindUsersByTagNames(tagNames []string) ([]models.User, error) {
	if len(tagNames) == 0 {
		return []models.User{}, nil
	}
	var users []models.User
	err := r.db.Where("tag_name IN ?", tagNames).Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepository) UpdateUser(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) DeleteUser(id uuid.UUID) error {
	return r.db.Delete(&models.User{}, "id = ?", id).Error
}

func (r *userRepository) GetAllUsers(page, limit int) ([]models.User, int64, error) {
	var users []models.User
	var total int64
	r.db.Model(&models.User{}).Count(&total)
	offset := (page - 1) * limit
	err := r.db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&users).Error
	if err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

// SearchUsersAdmin - Search users by username or email with pagination
func (r *userRepository) SearchUsersAdmin(query string, page, limit int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	db := r.db.Model(&models.User{})

	if query != "" {
		searchQuery := "%" + query + "%"
		db = db.Where("username ILIKE ? OR email ILIKE ?", searchQuery, searchQuery)
	}

	db.Count(&total)

	offset := (page - 1) * limit
	err := db.Offset(offset).Limit(limit).Order("created_at DESC").Find(&users).Error
	if err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

// SearchUsersByUsername - Search active users by username or tag_name (for @mention)
func (r *userRepository) SearchUsersByUsername(query string, limit int) ([]models.User, error) {
	var users []models.User
	if query == "" {
		return users, nil
	}

	// Search both username (contains) and tag_name (prefix)
	searchQuery := "%" + query + "%"
	tagQuery := strings.ToLower(query) + "%"
	
	err := r.db.Select("id", "username", "tag_name", "avatar_url").
		Where("(username ILIKE ? OR tag_name ILIKE ?) AND is_active = ?", searchQuery, tagQuery, true).
		Limit(limit).
		Order("tag_name ASC").
		Find(&users).Error
	return users, err
}
