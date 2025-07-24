package models

import (
	"database/sql" // Still needed for sql.ErrNoRows
	"fmt"
	"log"
	"time"

	"github.com/anpsniper/anpbayu-be/database" // Import the database package
	"github.com/google/uuid"                   // For generating UUIDs
	"golang.org/x/crypto/bcrypt"               // For password hashing
)

// SeedRoles ensures that default roles (admin, user) exist in the database.
// It now uses the global DB connection from the database package.
func SeedRoles() error {
	// Ensure the database connection is available
	if database.DB == nil {
		return fmt.Errorf("database connection is not initialized. Call database.InitDatabase() first")
	}

	rolesToSeed := []struct {
		Name        string
		Description string
	}{
		{"admin", "Administrator role with full system access."},
		{"user", "Standard user role with general access."},
	}

	for _, roleData := range rolesToSeed {
		var existingRoleID string
		// Check if the role already exists by name
		err := database.DB.QueryRow("SELECT id FROM roles WHERE name = $1", roleData.Name).Scan(&existingRoleID)

		if err == sql.ErrNoRows {
			// Role does not exist, insert it
			newRoleID := uuid.New().String() // Generate a new UUID for the role
			_, err := database.DB.Exec(
				"INSERT INTO roles (id, name, description, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)",
				newRoleID,
				roleData.Name,
				roleData.Description,
				time.Now(),
				time.Now(),
			)
			if err != nil {
				return fmt.Errorf("failed to insert role %s: %w", roleData.Name, err)
			}
			log.Printf("Role '%s' seeded successfully with ID: %s", roleData.Name, newRoleID)
		} else if err != nil {
			// Other database error
			return fmt.Errorf("failed to check for existing role %s: %w", roleData.Name, err)
		} else {
			// Role already exists
			log.Printf("Role '%s' already exists with ID: %s", roleData.Name, existingRoleID)
		}
	}
	return nil
}

// SeedExampleUser creates an example user and assigns them the 'user' role.
// It now uses the global DB connection from the database package.
func SeedExampleUser() error {
	// Ensure the database connection is available
	if database.DB == nil {
		return fmt.Errorf("database connection is not initialized. Call database.InitDatabase() first")
	}

	// Find the 'user' role ID
	var userRoleID string
	err := database.DB.QueryRow("SELECT id FROM roles WHERE name = 'user'").Scan(&userRoleID)
	if err == sql.ErrNoRows {
		return fmt.Errorf("'user' role not found. Please ensure roles are seeded first")
	}
	if err != nil {
		return fmt.Errorf("failed to retrieve 'user' role ID: %w", err)
	}

	exampleEmail := "user@example.com"
	exampleUsername := "exampleuser"
	examplePassword := "password123" // This should be a strong, default password

	// Check if the example user already exists
	var existingUserID string
	err = database.DB.QueryRow("SELECT id FROM users WHERE email = $1", exampleEmail).Scan(&existingUserID)

	if err == sql.ErrNoRows {
		// User does not exist, create and insert them
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(examplePassword), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password for example user: %w", err)
		}

		newUserID := uuid.New().String() // Generate a new UUID for the user
		_, err = database.DB.Exec(
			"INSERT INTO users (id, username, email, password_hash, role_id, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)",
			newUserID,
			exampleUsername,
			exampleEmail,
			string(hashedPassword),
			userRoleID,
			time.Now(),
			time.Now(),
		)
		if err != nil {
			return fmt.Errorf("failed to insert example user: %w", err)
		}
		log.Printf("Example user '%s' seeded successfully with ID: %s and Role ID: %s", exampleEmail, newUserID, userRoleID)
	} else if err != nil {
		// Other database error
		return fmt.Errorf("failed to check for existing example user: %w", err)
	} else {
		// User already exists
		log.Printf("Example user '%s' already exists with ID: %s and Role ID: %s", exampleEmail, existingUserID, userRoleID)
	}

	return nil
}
