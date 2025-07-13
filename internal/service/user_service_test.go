package service

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"multilayer/internal/entity"
	"testing"
)

// MockUserRepository должен реализовывать интерфейс репозитория
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *entity.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByID(id uint) (*entity.User, error) {
	args := m.Called(id)
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *entity.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func TestUserService_UpdateUser(t *testing.T) {
	// Создаем mock репозитория
	mockRepo := new(MockUserRepository)

	// Создаем сервис с mock репозиторием
	service := &UserService{
		userRepo: mockRepo,
	}

	// Тестовые данные
	testUser := &entity.User{
		ID:       1,
		Username: "old",
		Email:    "old@example.com",
	}

	// Настраиваем ожидания
	mockRepo.On("FindByID", uint(1)).Return(testUser, nil)
	mockRepo.On("Update", mock.AnythingOfType("*entity.User")).Return(nil)

	// Вызываем метод
	updatedUser, err := service.UpdateUser(1, "new", "new@example.com")

	// Проверяем результаты
	assert.NoError(t, err)
	assert.Equal(t, "new", updatedUser.Username)
	assert.Equal(t, "new@example.com", updatedUser.Email)
	mockRepo.AssertExpectations(t)
}
