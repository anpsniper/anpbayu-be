package models

import (
	"time"
)

// User represents a user record in the database.
// This struct will be used for storing and retrieving user data.
type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name,omitempty"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`              // Password should not be marshaled to JSON
	RoleID    string    `json:"role_id"`        // Foreign key to the roles table
	Role      *Role     `json:"role,omitempty"` // Embedded Role struct for eager loading, omitempty to exclude if nil
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserCreateRequest represents the expected payload for creating a new user (e.g., registration).
// This would be used in a registration controller.
type UserCreateRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name,omitempty"` // Optional name field for registration
}

// UserResponse represents the structure of a user object returned in API responses.
// It typically excludes sensitive information like the password hash.
type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name,omitempty"` // Example: optional name field
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
