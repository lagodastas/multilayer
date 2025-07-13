package controller_test

import (
	"bytes"
	"encoding/json"
	"multilayer/internal/controller"
	"multilayer/internal/entity"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) UpdateUser(id uint, username, email string) (*entity.User, error) {
	args := m.Called(id, username, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserService) RegisterUser(username, email string) (*entity.User, error) {
	args := m.Called(username, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func (m *MockUserService) GetUser(id uint) (*entity.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.User), args.Error(1)
}

func TestUserController_UpdateUser(t *testing.T) {
	// Создаем Fiber app для тестов
	app := fiber.New()

	// Мок сервиса
	mockService := new(MockUserService)
	userController := controller.NewUserController(mockService)

	// Тестовый маршрут
	app.Put("/users/:id", userController.UpdateUser)

	// Тестовый случай
	t.Run("Success", func(t *testing.T) {
		// Ожидаемый пользователь
		expectedUser := &entity.User{
			ID:       1,
			Username: "new",
			Email:    "new@example.com",
		}

		// Настраиваем мок
		mockService.On("UpdateUser", uint(1), "new", "new@example.com").
			Return(expectedUser, nil)

		// Создаем тестовый запрос
		requestBody := map[string]string{
			"username": "new",
			"email":    "new@example.com",
		}
		jsonBody, _ := json.Marshal(requestBody)

		req := httptest.NewRequest("PUT", "/users/1", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)

		// Проверки
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
		mockService.AssertExpectations(t)
	})
}
