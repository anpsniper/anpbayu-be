package models

import (
	"time"
)

// User represents a user record in the database.
// This struct will be used for storing and retrieving user data.
type User struct {
	ID           string    `json:"id"`             // Unique identifier for the user
	Email        string    `json:"email"`          // User's email, typically unique
	PasswordHash string    `json:"-"`              // Hashed password, omitted from JSON output
	CreatedAt    time.Time `json:"created_at"`     // Timestamp when the user was created
	UpdatedAt    time.Time `json:"updated_at"`     // Timestamp when the user record was last updated
	Name         string    `json:"name,omitempty"` // User's full name (optional), added for GetProfile
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
