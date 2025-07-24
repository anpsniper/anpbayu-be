package services

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/anpsniper/anpbayu-be/database"
	"github.com/anpsniper/anpbayu-be/models"
	"github.com/google/uuid"
)

// RoleServiceInterface defines the methods that any role service implementation must provide.
type RoleServiceInterface interface {
	GetAllRoles(search string, page, limit int) ([]models.Role, int, int, error) // Returns roles, totalPages, totalItems
	GetRoleByID(id string) (*models.Role, error)
	GetRoleByName(name string) (*models.Role, error) // Added for convenience
	CreateRole(role *models.Role) error
	UpdateRole(role *models.Role) error
	DeleteRole(id string) error
}

// RoleService provides methods for role-related business logic, implementing RoleServiceInterface.
type RoleService struct {
	// No fields needed here if we're using a global DB connection from the database package.
}

// NewRoleService creates and returns a new RoleService instance.
func NewRoleService() *RoleService {
	return &RoleService{}
}

// GetAllRoles fetches all roles from the database with search and pagination.
func (s *RoleService) GetAllRoles(search string, page, limit int) ([]models.Role, int, int, error) {
	if database.DB == nil {
		return nil, 0, 0, fmt.Errorf("database connection is not initialized")
	}

	var roles []models.Role
	var totalItems int

	// Build the base query
	countQuery := "SELECT COUNT(id) FROM roles WHERE 1=1"
	selectQuery := "SELECT id, name, description, created_at, updated_at FROM roles WHERE 1=1"
	args := []interface{}{}
	argCounter := 1

	// Add search condition if provided
	if search != "" {
		searchPattern := "%" + search + "%"
		countQuery += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)", argCounter, argCounter+1)
		selectQuery += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d)", argCounter, argCounter+1)
		args = append(args, searchPattern, searchPattern)
		argCounter += 2
	}

	// Get total items
	err := database.DB.QueryRow(countQuery, args...).Scan(&totalItems)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to count roles: %w", err)
	}

	// Calculate pagination offsets
	offset := (page - 1) * limit
	selectQuery += fmt.Sprintf(" ORDER BY name ASC LIMIT $%d OFFSET $%d", argCounter, argCounter+1)
	args = append(args, limit, offset)

	rows, err := database.DB.Query(selectQuery, args...)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to query roles: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var role models.Role
		err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt)
		if err != nil {
			log.Printf("Error scanning role row: %v", err)
			return nil, 0, 0, fmt.Errorf("failed to scan role: %w", err)
		}
		roles = append(roles, role)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, 0, fmt.Errorf("error iterating role rows: %w", err)
	}

	totalPages := (totalItems + limit - 1) / limit
	if totalPages == 0 && totalItems > 0 { // Handle case where totalItems < limit
		totalPages = 1
	}

	return roles, totalPages, totalItems, nil
}

// GetRoleByID fetches a role by its ID.
func (s *RoleService) GetRoleByID(id string) (*models.Role, error) {
	if database.DB == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	role := &models.Role{}
	query := "SELECT id, name, description, created_at, updated_at FROM roles WHERE id = $1"
	err := database.DB.QueryRow(query, id).Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil // Role not found
	}
	if err != nil {
		log.Printf("Error fetching role by ID %s: %v", id, err)
		return nil, fmt.Errorf("failed to fetch role by ID: %w", err)
	}
	return role, nil
}

// GetRoleByName fetches a role by its name.
func (s *RoleService) GetRoleByName(name string) (*models.Role, error) {
	if database.DB == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	role := &models.Role{}
	query := "SELECT id, name, description, created_at, updated_at FROM roles WHERE name = $1"
	err := database.DB.QueryRow(query, name).Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil // Role not found
	}
	if err != nil {
		log.Printf("Error fetching role by name %s: %v", name, err)
		return nil, fmt.Errorf("failed to fetch role by name: %w", err)
	}
	return role, nil
}

// CreateRole inserts a new role into the database.
func (s *RoleService) CreateRole(role *models.Role) error {
	if database.DB == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	// Generate a new UUID for the role
	role.ID = uuid.New().String()
	role.CreatedAt = time.Now()
	role.UpdatedAt = time.Now()

	query := `
		INSERT INTO roles (id, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := database.DB.Exec(
		query,
		role.ID,
		role.Name,
		role.Description,
		role.CreatedAt,
		role.UpdatedAt,
	)
	if err != nil {
		log.Printf("Error creating role %s: %v", role.Name, err)
		return fmt.Errorf("failed to create role: %w", err)
	}
	return nil
}

// UpdateRole updates an existing role's information in the database.
func (s *RoleService) UpdateRole(role *models.Role) error {
	if database.DB == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	role.UpdatedAt = time.Now() // Update the timestamp

	query := `
		UPDATE roles
		SET name = $1, description = $2, updated_at = $3
		WHERE id = $4
	`
	result, err := database.DB.Exec(
		query,
		role.Name,
		role.Description,
		role.UpdatedAt,
		role.ID,
	)
	if err != nil {
		log.Printf("Error updating role %s: %v", role.ID, err)
		return fmt.Errorf("failed to update role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected after update: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("role with ID %s not found for update", role.ID)
	}

	return nil
}

// DeleteRole deletes a role from the database by its ID.
func (s *RoleService) DeleteRole(id string) error {
	if database.DB == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	query := `DELETE FROM roles WHERE id = $1`
	result, err := database.DB.Exec(query, id)
	if err != nil {
		log.Printf("Error deleting role by ID %s: %v", id, err)
		return fmt.Errorf("failed to delete role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected after delete: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("role with ID %s not found for deletion", id)
	}

	return nil
}
