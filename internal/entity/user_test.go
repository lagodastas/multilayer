package entity

import (
	"strings"
	"testing"
)

func TestNewUser(t *testing.T) {
	tests := []struct {
		name     string
		username string
		email    string
		wantErr  bool
	}{
		{
			name:     "Valid user",
			username: "john_doe",
			email:    "john@example.com",
			wantErr:  false,
		},
		{
			name:     "Empty username",
			username: "",
			email:    "john@example.com",
			wantErr:  true,
		},
		{
			name:     "Empty email",
			username: "john_doe",
			email:    "",
			wantErr:  true,
		},
		{
			name:     "Invalid email format",
			username: "john_doe",
			email:    "invalid-email",
			wantErr:  true,
		},
		{
			name:     "Username too short",
			username: "jo",
			email:    "john@example.com",
			wantErr:  true,
		},
		{
			name:     "Username with spaces",
			username: "  john_doe  ",
			email:    "john@example.com",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := NewUser(tt.username, tt.email)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewUser() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("NewUser() unexpected error: %v", err)
				return
			}

			if user.Username != tt.username {
				// Проверяем, что пробелы были удалены
				expected := strings.TrimSpace(tt.username)
				if user.Username != expected {
					t.Errorf("NewUser() username = %v, want %v", user.Username, expected)
				}
			}

			if user.Email != tt.email {
				// Проверяем, что пробелы были удалены
				expected := strings.TrimSpace(tt.email)
				if user.Email != expected {
					t.Errorf("NewUser() email = %v, want %v", user.Email, expected)
				}
			}
		})
	}
}

func TestUser_Validate(t *testing.T) {
	tests := []struct {
		name    string
		user    *User
		wantErr bool
	}{
		{
			name: "Valid user",
			user: &User{
				Username: "john_doe",
				Email:    "john@example.com",
			},
			wantErr: false,
		},
		{
			name: "Empty username",
			user: &User{
				Username: "",
				Email:    "john@example.com",
			},
			wantErr: true,
		},
		{
			name: "Username too short",
			user: &User{
				Username: "jo",
				Email:    "john@example.com",
			},
			wantErr: true,
		},
		{
			name: "Username too long",
			user: &User{
				Username: "very_long_username_that_exceeds_fifty_characters_limit_12345",
				Email:    "john@example.com",
			},
			wantErr: true,
		},
		{
			name: "Empty email",
			user: &User{
				Username: "john_doe",
				Email:    "",
			},
			wantErr: true,
		},
		{
			name: "Invalid email format",
			user: &User{
				Username: "john_doe",
				Email:    "invalid-email",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("User.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUser_Update(t *testing.T) {
	user := &User{
		ID:       1,
		Username: "old_username",
		Email:    "old@example.com",
	}

	tests := []struct {
		name     string
		username string
		email    string
		wantErr  bool
	}{
		{
			name:     "Valid update",
			username: "new_username",
			email:    "new@example.com",
			wantErr:  false,
		},
		{
			name:     "Invalid email format",
			username: "new_username",
			email:    "invalid-email",
			wantErr:  true,
		},
		{
			name:     "Username too short",
			username: "ne",
			email:    "new@example.com",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalUsername := user.Username
			originalEmail := user.Email

			err := user.Update(tt.username, tt.email)

			if tt.wantErr {
				if err == nil {
					t.Errorf("User.Update() expected error but got none")
				}
				// Проверяем, что данные не изменились при ошибке
				if user.Username != originalUsername {
					t.Errorf("User.Update() username changed despite error: got %v, want %v", user.Username, originalUsername)
				}
				if user.Email != originalEmail {
					t.Errorf("User.Update() email changed despite error: got %v, want %v", user.Email, originalEmail)
				}
				return
			}

			if err != nil {
				t.Errorf("User.Update() unexpected error: %v", err)
				return
			}

			if user.Username != tt.username {
				t.Errorf("User.Update() username = %v, want %v", user.Username, tt.username)
			}
			if user.Email != tt.email {
				t.Errorf("User.Update() email = %v, want %v", user.Email, tt.email)
			}
		})
	}
}

func TestUser_IsValidEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{
			name:  "Valid email",
			email: "john@example.com",
			want:  true,
		},
		{
			name:  "Valid email with subdomain",
			email: "john@sub.example.com",
			want:  true,
		},
		{
			name:  "Invalid email - no @",
			email: "invalid-email",
			want:  false,
		},
		{
			name:  "Invalid email - no domain",
			email: "john@",
			want:  false,
		},
		{
			name:  "Invalid email - no local part",
			email: "@example.com",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{Email: tt.email}
			if got := user.IsValidEmail(); got != tt.want {
				t.Errorf("User.IsValidEmail() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_GetDisplayName(t *testing.T) {
	tests := []struct {
		name string
		user *User
		want string
	}{
		{
			name: "With username",
			user: &User{
				Username: "john_doe",
				Email:    "john@example.com",
			},
			want: "john_doe",
		},
		{
			name: "Without username, with email",
			user: &User{
				Username: "",
				Email:    "john@example.com",
			},
			want: "john",
		},
		{
			name: "Without username and email",
			user: &User{
				Username: "",
				Email:    "",
			},
			want: "Unknown User",
		},
		{
			name: "Email with special characters",
			user: &User{
				Username: "",
				Email:    "john.doe+test@example.com",
			},
			want: "john.doe+test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.user.GetDisplayName(); got != tt.want {
				t.Errorf("User.GetDisplayName() = %v, want %v", got, tt.want)
			}
		})
	}
}
