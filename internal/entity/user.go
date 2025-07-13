// (Слой сущностей)
package entity

import (
	"errors"
	"regexp"
	"strings"
)

type User struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Username string `gorm:"unique" json:"username"`
	Email    string `gorm:"unique" json:"email"`
}

// NewUser создает нового пользователя с валидацией
func NewUser(username, email string) (*User, error) {
	user := &User{
		Username: strings.TrimSpace(username),
		Email:    strings.TrimSpace(email),
	}

	if err := user.Validate(); err != nil {
		return nil, err
	}

	return user, nil
}

// Validate проверяет корректность данных пользователя
func (u *User) Validate() error {
	if u.Username == "" {
		return errors.New("username cannot be empty")
	}

	if len(u.Username) < 3 {
		return errors.New("username must be at least 3 characters long")
	}

	if len(u.Username) > 50 {
		return errors.New("username cannot exceed 50 characters")
	}

	if u.Email == "" {
		return errors.New("email cannot be empty")
	}

	// Простая валидация email
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(u.Email) {
		return errors.New("invalid email format")
	}

	return nil
}

// Update обновляет данные пользователя
func (u *User) Update(username, email string) error {
	oldUsername := u.Username
	oldEmail := u.Email

	u.Username = strings.TrimSpace(username)
	u.Email = strings.TrimSpace(email)

	if err := u.Validate(); err != nil {
		// Откатываем изменения при ошибке валидации
		u.Username = oldUsername
		u.Email = oldEmail
		return err
	}

	return nil
}

// IsValidEmail проверяет корректность email
func (u *User) IsValidEmail() bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(u.Email)
}

// GetDisplayName возвращает отображаемое имя пользователя
func (u *User) GetDisplayName() string {
	if u.Username != "" {
		return u.Username
	}
	// Если username пустой, возвращаем часть email до @
	if u.Email != "" {
		parts := strings.Split(u.Email, "@")
		if len(parts) > 0 {
			return parts[0]
		}
	}
	return "Unknown User"
}
