package repositories

import (
	"nekozanedex/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	CreateUser(user *models.User) error
	FindUserByID(id uuid.UUID) (*models.User,error)
	FindUserByEmail(email string) (*models.User,error)
	FindUserByUsername(username string) (*models.User,error)
	UpdateUser(user *models.User) error
	DeleteUser(id uuid.UUID) error
	GetAllUsers(page,limit int) ([]models.User,int64 ,error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository{
	return &userRepository{db:db}
}

//Create User - Tạo User
func (r *userRepository) CreateUser(user *models.User) error{
	return r.db.Create(user).Error
}

func (r *userRepository) FindUserByID(id uuid.UUID) (*models.User, error){
	var user models.User
	err  := r.db.First(&user, "id = ?", id).Error
	if err != nil {
		return nil,err
	}
	return &user,nil
}

func (r *userRepository) FindUserByEmail(email string) (*models.User, error){
	var user models.User
	err  := r.db.First(&user, "email = ?", email).Error
	if err != nil {
		return nil,err
	}
	return &user,nil
}

func (r *userRepository) FindUserByUsername(username string) (*models.User, error){
	var user models.User
	err  := r.db.First(&user, "username = ?", username).Error
	if err != nil {
		return nil,err
	}
	return &user,nil
}

func (r *userRepository) UpdateUser(user *models.User) error{
	return r.db.Save(user).Error
}

func (r *userRepository) DeleteUser(id uuid.UUID) error{
	return r.db.Delete(&models.User{}, "id = ?", id).Error
}

func (r *userRepository) GetAllUsers(page,limit int) ([]models.User,int64 ,error){
	var users []models.User
	var total int64
	r.db.Model(&models.User{}).Count(&total)
	offset := (page - 1) * limit
	err := r.db.Offset(offset).Limit(limit).Find(&users).Error
	if err != nil {
		return nil,0,err
	}
	return users,total,nil
}

