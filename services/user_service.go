package services

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/anpsniper/anpbayu-be/database" // Import the database package
	"github.com/anpsniper/anpbayu-be/models"   // Import the models package
	"github.com/google/uuid"                   // For generating UUIDs
)

// UserServiceInterface defines the methods that any user service implementation must provide.
// This allows for dependency inversion and easier testing (e.g., by mocking the service).
type UserServiceInterface interface {
	GetUserByID(id string) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	CreateUser(user *models.User) error
	UpdateUser(user *models.User) error
	DeleteUser(id string) error
}

// UserService provides methods for user-related business logic, implementing UserServiceInterface.
type UserService struct {
	// No fields needed here if we're using a global DB connection from the database package.
	// If you were to pass the DB connection, it would be: db *sql.DB
}

// NewUserService creates and returns a new UserService instance.
// It returns the concrete *UserService type, which satisfies the UserServiceInterface.
func NewUserService() *UserService { // Changed return type to *UserService
	return &UserService{}
}

// GetUserByID fetches a user by their ID, including their associated role.
func (s *UserService) GetUserByID(id string) (*models.User, error) {
	if database.DB == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	user := &models.User{}
	role := &models.Role{} // To store role data

	query := `
		SELECT
			u.id, u.username, u.email, u.password_hash, u.role_id, u.created_at, u.updated_at,
			r.id, r.name, r.description, r.created_at, r.updated_at
		FROM
			users u
		JOIN
			roles r ON u.role_id = r.id
		WHERE
			u.id = $1
	`
	err := database.DB.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password, &user.RoleID, &user.CreatedAt, &user.UpdatedAt,
		&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // User not found
	}
	if err != nil {
		log.Printf("Error fetching user by ID %s: %v", id, err)
		return nil, fmt.Errorf("failed to fetch user by ID: %w", err)
	}

	user.Role = role // Assign the fetched role to the user
	return user, nil
}

// GetUserByEmail fetches a user by their email, including their associated role.
func (s *UserService) GetUserByEmail(email string) (*models.User, error) {
	if database.DB == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	user := &models.User{}
	role := &models.Role{} // To store role data

	query := `
		SELECT
			u.id, u.username, u.email, u.password_hash, u.role_id, u.created_at, u.updated_at,
			r.id, r.name, r.description, r.created_at, r.updated_at
		FROM
			users u
		JOIN
			roles r ON u.role_id = r.id
		WHERE
			u.email = $1
	`
	err := database.DB.QueryRow(query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.Password, &user.RoleID, &user.CreatedAt, &user.UpdatedAt,
		&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // User not found
	}
	if err != nil {
		log.Printf("Error fetching user by email %s: %v", email, err)
		return nil, fmt.Errorf("failed to fetch user by email: %w", err)
	}

	user.Role = role // Assign the fetched role to the user
	return user, nil
}

// CreateUser inserts a new user into the database.
func (s *UserService) CreateUser(user *models.User) error {
	if database.DB == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	// Generate a new UUID for the user
	user.ID = uuid.New().String()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	query := `
		INSERT INTO users (id, username, email, password_hash, role_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := database.DB.Exec(
		query,
		user.ID,
		user.Username,
		user.Email,
		user.Password, // This should be the hashed password
		user.RoleID,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		log.Printf("Error creating user %s: %v", user.Email, err)
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// UpdateUser updates an existing user's information in the database.
// It updates username, email, and role_id. Password is not updated here.
func (s *UserService) UpdateUser(user *models.User) error {
	if database.DB == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	user.UpdatedAt = time.Now() // Update the timestamp

	query := `
		UPDATE users
		SET username = $1, email = $2, role_id = $3, updated_at = $4
		WHERE id = $5
	`
	result, err := database.DB.Exec(
		query,
		user.Username,
		user.Email,
		user.RoleID,
		user.UpdatedAt,
		user.ID,
	)
	if err != nil {
		log.Printf("Error updating user %s: %v", user.ID, err)
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected after update: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user with ID %s not found for update", user.ID)
	}

	return nil
}

// DeleteUser deletes a user from the database by their ID.
func (s *UserService) DeleteUser(id string) error {
	if database.DB == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	query := `DELETE FROM users WHERE id = $1`
	result, err := database.DB.Exec(query, id)
	if err != nil {
		log.Printf("Error deleting user by ID %s: %v", id, err)
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected after delete: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user with ID %s not found for deletion", id)
	}

	return nil
}
