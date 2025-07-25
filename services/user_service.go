package services

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/anpsniper/anpbayu-be/database" // Import the database package
	"github.com/anpsniper/anpbayu-be/models"   // Import the models package
	"github.com/google/uuid"                   // For generating UUIDs
	"golang.org/x/crypto/bcrypt"
)

// UserServiceInterface defines the methods that any user service implementation must provide.
// This allows for dependency inversion and easier testing (e.g., by mocking the service).
type UserServiceInterface interface {
	GetAllUsers(search string, page, limit int) ([]models.User, int, int, error) // Returns users, totalPages, totalItems
	GetUserByID(id string) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	CreateUser(user *models.User) error
	UpdateUser(req *models.UpdateUserRequest) error
	DeleteUser(id string) error
	GetAllRoles() ([]models.LstRole, error)
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

func (s *UserService) GetAllUsers(search string, page, limit int) ([]models.User, int, int, error) {
	if database.DB == nil {
		return nil, 0, 0, fmt.Errorf("database connection is not initialized")
	}

	var users []models.User
	var totalItems int

	// Build the base query
	countQuery := "SELECT COUNT(a.id) FROM users a LEFT JOIN roles b ON a.role_id = b.id WHERE 1=1"
	selectQuery := "SELECT a.id, a.username, a.email, a.role_id, b.name AS role_name, a.created_at, a.updated_at FROM users a LEFT JOIN roles b ON a.role_id = b.id WHERE 1=1"
	args := []interface{}{}
	argCounter := 1

	// Add search condition if provided
	if search != "" {
		searchPattern := "%" + search + "%"
		// Ensure search applies to both username/email and role name if desired
		countQuery += fmt.Sprintf(" AND (a.username ILIKE $%d OR a.email ILIKE $%d OR b.name ILIKE $%d)", argCounter, argCounter+1, argCounter+2)
		selectQuery += fmt.Sprintf(" AND (a.username ILIKE $%d OR a.email ILIKE $%d OR b.name ILIKE $%d)", argCounter, argCounter+1, argCounter+2)
		args = append(args, searchPattern, searchPattern, searchPattern)
		argCounter += 3 // Increment by 3 for 3 placeholders
	}

	// Get total items
	err := database.DB.QueryRow(countQuery, args...).Scan(&totalItems)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Calculate pagination offsets
	offset := (page - 1) * limit
	selectQuery += fmt.Sprintf(" ORDER BY a.username ASC LIMIT $%d OFFSET $%d", argCounter, argCounter+1)
	args = append(args, limit, offset)

	rows, err := database.DB.Query(selectQuery, args...)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var user models.User
		// Scan directly into user.RoleName
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.RoleID, &user.RoleName, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			log.Printf("Error scanning user row: %v", err)
			return nil, 0, 0, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, 0, fmt.Errorf("error iterating user rows: %w", err)
	}

	totalPages := (totalItems + limit - 1) / limit
	if totalPages == 0 && totalItems > 0 { // Handle case where totalItems < limit
		totalPages = 1
	}

	return users, totalPages, totalItems, nil
}

func (s *UserService) GetAllRoles() ([]models.LstRole, error) {
	if database.DB == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	var roles []models.LstRole

	// Query only for ID and Name, as models.LstRole likely only contains these fields.
	query := `SELECT id, name FROM roles ORDER BY name ASC`

	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query roles: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var role models.LstRole
		// Scan only ID and Name
		err := rows.Scan(&role.ID, &role.Name)
		if err != nil {
			log.Printf("Error scanning role row: %v", err)
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}
		roles = append(roles, role)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating role rows: %w", err)
	}

	return roles, nil
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
func (s *UserService) UpdateUser(req *models.UpdateUserRequest) error {
	if database.DB == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	// Start building the query and arguments
	// Always update username, email, role_id, and updated_at
	query := "UPDATE users SET username = $1, email = $2, role_id = $3, updated_at = $4"
	args := []interface{}{
		req.Username,
		req.Email,
		req.RoleID,
		time.Now(), // updated_at
	}
	argCounter := 5 // Next placeholder will be $5 for WHERE clause or password

	// Conditionally add password update if a new password is provided
	if req.Password != nil && *req.Password != "" {
		hashedPassword, err := HashPassword(*req.Password)
		if err != nil {
			return fmt.Errorf("failed to hash new password: %w", err)
		}
		query += fmt.Sprintf(", password_hash = $%d", argCounter)
		args = append(args, hashedPassword)
		argCounter++
	}

	// Add the WHERE clause
	query += fmt.Sprintf(" WHERE id = $%d", argCounter)
	args = append(args, req.ID)

	result, err := database.DB.Exec(query, args...)
	if err != nil {
		log.Printf("Error updating user %s: %v", req.ID, err)
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected after update: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("user with ID %s not found for update", req.ID)
	}

	return nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
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
