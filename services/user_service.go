package services

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt" // For password hashing

	"github.com/anpsniper/anpbayu-be/database" // Import your database package
	"github.com/anpsniper/anpbayu-be/models"   // Import your models package
)

// UserService defines the interface for user-related operations.
type UserService interface {
	CreateUser(email, password string) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	// Add other user-related methods here (e.g., UpdateUser, DeleteUser)
}

// UserServiceImpl implements the UserService interface.
type UserServiceImpl struct {
	db *sql.DB
}

// NewUserService creates and returns a new instance of UserServiceImpl.
func NewUserService() *UserServiceImpl {
	return &UserServiceImpl{
		db: database.DB, // Use the global DB connection from the database package
	}
}

// CreateUser hashes the password and inserts a new user into the database.
func (s *UserServiceImpl) CreateUser(email, password string) (*models.User, error) {
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		return nil, fmt.Errorf("failed to hash password")
	}

	// Prepare the SQL insert statement
	query := `
		INSERT INTO db_bayneta.users (email, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, email, created_at, updated_at;
	`
	// Note: 'name' is optional, so we don't include it in this basic create.
	// You'd add it if your registration form collected it.

	now := time.Now()
	var user models.User
	// Execute the query and scan the returned values into the user struct
	err = s.db.QueryRow(query, email, string(hashedPassword), now, now).Scan(
		&user.ID,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		// Check for unique constraint violation (e.g., email already exists)
		if err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"` {
			return nil, fmt.Errorf("user with this email already exists")
		}
		log.Printf("Error creating user in database: %v", err)
		return nil, fmt.Errorf("failed to create user")
	}

	log.Printf("User created successfully: %s", user.Email)
	return &user, nil
}

// GetUserByEmail retrieves a user from the database by their email.
func (s *UserServiceImpl) GetUserByEmail(email string) (*models.User, error) {
	query := `
		SELECT id, email, password, created_at, updated_at, name
		FROM db_bayneta.users
		WHERE email = $1;
	`
	var user models.User
	var name sql.NullString // Use sql.NullString for nullable columns

	err := s.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
		&name, // Scan into sql.NullString
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		log.Printf("Error getting user by email from database: %v", err)
		return nil, fmt.Errorf("failed to retrieve user")
	}

	if name.Valid {
		user.Name = name.String
	}

	return &user, nil
}

// ComparePasswords compares a plaintext password with a hashed password.
func ComparePasswords(hashedPassword, password []byte) bool {
	err := bcrypt.CompareHashAndPassword(hashedPassword, password)
	return err == nil
}
