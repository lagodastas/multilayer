package repository_test

import (
	"multilayer/internal/entity"
	"multilayer/internal/repository"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	err = db.AutoMigrate(&entity.User{})
	if err != nil {
		panic("failed to migrate database")
	}
	return db
}

func TestUserRepository_Create(t *testing.T) {
	db := setupTestDB()
	repo := repository.NewUserRepository(db)

	user := &entity.User{
		Username: "testuser",
		Email:    "test@example.com",
	}

	err := repo.Create(user)

	assert.NoError(t, err)
	assert.NotZero(t, user.ID)
}

func TestUserRepository_Update(t *testing.T) {
	db := setupTestDB()
	repo := repository.NewUserRepository(db)

	// Создаём пользователя для теста
	user := &entity.User{Username: "old", Email: "old@example.com"}
	err := repo.Create(user)
	assert.NoError(t, err)

	// Обновляем
	user.Username = "new"
	err = repo.Update(user)

	assert.NoError(t, err)

	// Проверяем в БД
	var updatedUser entity.User
	db.First(&updatedUser, user.ID)
	assert.Equal(t, "new", updatedUser.Username)
}
