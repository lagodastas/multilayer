package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type DockerTestSetup struct {
	client    *client.Client
	container string
	port      string
}

func setupDockerTest(t *testing.T) *DockerTestSetup {
	// Создаем Docker клиент
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	require.NoError(t, err)
	defer cli.Close()

	// Проверяем подключение к Docker
	_, err = cli.Ping(context.Background())
	require.NoError(t, err, "Docker daemon should be running")

	// Создаем уникальное имя контейнера
	containerName := fmt.Sprintf("multilayer-test-%d", time.Now().Unix())

	exposedPorts := nat.PortSet{
		"8080/tcp": struct{}{},
	}
	portBindings := nat.PortMap{
		"8080/tcp": []nat.PortBinding{
			{HostIP: "127.0.0.1", HostPort: "0"},
		},
	}

	// Создаем контейнер
	resp, err := cli.ContainerCreate(
		context.Background(),
		&container.Config{
			Image:        "multilayer-app:latest", // Предполагаем, что образ уже собран
			ExposedPorts: exposedPorts,
			Env: []string{
				"DB_TYPE=sqlite",
				"DB_PATH=/app/test.db",
				"PORT=8080",
			},
		},
		&container.HostConfig{
			PortBindings: portBindings,
		},
		nil, nil, containerName)
	require.NoError(t, err)

	// Запускаем контейнер
	err = cli.ContainerStart(context.Background(), resp.ID, types.ContainerStartOptions{})
	require.NoError(t, err)

	// Ждем запуска контейнера
	time.Sleep(5 * time.Second)

	// Получаем информацию о портах
	inspect, err := cli.ContainerInspect(context.Background(), resp.ID)
	require.NoError(t, err)

	// Получаем привязанный порт
	var hostPort string
	for _, binding := range inspect.NetworkSettings.Ports["8080/tcp"] {
		if binding.HostIP == "localhost" {
			hostPort = binding.HostPort
			break
		}
	}
	require.NotEmpty(t, hostPort, "Container port should be bound")

	return &DockerTestSetup{
		client:    cli,
		container: resp.ID,
		port:      hostPort,
	}
}

func (d *DockerTestSetup) cleanup() {
	if d.client != nil && d.container != "" {
		// Останавливаем контейнер
		d.client.ContainerStop(context.Background(), d.container, container.StopOptions{})
		// Удаляем контейнер
		d.client.ContainerRemove(context.Background(), d.container, types.ContainerRemoveOptions{})
	}
}

func TestDockerContainerStartup(t *testing.T) {
	setup := setupDockerTest(t)
	defer setup.cleanup()

	// Проверяем, что контейнер запущен
	inspect, err := setup.client.ContainerInspect(context.Background(), setup.container)
	require.NoError(t, err)
	assert.True(t, inspect.State.Running, "Container should be running")

	// Проверяем health check endpoint
	resp, err := http.Get(fmt.Sprintf("http://localhost:%s/health", setup.port))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var healthResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&healthResponse)
	require.NoError(t, err)

	assert.Equal(t, "healthy", healthResponse["status"])
}

func TestDockerAPIIntegration(t *testing.T) {
	setup := setupDockerTest(t)
	defer setup.cleanup()

	baseURL := fmt.Sprintf("http://localhost:%s", setup.port)

	t.Run("User Registration in Container", func(t *testing.T) {
		// Регистрируем пользователя
		userData := map[string]string{
			"username": "dockeruser",
			"email":    "docker@example.com",
		}
		jsonData, _ := json.Marshal(userData)

		resp, err := http.Post(
			fmt.Sprintf("%s/users", baseURL),
			"application/json",
			strings.NewReader(string(jsonData)),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var user map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&user)
		require.NoError(t, err)

		assert.NotNil(t, user["id"])
		assert.Equal(t, "dockeruser", user["username"])
		assert.Equal(t, "docker@example.com", user["email"])
	})

	t.Run("User Retrieval in Container", func(t *testing.T) {
		// Сначала создаем пользователя
		userData := map[string]string{
			"username": "retrievaluser",
			"email":    "retrieval@example.com",
		}
		jsonData, _ := json.Marshal(userData)

		resp, err := http.Post(
			fmt.Sprintf("%s/users", baseURL),
			"application/json",
			strings.NewReader(string(jsonData)),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		var createdUser map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&createdUser)
		require.NoError(t, err)

		// Получаем пользователя
		resp, err = http.Get(fmt.Sprintf("%s/users/%v", baseURL, createdUser["id"]))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var retrievedUser map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&retrievedUser)
		require.NoError(t, err)

		assert.Equal(t, createdUser["id"], retrievedUser["id"])
		assert.Equal(t, "retrievaluser", retrievedUser["username"])
	})

	t.Run("User Update in Container", func(t *testing.T) {
		// Сначала создаем пользователя
		userData := map[string]string{
			"username": "updateuser",
			"email":    "update@example.com",
		}
		jsonData, _ := json.Marshal(userData)

		resp, err := http.Post(
			fmt.Sprintf("%s/users", baseURL),
			"application/json",
			strings.NewReader(string(jsonData)),
		)
		require.NoError(t, err)
		defer resp.Body.Close()

		var createdUser map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&createdUser)
		require.NoError(t, err)

		// Обновляем пользователя
		updateData := map[string]string{
			"username": "updateduser",
			"email":    "updated@example.com",
		}
		jsonData, _ = json.Marshal(updateData)

		req, err := http.NewRequest(
			"PUT",
			fmt.Sprintf("%s/users/%v", baseURL, createdUser["id"]),
			strings.NewReader(string(jsonData)),
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

		assert.Equal(t, "updateduser", updatedUser["username"])
		assert.Equal(t, "updated@example.com", updatedUser["email"])
	})
}

func TestDockerContainerLogs(t *testing.T) {
	setup := setupDockerTest(t)
	defer setup.cleanup()

	// Получаем логи контейнера
	logs, err := setup.client.ContainerLogs(context.Background(), setup.container, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	})
	require.NoError(t, err)
	defer logs.Close()

	// Читаем логи
	logContent, err := io.ReadAll(logs)
	require.NoError(t, err)

	logString := string(logContent)

	// Проверяем, что в логах есть информация о запуске
	assert.Contains(t, logString, "server", "Logs should contain server startup information")
}

func TestDockerContainerRestart(t *testing.T) {
	setup := setupDockerTest(t)
	defer setup.cleanup()

	baseURL := fmt.Sprintf("http://localhost:%s", setup.port)

	// Создаем пользователя
	userData := map[string]string{
		"username": "restartuser",
		"email":    "restart@example.com",
	}
	jsonData, _ := json.Marshal(userData)

	resp, err := http.Post(
		fmt.Sprintf("%s/users", baseURL),
		"application/json",
		strings.NewReader(string(jsonData)),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	var createdUser map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&createdUser)
	require.NoError(t, err)

	// Перезапускаем контейнер
	err = setup.client.ContainerRestart(context.Background(), setup.container, container.StopOptions{})
	require.NoError(t, err)

	// Ждем перезапуска
	time.Sleep(10 * time.Second)

	// Проверяем, что контейнер снова работает
	resp, err = http.Get(fmt.Sprintf("%s/health", baseURL))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Проверяем, что данные сохранились (если используется persistent storage)
	resp, err = http.Get(fmt.Sprintf("%s/users/%v", baseURL, createdUser["id"]))
	if err == nil && resp.StatusCode == http.StatusOK {
		// Данные сохранились (возможно, используется volume)
		var retrievedUser map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&retrievedUser)
		require.NoError(t, err)
		assert.Equal(t, createdUser["id"], retrievedUser["id"])
	} else {
		// Данные не сохранились (in-memory база)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	}
}

func TestDockerContainerResourceLimits(t *testing.T) {
	setup := setupDockerTest(t)
	defer setup.cleanup()

	// Получаем статистику контейнера
	stats, err := setup.client.ContainerStats(context.Background(), setup.container, false)
	require.NoError(t, err)
	defer stats.Body.Close()

	// Читаем статистику
	statsData, err := io.ReadAll(stats.Body)
	require.NoError(t, err)

	var containerStats types.Stats
	err = json.Unmarshal(statsData, &containerStats)
	require.NoError(t, err)

	// Проверяем, что статистика доступна
	assert.NotNil(t, containerStats.CPUStats, "CPU stats should be available")
	assert.NotNil(t, containerStats.MemoryStats, "Memory stats should be available")
}

func TestDockerNetworkIntegration(t *testing.T) {
	setup := setupDockerTest(t)
	defer setup.cleanup()

	// Получаем информацию о сети контейнера
	inspect, err := setup.client.ContainerInspect(context.Background(), setup.container)
	require.NoError(t, err)

	// Проверяем, что контейнер подключен к сети
	assert.NotEmpty(t, inspect.NetworkSettings.Networks, "Container should be connected to a network")

	// Проверяем, что порт привязан
	assert.NotEmpty(t, inspect.NetworkSettings.Ports["8080/tcp"], "Port 8080 should be bound")
}

func TestDockerEnvironmentVariables(t *testing.T) {
	setup := setupDockerTest(t)
	defer setup.cleanup()

	// Получаем информацию о контейнере
	inspect, err := setup.client.ContainerInspect(context.Background(), setup.container)
	require.NoError(t, err)

	// Проверяем переменные окружения
	envVars := make(map[string]string)
	for _, env := range inspect.Config.Env {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			envVars[parts[0]] = parts[1]
		}
	}

	// Проверяем необходимые переменные
	assert.Equal(t, "sqlite", envVars["DB_TYPE"], "DB_TYPE should be set to sqlite")
	assert.Equal(t, "/app/test.db", envVars["DB_PATH"], "DB_PATH should be set")
	assert.Equal(t, "8080", envVars["PORT"], "PORT should be set to 8080")
}

func TestDockerContainerHealthCheck(t *testing.T) {
	setup := setupDockerTest(t)
	defer setup.cleanup()

	// Ждем, чтобы контейнер полностью запустился
	time.Sleep(10 * time.Second)

	// Получаем информацию о состоянии контейнера
	inspect, err := setup.client.ContainerInspect(context.Background(), setup.container)
	require.NoError(t, err)

	// Проверяем health check
	if inspect.State.Health != nil {
		assert.Equal(t, "healthy", inspect.State.Health.Status, "Container should be healthy")
	} else {
		// Если health check не настроен, проверяем что контейнер работает
		assert.True(t, inspect.State.Running, "Container should be running")
	}
}
