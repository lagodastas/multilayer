// (Сервисный слой)
package service

import (
	"multilayer/internal/entity"
	"multilayer/internal/repository"
)

type UserServiceInterface interface {
	UpdateUser(id uint, username, email string) (*entity.User, error)
	RegisterUser(username, email string) (*entity.User, error)
	GetUser(id uint) (*entity.User, error)
}

type UserService struct {
	userRepo repository.UserRepositoryInterface // Используем интерфейс
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) UpdateUser(id uint, username, email string) (*entity.User, error) {
	// Сначала получаем пользователя
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Обновляем поля
	user.Username = username
	user.Email = email

	// Сохраняем изменения
	err = s.userRepo.Update(user)
	return user, err
}

func (s *UserService) RegisterUser(username, email string) (*entity.User, error) {
	user := &entity.User{
		Username: username,
		Email:    email,
	}
	err := s.userRepo.Create(user)
	return user, err
}

func (s *UserService) GetUser(id uint) (*entity.User, error) {
	return s.userRepo.FindByID(id)
}
