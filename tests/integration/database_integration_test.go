package integration

import (
	"fmt"
	"multilayer/internal/entity"
	"multilayer/internal/repository"
	"multilayer/internal/service"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DatabaseTestSetup struct {
	db          *gorm.DB
	userRepo    *repository.UserRepository
	userService *service.UserService
}

func setupDatabaseTest(t *testing.T) *DatabaseTestSetup {
	// Создаем in-memory SQLite базу для тестов
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Мигрируем схему
	err = db.AutoMigrate(&entity.User{})
	require.NoError(t, err)

	// Инициализируем слои
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)

	return &DatabaseTestSetup{
		db:          db,
		userRepo:    userRepo,
		userService: userService,
	}
}

func TestDatabaseConnectionAndMigration(t *testing.T) {
	setup := setupDatabaseTest(t)
	defer setup.db.Migrator().DropTable(&entity.User{})

	// Проверяем, что таблица создана
	hasTable := setup.db.Migrator().HasTable(&entity.User{})
	assert.True(t, hasTable, "User table should exist")

	// Проверяем структуру таблицы
	columns, err := setup.db.Migrator().ColumnTypes(&entity.User{})
	require.NoError(t, err)

	// Проверяем наличие необходимых колонок
	columnNames := make(map[string]bool)
	for _, col := range columns {
		columnNames[col.Name()] = true
	}

	assert.True(t, columnNames["id"], "ID column should exist")
	assert.True(t, columnNames["username"], "Username column should exist")
	assert.True(t, columnNames["email"], "Email column should exist")
}

func TestUserRepositoryIntegration(t *testing.T) {
	setup := setupDatabaseTest(t)
	defer setup.db.Migrator().DropTable(&entity.User{})

	t.Run("Create and Retrieve User", func(t *testing.T) {
		// Создаем пользователя
		user, err := entity.NewUser("testuser", "test@example.com")
		require.NoError(t, err)

		// Сохраняем в БД
		err = setup.userRepo.Create(user)
		require.NoError(t, err)
		assert.NotZero(t, user.ID, "User ID should be set after creation")

		// Получаем из БД
		retrievedUser, err := setup.userRepo.GetByID(user.ID)
		require.NoError(t, err)
		assert.Equal(t, user.ID, retrievedUser.ID)
		assert.Equal(t, "testuser", retrievedUser.Username)
		assert.Equal(t, "test@example.com", retrievedUser.Email)
	})

	t.Run("Update User", func(t *testing.T) {
		// Создаем пользователя
		user, err := entity.NewUser("updateuser", "update@example.com")
		require.NoError(t, err)

		err = setup.userRepo.Create(user)
		require.NoError(t, err)

		// Обновляем пользователя
		user.Username = "updateduser"
		user.Email = "updated@example.com"

		err = setup.userRepo.Update(user)
		require.NoError(t, err)

		// Проверяем обновление
		retrievedUser, err := setup.userRepo.GetByID(user.ID)
		require.NoError(t, err)
		assert.Equal(t, "updateduser", retrievedUser.Username)
		assert.Equal(t, "updated@example.com", retrievedUser.Email)
	})

	t.Run("Get All Users", func(t *testing.T) {
		// Создаем несколько пользователей
		user1, _ := entity.NewUser("user1", "user1@example.com")
		user2, _ := entity.NewUser("user2", "user2@example.com")
		user3, _ := entity.NewUser("user3", "user3@example.com")

		setup.userRepo.Create(user1)
		setup.userRepo.Create(user2)
		setup.userRepo.Create(user3)

		// Получаем всех пользователей
		users, err := setup.userRepo.GetAll()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(users), 3, "Should have at least 3 users")
	})

	t.Run("Get User by Email", func(t *testing.T) {
		// Создаем пользователя
		user, err := entity.NewUser("emailuser", "email@example.com")
		require.NoError(t, err)

		err = setup.userRepo.Create(user)
		require.NoError(t, err)

		// Получаем по email
		retrievedUser, err := setup.userRepo.GetByEmail("email@example.com")
		require.NoError(t, err)
		assert.Equal(t, user.ID, retrievedUser.ID)
		assert.Equal(t, "emailuser", retrievedUser.Username)
	})

	t.Run("Get User by Username", func(t *testing.T) {
		// Создаем пользователя
		user, err := entity.NewUser("usernameuser", "username@example.com")
		require.NoError(t, err)

		err = setup.userRepo.Create(user)
		require.NoError(t, err)

		// Получаем по username
		retrievedUser, err := setup.userRepo.GetByUsername("usernameuser")
		require.NoError(t, err)
		assert.Equal(t, user.ID, retrievedUser.ID)
		assert.Equal(t, "username@example.com", retrievedUser.Email)
	})
}

func TestUserServiceIntegration(t *testing.T) {
	setup := setupDatabaseTest(t)
	defer setup.db.Migrator().DropTable(&entity.User{})

	t.Run("Register User Through Service", func(t *testing.T) {
		// Регистрируем пользователя через сервис
		user, err := setup.userService.RegisterUser("serviceuser", "service@example.com")
		require.NoError(t, err)
		assert.NotZero(t, user.ID)

		// Проверяем, что пользователь сохранен в БД
		retrievedUser, err := setup.userRepo.GetByID(user.ID)
		require.NoError(t, err)
		assert.Equal(t, "serviceuser", retrievedUser.Username)
		assert.Equal(t, "service@example.com", retrievedUser.Email)
	})

	t.Run("Get User Through Service", func(t *testing.T) {
		// Создаем пользователя
		user, err := setup.userService.RegisterUser("getuser", "get@example.com")
		require.NoError(t, err)

		// Получаем через сервис
		retrievedUser, err := setup.userService.GetUser(user.ID)
		require.NoError(t, err)
		assert.Equal(t, user.ID, retrievedUser.ID)
		assert.Equal(t, "getuser", retrievedUser.Username)
	})

	t.Run("Update User Through Service", func(t *testing.T) {
		// Создаем пользователя
		user, err := setup.userService.RegisterUser("updateuser", "update@example.com")
		require.NoError(t, err)

		// Обновляем через сервис
		updatedUser, err := setup.userService.UpdateUser(user.ID, "updateduser", "updated@example.com")
		require.NoError(t, err)
		assert.Equal(t, "updateduser", updatedUser.Username)
		assert.Equal(t, "updated@example.com", updatedUser.Email)

		// Проверяем в БД
		dbUser, err := setup.userRepo.GetByID(user.ID)
		require.NoError(t, err)
		assert.Equal(t, "updateduser", dbUser.Username)
		assert.Equal(t, "updated@example.com", dbUser.Email)
	})

	t.Run("Duplicate Email Registration", func(t *testing.T) {
		// Регистрируем первого пользователя
		_, err := setup.userService.RegisterUser("user1", "duplicate@example.com")
		require.NoError(t, err)

		// Пытаемся зарегистрировать второго с тем же email
		_, err = setup.userService.RegisterUser("user2", "duplicate@example.com")
		assert.Error(t, err, "Should fail with duplicate email")
	})

	t.Run("Duplicate Username Registration", func(t *testing.T) {
		// Регистрируем первого пользователя
		_, err := setup.userService.RegisterUser("duplicateuser", "email1@example.com")
		require.NoError(t, err)

		// Пытаемся зарегистрировать второго с тем же username
		_, err = setup.userService.RegisterUser("duplicateuser", "email2@example.com")
		assert.Error(t, err, "Should fail with duplicate username")
	})
}

func TestDatabaseConstraints(t *testing.T) {
	setup := setupDatabaseTest(t)
	defer setup.db.Migrator().DropTable(&entity.User{})

	t.Run("Unique Email Constraint", func(t *testing.T) {
		// Создаем первого пользователя
		user1, err := entity.NewUser("user1", "unique@example.com")
		require.NoError(t, err)
		err = setup.userRepo.Create(user1)
		require.NoError(t, err)

		// Пытаемся создать второго с тем же email
		user2, err := entity.NewUser("user2", "unique@example.com")
		require.NoError(t, err)
		err = setup.userRepo.Create(user2)
		assert.Error(t, err, "Should fail due to unique constraint on email")
	})

	t.Run("Unique Username Constraint", func(t *testing.T) {
		// Создаем первого пользователя
		user1, err := entity.NewUser("uniqueuser", "email1@example.com")
		require.NoError(t, err)
		err = setup.userRepo.Create(user1)
		require.NoError(t, err)

		// Пытаемся создать второго с тем же username
		user2, err := entity.NewUser("uniqueuser", "email2@example.com")
		require.NoError(t, err)
		err = setup.userRepo.Create(user2)
		assert.Error(t, err, "Should fail due to unique constraint on username")
	})
}

func TestDatabaseTransactionRollback(t *testing.T) {
	setup := setupDatabaseTest(t)
	defer setup.db.Migrator().DropTable(&entity.User{})

	t.Run("Transaction Rollback on Error", func(t *testing.T) {
		// Начинаем транзакцию
		tx := setup.db.Begin()
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		// Создаем пользователя в транзакции
		user, err := entity.NewUser("txuser", "tx@example.com")
		require.NoError(t, err)

		userRepo := repository.NewUserRepository(tx)
		err = userRepo.Create(user)
		require.NoError(t, err)

		// Проверяем, что пользователь существует в транзакции
		retrievedUser, err := userRepo.GetByID(user.ID)
		require.NoError(t, err)
		assert.Equal(t, "txuser", retrievedUser.Username)

		// Откатываем транзакцию
		tx.Rollback()

		// Проверяем, что пользователь не существует после отката
		_, err = setup.userRepo.GetByID(user.ID)
		assert.Error(t, err, "User should not exist after rollback")
	})
}

func TestDatabasePerformance(t *testing.T) {
	setup := setupDatabaseTest(t)
	defer setup.db.Migrator().DropTable(&entity.User{})

	t.Run("Bulk User Creation", func(t *testing.T) {
		// Создаем много пользователей
		users := make([]*entity.User, 100)
		for i := 0; i < 100; i++ {
			user, err := entity.NewUser(
				fmt.Sprintf("user%d", i),
				fmt.Sprintf("user%d@example.com", i),
			)
			require.NoError(t, err)
			users[i] = user
		}

		// Сохраняем всех пользователей
		for _, user := range users {
			err := setup.userRepo.Create(user)
			require.NoError(t, err)
		}

		// Проверяем, что все пользователи созданы
		allUsers, err := setup.userRepo.GetAll()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(allUsers), 100)
	})

	t.Run("Concurrent User Creation", func(t *testing.T) {
		// Создаем пользователей конкурентно
		const numGoroutines = 10
		done := make(chan bool, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer func() { done <- true }()

				user, err := entity.NewUser(
					fmt.Sprintf("concurrentuser%d", id),
					fmt.Sprintf("concurrent%d@example.com", id),
				)
				if err != nil {
					t.Errorf("Failed to create user: %v", err)
					return
				}

				err = setup.userRepo.Create(user)
				if err != nil {
					t.Errorf("Failed to save user: %v", err)
					return
				}
			}(i)
		}

		// Ждем завершения всех горутин
		for i := 0; i < numGoroutines; i++ {
			<-done
		}

		// Проверяем результат
		allUsers, err := setup.userRepo.GetAll()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(allUsers), numGoroutines)
	})
}
