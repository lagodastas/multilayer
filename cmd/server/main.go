package main

import (
	"fmt"
	"multilayer/internal/controller"
	"multilayer/internal/entity"
	"multilayer/internal/repository"
	"multilayer/internal/service"
	"os"

	"github.com/gofiber/fiber/v2"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func getDatabaseConnection() (*gorm.DB, error) {
	dbType := os.Getenv("DB_TYPE")
	if dbType == "" {
		dbType = "sqlite" // default to sqlite
	}

	switch dbType {
	case "postgres":
		host := os.Getenv("DB_HOST")
		if host == "" {
			host = "localhost"
		}
		port := os.Getenv("DB_PORT")
		if port == "" {
			port = "5432"
		}
		user := os.Getenv("DB_USER")
		if user == "" {
			user = "postgres"
		}
		password := os.Getenv("DB_PASSWORD")
		if password == "" {
			password = "password"
		}
		dbname := os.Getenv("DB_NAME")
		if dbname == "" {
			dbname = "multilayer"
		}

		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
			host, user, password, dbname, port)
		return gorm.Open(postgres.Open(dsn), &gorm.Config{})

	case "sqlite":
		fallthrough
	default:
		dbPath := os.Getenv("DB_PATH")
		if dbPath == "" {
			dbPath = "test.db"
		}
		return gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	}
}

func main() {
	// Инициализация БД
	db, err := getDatabaseConnection()
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}

	err = db.AutoMigrate(&entity.User{})
	if err != nil {
		panic("failed to migrate database: " + err.Error())
	}

	// Инициализация слоёв
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userController := controller.NewUserController(userService)

	// Создаем Fiber приложение
	app := fiber.New()

	// Health check endpoint для Kubernetes
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"message": "Service is running",
		})
	})

	// Настраиваем роуты
	app.Post("/users", userController.Register)
	app.Get("/users/:id", userController.GetUser)
	app.Put("/users/:id", userController.UpdateUser)

	// Получаем порт из переменной окружения
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Запускаем сервер
	err = app.Listen(":" + port)
	if err != nil {
		panic("failed to start server: " + err.Error())
	}
}
