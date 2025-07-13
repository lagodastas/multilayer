package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type E2ETestSetup struct {
	serverProcess *exec.Cmd
	baseURL       string
	cleanupFuncs  []func()
}

func setupE2ETest(t *testing.T) *E2ETestSetup {
	// Настраиваем переменные окружения для тестов
	os.Setenv("DB_TYPE", "sqlite")
	os.Setenv("DB_PATH", ":memory:")
	os.Setenv("PORT", "0") // Случайный порт

	// Запускаем сервер
	cmd := exec.Command("go", "run", "./cmd/server/main.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	require.NoError(t, err, "Failed to start server")

	// Ждем запуска сервера
	time.Sleep(3 * time.Second)

	// Получаем порт из процесса (в реальном сценарии нужно парсить вывод)
	// Для простоты используем фиксированный порт в тестах
	baseURL := "http://localhost:8080"

	// Проверяем, что сервер отвечает
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			require.Fail(t, "Server failed to start within timeout")
		default:
			resp, err := http.Get(fmt.Sprintf("%s/health", baseURL))
			if err == nil && resp.StatusCode == http.StatusOK {
				resp.Body.Close()
				break
			}
			if resp != nil {
				resp.Body.Close()
			}
			time.Sleep(1 * time.Second)
		}
		break
	}

	cleanup := func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}

	return &E2ETestSetup{
		serverProcess: cmd,
		baseURL:       baseURL,
		cleanupFuncs:  []func(){cleanup},
	}
}

func (e *E2ETestSetup) cleanup() {
	for _, cleanup := range e.cleanupFuncs {
		cleanup()
	}
}

func TestE2EUserLifecycle(t *testing.T) {
	setup := setupE2ETest(t)
	defer setup.cleanup()

	t.Run("Complete User Lifecycle", func(t *testing.T) {
		// 1. Регистрация пользователя
		userData := map[string]string{
			"username": "e2euser",
			"email":    "e2e@example.com",
		}
		jsonData, _ := json.Marshal(userData)

		resp, err := http.Post(
			fmt.Sprintf("%s/users", setup.baseURL),
			"application/json",
			bytes.NewBuffer(jsonData),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var createdUser map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&createdUser)
		require.NoError(t, err)

		userID := createdUser["id"]
		assert.NotNil(t, userID)
		assert.Equal(t, "e2euser", createdUser["username"])
		assert.Equal(t, "e2e@example.com", createdUser["email"])

		// 2. Получение пользователя
		resp, err = http.Get(fmt.Sprintf("%s/users/%v", setup.baseURL, userID))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var retrievedUser map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&retrievedUser)
		require.NoError(t, err)

		assert.Equal(t, userID, retrievedUser["id"])
		assert.Equal(t, "e2euser", retrievedUser["username"])
		assert.Equal(t, "e2e@example.com", retrievedUser["email"])

		// 3. Обновление пользователя
		updateData := map[string]string{
			"username": "updatede2euser",
			"email":    "updatede2e@example.com",
		}
		jsonData, _ = json.Marshal(updateData)

		req, err := http.NewRequest(
			"PUT",
			fmt.Sprintf("%s/users/%v", setup.baseURL, userID),
			bytes.NewBuffer(jsonData),
		)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err = http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var updatedUser map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&updatedUser)
		require.NoError(t, err)

		assert.Equal(t, userID, updatedUser["id"])
		assert.Equal(t, "updatede2euser", updatedUser["username"])
		assert.Equal(t, "updatede2e@example.com", updatedUser["email"])

		// 4. Проверка, что изменения сохранились
		resp, err = http.Get(fmt.Sprintf("%s/users/%v", setup.baseURL, userID))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var finalUser map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&finalUser)
		require.NoError(t, err)

		assert.Equal(t, "updatede2euser", finalUser["username"])
		assert.Equal(t, "updatede2e@example.com", finalUser["email"])
	})
}

func TestE2EMultipleUsers(t *testing.T) {
	setup := setupE2ETest(t)
	defer setup.cleanup()

	t.Run("Multiple Users Registration", func(t *testing.T) {
		users := []map[string]string{
			{"username": "user1", "email": "user1@example.com"},
			{"username": "user2", "email": "user2@example.com"},
			{"username": "user3", "email": "user3@example.com"},
		}

		createdUsers := make([]map[string]interface{}, len(users))

		// Регистрируем всех пользователей
		for i, userData := range users {
			jsonData, _ := json.Marshal(userData)

			resp, err := http.Post(
				fmt.Sprintf("%s/users", setup.baseURL),
				"application/json",
				bytes.NewBuffer(jsonData),
			)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusCreated, resp.StatusCode)

			var createdUser map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&createdUser)
			require.NoError(t, err)

			createdUsers[i] = createdUser
		}

		// Проверяем, что все пользователи созданы с разными ID
		userIDs := make(map[interface{}]bool)
		for _, user := range createdUsers {
			userID := user["id"]
			assert.False(t, userIDs[userID], "User IDs should be unique")
			userIDs[userID] = true
		}

		// Проверяем, что можем получить каждого пользователя
		for _, user := range createdUsers {
			resp, err := http.Get(fmt.Sprintf("%s/users/%v", setup.baseURL, user["id"]))
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var retrievedUser map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&retrievedUser)
			require.NoError(t, err)

			assert.Equal(t, user["id"], retrievedUser["id"])
			assert.Equal(t, user["username"], retrievedUser["username"])
			assert.Equal(t, user["email"], retrievedUser["email"])
		}
	})
}

func TestE2EErrorHandling(t *testing.T) {
	setup := setupE2ETest(t)
	defer setup.cleanup()

	t.Run("Invalid User Registration", func(t *testing.T) {
		// Тест с невалидными данными
		testCases := []struct {
			name     string
			userData map[string]string
			expected int
		}{
			{
				name: "Empty Username",
				userData: map[string]string{
					"username": "",
					"email":    "test@example.com",
				},
				expected: http.StatusInternalServerError,
			},
			{
				name: "Short Username",
				userData: map[string]string{
					"username": "ab",
					"email":    "test@example.com",
				},
				expected: http.StatusInternalServerError,
			},
			{
				name: "Invalid Email",
				userData: map[string]string{
					"username": "testuser",
					"email":    "invalid-email",
				},
				expected: http.StatusInternalServerError,
			},
			{
				name: "Empty Email",
				userData: map[string]string{
					"username": "testuser",
					"email":    "",
				},
				expected: http.StatusInternalServerError,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				jsonData, _ := json.Marshal(tc.userData)

				resp, err := http.Post(
					fmt.Sprintf("%s/users", setup.baseURL),
					"application/json",
					bytes.NewBuffer(jsonData),
				)
				require.NoError(t, err)
				defer resp.Body.Close()

				assert.Equal(t, tc.expected, resp.StatusCode)
			})
		}
	})

	t.Run("Duplicate User Registration", func(t *testing.T) {
		// Создаем первого пользователя
		userData := map[string]string{
			"username": "duplicateuser",
			"email":    "duplicate@example.com",
		}
		jsonData, _ := json.Marshal(userData)

		resp, err := http.Post(
			fmt.Sprintf("%s/users", setup.baseURL),
			"application/json",
			bytes.NewBuffer(jsonData),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		// Пытаемся создать второго с тем же email
		resp, err = http.Post(
			fmt.Sprintf("%s/users", setup.baseURL),
			"application/json",
			bytes.NewBuffer(jsonData),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	})

	t.Run("Non-existent User Retrieval", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("%s/users/99999", setup.baseURL))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Invalid User ID", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("%s/users/invalid", setup.baseURL))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestE2EConcurrentRequests(t *testing.T) {
	setup := setupE2ETest(t)
	defer setup.cleanup()

	t.Run("Concurrent User Registration", func(t *testing.T) {
		const numGoroutines = 10
		done := make(chan bool, numGoroutines)
		errors := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer func() { done <- true }()

				userData := map[string]string{
					"username": fmt.Sprintf("concurrentuser%d", id),
					"email":    fmt.Sprintf("concurrent%d@example.com", id),
				}
				jsonData, _ := json.Marshal(userData)

				resp, err := http.Post(
					fmt.Sprintf("%s/users", setup.baseURL),
					"application/json",
					bytes.NewBuffer(jsonData),
				)
				if err != nil {
					errors <- err
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusCreated {
					errors <- fmt.Errorf("expected status 201, got %d", resp.StatusCode)
					return
				}
			}(i)
		}

		// Ждем завершения всех горутин
		for i := 0; i < numGoroutines; i++ {
			<-done
		}

		// Проверяем ошибки
		close(errors)
		for err := range errors {
			require.NoError(t, err)
		}
	})
}

func TestE2EPerformance(t *testing.T) {
	setup := setupE2ETest(t)
	defer setup.cleanup()

	t.Run("Bulk User Creation Performance", func(t *testing.T) {
		const numUsers = 50
		startTime := time.Now()

		for i := 0; i < numUsers; i++ {
			userData := map[string]string{
				"username": fmt.Sprintf("perfuser%d", i),
				"email":    fmt.Sprintf("perf%d@example.com", i),
			}
			jsonData, _ := json.Marshal(userData)

			resp, err := http.Post(
				fmt.Sprintf("%s/users", setup.baseURL),
				"application/json",
				bytes.NewBuffer(jsonData),
			)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusCreated, resp.StatusCode)
		}

		duration := time.Since(startTime)
		avgTime := duration / numUsers

		t.Logf("Created %d users in %v (avg: %v per user)", numUsers, duration, avgTime)

		// Проверяем, что среднее время создания пользователя разумное
		assert.Less(t, avgTime, 100*time.Millisecond, "Average user creation time should be less than 100ms")
	})
}

func TestE2EHealthCheck(t *testing.T) {
	setup := setupE2ETest(t)
	defer setup.cleanup()

	t.Run("Health Check Endpoint", func(t *testing.T) {
		resp, err := http.Get(fmt.Sprintf("%s/health", setup.baseURL))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var healthResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&healthResponse)
		require.NoError(t, err)

		assert.Equal(t, "healthy", healthResponse["status"])
		assert.Equal(t, "Service is running", healthResponse["message"])
	})

	t.Run("Health Check Under Load", func(t *testing.T) {
		const numRequests = 100
		startTime := time.Now()

		for i := 0; i < numRequests; i++ {
			resp, err := http.Get(fmt.Sprintf("%s/health", setup.baseURL))
			require.NoError(t, err)
			resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)
		}

		duration := time.Since(startTime)
		avgTime := duration / numRequests

		t.Logf("Made %d health check requests in %v (avg: %v per request)", numRequests, duration, avgTime)

		// Проверяем, что health check работает быстро
		assert.Less(t, avgTime, 10*time.Millisecond, "Average health check time should be less than 10ms")
	})
}

func TestE2EDataPersistence(t *testing.T) {
	setup := setupE2ETest(t)
	defer setup.cleanup()

	t.Run("Data Persistence Across Requests", func(t *testing.T) {
		// Создаем пользователя
		userData := map[string]string{
			"username": "persistenceuser",
			"email":    "persistence@example.com",
		}
		jsonData, _ := json.Marshal(userData)

		resp, err := http.Post(
			fmt.Sprintf("%s/users", setup.baseURL),
			"application/json",
			bytes.NewBuffer(jsonData),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		var createdUser map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&createdUser)
		require.NoError(t, err)

		userID := createdUser["id"]

		// Делаем несколько запросов на получение пользователя
		for i := 0; i < 5; i++ {
			resp, err := http.Get(fmt.Sprintf("%s/users/%v", setup.baseURL, userID))
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var retrievedUser map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&retrievedUser)
			require.NoError(t, err)

			assert.Equal(t, userID, retrievedUser["id"])
			assert.Equal(t, "persistenceuser", retrievedUser["username"])
			assert.Equal(t, "persistence@example.com", retrievedUser["email"])
		}
	})
}

func TestE2EInvalidRequests(t *testing.T) {
	setup := setupE2ETest(t)
	defer setup.cleanup()

	t.Run("Invalid HTTP Methods", func(t *testing.T) {
		// Тестируем несуществующие endpoints
		resp, err := http.Get(fmt.Sprintf("%s/nonexistent", setup.baseURL))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		// Тестируем неподдерживаемые методы
		req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/users/1", setup.baseURL), nil)
		require.NoError(t, err)

		resp, err = http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Malformed JSON", func(t *testing.T) {
		// Отправляем невалидный JSON
		resp, err := http.Post(
			fmt.Sprintf("%s/users", setup.baseURL),
			"application/json",
			strings.NewReader(`{"username": "test", "email": "test@example.com"`), // Неполный JSON
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Wrong Content-Type", func(t *testing.T) {
		// Отправляем данные с неправильным Content-Type
		userData := map[string]string{
			"username": "testuser",
			"email":    "test@example.com",
		}
		jsonData, _ := json.Marshal(userData)

		req, err := http.NewRequest("POST", fmt.Sprintf("%s/users", setup.baseURL), bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "text/plain") // Неправильный Content-Type

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Fiber может обработать JSON даже с неправильным Content-Type
		// поэтому проверяем, что запрос не завершился с ошибкой 500
		assert.NotEqual(t, http.StatusInternalServerError, resp.StatusCode)
	})
}
