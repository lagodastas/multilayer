# Интеграционные тесты для Multilayer

Этот каталог содержит интеграционные тесты для приложения Multilayer, которые проверяют взаимодействие между различными компонентами системы.

## Структура тестов

```
tests/integration/
├── api_integration_test.go      # Тесты интеграции API
├── database_integration_test.go # Тесты интеграции с базой данных
├── docker_integration_test.go   # Тесты интеграции с Docker
├── e2e_test.go                  # End-to-end тесты
└── README.md                    # Эта документация
```

## Типы интеграционных тестов

### 1. API Integration Tests (`api_integration_test.go`)

Тестируют HTTP API endpoints и их взаимодействие с бизнес-логикой:

- **Health Check** - проверка работоспособности сервиса
- **User Registration Flow** - полный цикл регистрации пользователя
- **User Retrieval Flow** - получение пользователей
- **User Update Flow** - обновление пользователей
- **Complete User Workflow** - полный жизненный цикл пользователя

**Особенности:**
- Используют in-memory SQLite для изоляции тестов
- Тестируют все слои приложения (Controller → Service → Repository → Database)
- Проверяют валидацию данных и обработку ошибок

### 2. Database Integration Tests (`database_integration_test.go`)

Тестируют взаимодействие с базой данных:

- **Database Connection and Migration** - подключение и миграции
- **User Repository Integration** - операции с репозиторием
- **User Service Integration** - бизнес-логика через сервис
- **Database Constraints** - проверка уникальности и ограничений
- **Transaction Rollback** - откат транзакций
- **Performance Tests** - массовое создание и конкурентность

**Особенности:**
- Используют in-memory SQLite для быстрых тестов
- Проверяют целостность данных
- Тестируют конкурентный доступ

### 3. Docker Integration Tests (`docker_integration_test.go`)

Тестируют работу приложения в Docker контейнерах:

- **Container Startup** - запуск и инициализация контейнера
- **API Integration in Container** - работа API внутри контейнера
- **Container Logs** - проверка логов
- **Container Restart** - перезапуск и восстановление
- **Resource Limits** - ограничения ресурсов
- **Network Integration** - сетевое взаимодействие
- **Environment Variables** - переменные окружения
- **Health Check** - проверка состояния контейнера

**Особенности:**
- Требуют Docker daemon
- Создают и управляют контейнерами программно
- Проверяют изоляцию и безопасность

### 4. End-to-End Tests (`e2e_test.go`)

Полные end-to-end тесты, имитирующие реальное использование:

- **User Lifecycle** - полный жизненный цикл пользователя
- **Multiple Users** - работа с несколькими пользователями
- **Error Handling** - обработка ошибок
- **Concurrent Requests** - конкурентные запросы
- **Performance** - производительность
- **Health Check** - проверка работоспособности
- **Data Persistence** - сохранение данных
- **Invalid Requests** - невалидные запросы

**Особенности:**
- Запускают реальный сервер
- Имитируют реальные HTTP запросы
- Проверяют полную функциональность

## Запуск тестов

### Все интеграционные тесты
```bash
make test-integration
```

### Только end-to-end тесты
```bash
make test-e2e
```

### Все тесты (unit + integration + e2e)
```bash
make test-all
```

### С покрытием кода
```bash
make test-integration-cover
```

### Отдельные файлы тестов
```bash
# API тесты
go test -v ./tests/integration/ -run "TestAPI"

# Database тесты
go test -v ./tests/integration/ -run "TestDatabase"

# Docker тесты
go test -v ./tests/integration/ -run "TestDocker"

# E2E тесты
go test -v ./tests/integration/ -run "TestE2E"
```

## Предварительные требования

### Для всех тестов
- Go 1.24+
- testify (уже в go.mod)

### Для Docker тестов
- Docker daemon запущен
- Docker API доступен

### Для E2E тестов
- Приложение может быть скомпилировано
- Порт 8080 доступен (или настраивается автоматически)

## Конфигурация

### Переменные окружения для тестов
```bash
# База данных
DB_TYPE=sqlite
DB_PATH=:memory:

# Сервер
PORT=0  # Случайный порт для E2E тестов
```

### Docker конфигурация
```bash
# Образ для тестов
DOCKER_IMAGE=multilayer-app:latest

# Переменные окружения контейнера
DB_TYPE=sqlite
DB_PATH=/app/test.db
PORT=8080
```

## Изоляция тестов

### База данных
- Каждый тест использует in-memory SQLite
- Данные не сохраняются между тестами
- Автоматическая очистка после каждого теста

### Контейнеры
- Уникальные имена контейнеров
- Автоматическое удаление после тестов
- Изолированные сети

### Серверы
- Случайные порты для E2E тестов
- Автоматическое завершение процессов
- Изолированные переменные окружения

## Отладка тестов

### Включение подробного вывода
```bash
go test -v ./tests/integration/
```

### Запуск одного теста
```bash
go test -v ./tests/integration/ -run "TestHealthCheck"
```

### Пропуск медленных тестов
```bash
go test -v ./tests/integration/ -short
```

### Параллельное выполнение
```bash
go test -v ./tests/integration/ -parallel 4
```

## Добавление новых тестов

### Структура нового теста
```go
func TestNewFeature(t *testing.T) {
    setup := setupTestApp(t)
    defer setup.db.Migrator().DropTable(&entity.User{})

    t.Run("Test Case Name", func(t *testing.T) {
        // Arrange
        // Act
        // Assert
    })
}
```

### Лучшие практики
1. **Изоляция** - каждый тест должен быть независимым
2. **Очистка** - всегда очищайте ресурсы после тестов
3. **Именование** - используйте описательные имена тестов
4. **Документация** - комментируйте сложные тесты
5. **Производительность** - избегайте медленных операций

## Мониторинг и метрики

### Покрытие кода
```bash
make test-integration-cover
```

### Время выполнения
```bash
go test -v ./tests/integration/ -bench=.
```

### Память
```bash
go test -v ./tests/integration/ -memprofile=mem.out
```

## CI/CD интеграция

### GitHub Actions
```yaml
- name: Run Integration Tests
  run: make test-integration

- name: Run E2E Tests
  run: make test-e2e
```

### Docker в CI
```yaml
- name: Run Docker Tests
  run: |
    docker build -t multilayer-app .
    make test-integration
```

## Устранение неполадок

### Проблемы с Docker
- Убедитесь, что Docker daemon запущен
- Проверьте права доступа к Docker socket
- Очистите неиспользуемые контейнеры: `docker system prune`

### Проблемы с портами
- Проверьте, что порт 8080 свободен
- Используйте случайные порты для E2E тестов
- Добавьте таймауты для ожидания запуска сервера

### Проблемы с базой данных
- Убедитесь, что SQLite доступен
- Проверьте права на создание файлов
- Используйте in-memory базу для тестов

## Дополнительные ресурсы

- [Go Testing Package](https://golang.org/pkg/testing/)
- [Testify Documentation](https://github.com/stretchr/testify)
- [Docker Go SDK](https://pkg.go.dev/github.com/docker/docker/client)
- [Fiber Testing](https://docs.gofiber.io/guide/testing) 