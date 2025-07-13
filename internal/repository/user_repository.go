// (Слой работы с БД)
package repository

import (
	"gorm.io/gorm"
	"multilayer/internal/entity"
)

// UserRepositoryInterface определяет контракт для репозитория
type UserRepositoryInterface interface {
	Create(user *entity.User) error
	FindByID(id uint) (*entity.User, error)
	Update(user *entity.User) error
}

type UserRepository struct {
	DB *gorm.DB
}

// NewUserRepository - конструктор для UserRepository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func (r *UserRepository) Update(user *entity.User) error {
	return r.DB.Save(user).Error
}

func (r *UserRepository) Create(user *entity.User) error {
	return r.DB.Create(user).Error
}

func (r *UserRepository) FindByID(id uint) (*entity.User, error) {
	var user entity.User
	err := r.DB.First(&user, id).Error
	return &user, err
}
