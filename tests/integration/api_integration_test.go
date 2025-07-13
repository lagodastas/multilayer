package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"multilayer/internal/controller"
	"multilayer/internal/entity"
	"multilayer/internal/repository"
	"multilayer/internal/service"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type TestSetup struct {
	app    *fiber.App
	db     *gorm.DB
	server *httptest.Server
}

func setupTestApp(t *testing.T) *TestSetup {
	// Создаем in-memory SQLite базу для тестов
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Мигрируем схему
	err = db.AutoMigrate(&entity.User{})
	require.NoError(t, err)

	// Инициализируем слои
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userController := controller.NewUserController(userService)

	// Создаем Fiber приложение
	app := fiber.New()

	// Настраиваем роуты
	app.Post("/users", userController.Register)
	app.Get("/users/:id", userController.GetUser)
	app.Put("/users/:id", userController.UpdateUser)

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"message": "Service is running",
		})
	})

	return &TestSetup{
		app: app,
		db:  db,
	}
}

func TestHealthCheck(t *testing.T) {
	setup := setupTestApp(t)
	defer setup.db.Migrator().DropTable(&entity.User{})

	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := setup.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "Service is running", response["message"])
}

func TestUserRegistrationFlow(t *testing.T) {
	setup := setupTestApp(t)
	defer setup.db.Migrator().DropTable(&entity.User{})

	// Тест 1: Успешная регистрация пользователя
	t.Run("Successful Registration", func(t *testing.T) {
		userData := map[string]string{
			"username": "testuser",
			"email":    "test@example.com",
		}
		jsonData, _ := json.Marshal(userData)

		req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		resp, err := setup.app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var user entity.User
		err = json.NewDecoder(resp.Body).Decode(&user)
		require.NoError(t, err)

		assert.NotZero(t, user.ID)
		assert.Equal(t, "testuser", user.Username)
		assert.Equal(t, "test@example.com", user.Email)
	})

	// Тест 2: Попытка регистрации с дублирующимся email
	t.Run("Duplicate Email Registration", func(t *testing.T) {
		userData := map[string]string{
			"username": "anotheruser",
			"email":    "test@example.com", // Тот же email
		}
		jsonData, _ := json.Marshal(userData)

		req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		resp, err := setup.app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	// Тест 3: Регистрация с невалидными данными
	t.Run("Invalid Data Registration", func(t *testing.T) {
		userData := map[string]string{
			"username": "ab", // Слишком короткий username
			"email":    "invalid-email",
		}
		jsonData, _ := json.Marshal(userData)

		req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		resp, err := setup.app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestUserRetrievalFlow(t *testing.T) {
	setup := setupTestApp(t)
	defer setup.db.Migrator().DropTable(&entity.User{})

	// Создаем пользователя для тестов
	userData := map[string]string{
		"username": "retrievaluser",
		"email":    "retrieval@example.com",
	}
	jsonData, _ := json.Marshal(userData)

	req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	resp, err := setup.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var createdUser entity.User
	err = json.NewDecoder(resp.Body).Decode(&createdUser)
	require.NoError(t, err)

	// Тест 1: Успешное получение пользователя
	t.Run("Successful User Retrieval", func(t *testing.T) {
		req := httptest.NewRequest("GET", fmt.Sprintf("/users/%d", createdUser.ID), nil)
		resp, err := setup.app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var user entity.User
		err = json.NewDecoder(resp.Body).Decode(&user)
		require.NoError(t, err)

		assert.Equal(t, createdUser.ID, user.ID)
		assert.Equal(t, "retrievaluser", user.Username)
		assert.Equal(t, "retrieval@example.com", user.Email)
	})

	// Тест 2: Получение несуществующего пользователя
	t.Run("Non-existent User Retrieval", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/users/99999", nil)
		resp, err := setup.app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	// Тест 3: Получение пользователя с невалидным ID
	t.Run("Invalid ID Retrieval", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/users/invalid", nil)
		resp, err := setup.app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestUserUpdateFlow(t *testing.T) {
	setup := setupTestApp(t)
	defer setup.db.Migrator().DropTable(&entity.User{})

	// Создаем пользователя для тестов
	userData := map[string]string{
		"username": "updateuser",
		"email":    "update@example.com",
	}
	jsonData, _ := json.Marshal(userData)

	req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	resp, err := setup.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var createdUser entity.User
	err = json.NewDecoder(resp.Body).Decode(&createdUser)
	require.NoError(t, err)

	// Тест 1: Успешное обновление пользователя
	t.Run("Successful User Update", func(t *testing.T) {
		updateData := map[string]string{
			"username": "updateduser",
			"email":    "updated@example.com",
		}
		jsonData, _ := json.Marshal(updateData)

		req := httptest.NewRequest("PUT", fmt.Sprintf("/users/%d", createdUser.ID), bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		resp, err := setup.app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var user entity.User
		err = json.NewDecoder(resp.Body).Decode(&user)
		require.NoError(t, err)

		assert.Equal(t, createdUser.ID, user.ID)
		assert.Equal(t, "updateduser", user.Username)
		assert.Equal(t, "updated@example.com", user.Email)
	})

	// Тест 2: Обновление несуществующего пользователя
	t.Run("Non-existent User Update", func(t *testing.T) {
		updateData := map[string]string{
			"username": "nonexistent",
			"email":    "nonexistent@example.com",
		}
		jsonData, _ := json.Marshal(updateData)

		req := httptest.NewRequest("PUT", "/users/99999", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		resp, err := setup.app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	// Тест 3: Обновление с невалидными данными
	t.Run("Invalid Data Update", func(t *testing.T) {
		updateData := map[string]string{
			"username": "ab", // Слишком короткий username
			"email":    "invalid-email",
		}
		jsonData, _ := json.Marshal(updateData)

		req := httptest.NewRequest("PUT", fmt.Sprintf("/users/%d", createdUser.ID), bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		resp, err := setup.app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

func TestCompleteUserWorkflow(t *testing.T) {
	setup := setupTestApp(t)
	defer setup.db.Migrator().DropTable(&entity.User{})

	// Полный цикл: создание -> получение -> обновление -> получение
	t.Run("Complete User Lifecycle", func(t *testing.T) {
		// 1. Создание пользователя
		userData := map[string]string{
			"username": "workflowuser",
			"email":    "workflow@example.com",
		}
		jsonData, _ := json.Marshal(userData)

		req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		resp, err := setup.app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var createdUser entity.User
		err = json.NewDecoder(resp.Body).Decode(&createdUser)
		require.NoError(t, err)

		// 2. Получение созданного пользователя
		req = httptest.NewRequest("GET", fmt.Sprintf("/users/%d", createdUser.ID), nil)
		resp, err = setup.app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var retrievedUser entity.User
		err = json.NewDecoder(resp.Body).Decode(&retrievedUser)
		require.NoError(t, err)
		assert.Equal(t, createdUser.ID, retrievedUser.ID)

		// 3. Обновление пользователя
		updateData := map[string]string{
			"username": "updatedworkflowuser",
			"email":    "updatedworkflow@example.com",
		}
		jsonData, _ = json.Marshal(updateData)

		req = httptest.NewRequest("PUT", fmt.Sprintf("/users/%d", createdUser.ID), bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		resp, err = setup.app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var updatedUser entity.User
		err = json.NewDecoder(resp.Body).Decode(&updatedUser)
		require.NoError(t, err)
		assert.Equal(t, "updatedworkflowuser", updatedUser.Username)
		assert.Equal(t, "updatedworkflow@example.com", updatedUser.Email)

		// 4. Проверка, что изменения сохранились
		req = httptest.NewRequest("GET", fmt.Sprintf("/users/%d", createdUser.ID), nil)
		resp, err = setup.app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var finalUser entity.User
		err = json.NewDecoder(resp.Body).Decode(&finalUser)
		require.NoError(t, err)
		assert.Equal(t, "updatedworkflowuser", finalUser.Username)
		assert.Equal(t, "updatedworkflow@example.com", finalUser.Email)
	})
}
