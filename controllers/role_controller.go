package controllers

import (
	"fmt"
	"log"
	"net/http"
	"strconv" // For converting string to int

	"github.com/gofiber/fiber/v2"

	"github.com/anpsniper/anpbayu-be/models"   // Import models package for Role struct
	"github.com/anpsniper/anpbayu-be/services" // Import services package for RoleServiceInterface
)

// RoleController handles role-related requests.
type RoleController struct {
	RoleService services.RoleServiceInterface // RoleService dependency (interface)
}

// NewRoleController creates and returns a new RoleController instance.
func NewRoleController(roleService services.RoleServiceInterface) *RoleController {
	return &RoleController{
		RoleService: roleService,
	}
}

// GetAllRoles retrieves all roles from the database with search and pagination.
func (c *RoleController) GetAllRoles(ctx *fiber.Ctx) error {
	search := ctx.Query("search", "")                 // Get search term, default to empty string
	page, err := strconv.Atoi(ctx.Query("page", "1")) // Get page number, default to 1
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(ctx.Query("limit", "10")) // Get limit per page, default to 10
	if err != nil || limit < 1 {
		limit = 10
	}

	roles, totalPages, totalItems, err := c.RoleService.GetAllRoles(search, page, limit)
	if err != nil {
		log.Printf("Error fetching all roles: %v", err)
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to retrieve roles",
		})
	}

	return ctx.Status(http.StatusOK).JSON(fiber.Map{
		"success":     true,
		"message":     "Roles retrieved successfully",
		"data":        roles,
		"currentPage": page,
		"totalPages":  totalPages,
		"totalItems":  totalItems,
	})
}

// GetRoleByID retrieves a single role by their ID.
func (c *RoleController) GetRoleByID(ctx *fiber.Ctx) error {
	id := ctx.Params("id") // Get role ID from URL parameters

	role, err := c.RoleService.GetRoleByID(id)
	if err != nil {
		log.Printf("Error fetching role by ID %s: %v", id, err)
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to retrieve role",
		})
	}

	if role == nil {
		return ctx.Status(http.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Role not found",
		})
	}

	return ctx.Status(http.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Role retrieved successfully",
		"data":    role,
	})
}

// CreateRoleRequest represents the expected structure for creating a new role.
type CreateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CreateRole creates a new role in the database.
func (c *RoleController) CreateRole(ctx *fiber.Ctx) error {
	req := new(CreateRoleRequest)
	if err := ctx.BodyParser(req); err != nil {
		log.Printf("Error parsing create role request body: %v", err)
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
		})
	}

	// Check if role with the same name already exists
	existingRole, err := c.RoleService.GetRoleByName(req.Name)
	if err != nil {
		log.Printf("Error checking for existing role name %s: %v", req.Name, err)
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Internal server error",
		})
	}
	if existingRole != nil {
		return ctx.Status(http.StatusConflict).JSON(fiber.Map{
			"success": false,
			"message": "Role with this name already exists",
		})
	}

	newRole := models.NewRole(req.Name, req.Description)

	err = c.RoleService.CreateRole(newRole)
	if err != nil {
		log.Printf("Error creating role %s: %v", req.Name, err)
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to create role",
		})
	}

	return ctx.Status(http.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Role created successfully",
		"data":    newRole,
	})
}

// UpdateRoleRequest represents the expected structure for updating an existing role.
type UpdateRoleRequest struct {
	Name        *string `json:"name"` // Use pointer to differentiate between zero value and not provided
	Description *string `json:"description"`
}

// UpdateRole updates an existing role's information.
func (c *RoleController) UpdateRole(ctx *fiber.Ctx) error {
	id := ctx.Params("id") // Get role ID from URL parameters

	existingRole, err := c.RoleService.GetRoleByID(id)
	if err != nil {
		log.Printf("Error fetching existing role for update %s: %v", id, err)
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to retrieve role for update",
		})
	}
	if existingRole == nil {
		return ctx.Status(http.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "Role not found for update",
		})
	}

	req := new(UpdateRoleRequest)
	if err := ctx.BodyParser(req); err != nil {
		log.Printf("Error parsing update role request body for ID %s: %v", id, err)
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
		})
	}

	// Apply updates only if provided in the request
	if req.Name != nil {
		// Check for name conflict if name is being updated
		if *req.Name != existingRole.Name {
			conflictRole, err := c.RoleService.GetRoleByName(*req.Name)
			if err != nil {
				log.Printf("Error checking for role name conflict %s: %v", *req.Name, err)
				return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"message": "Internal server error",
				})
			}
			if conflictRole != nil {
				return ctx.Status(http.StatusConflict).JSON(fiber.Map{
					"success": false,
					"message": "Role with this name already exists",
				})
			}
		}
		existingRole.Name = *req.Name
	}
	if req.Description != nil {
		existingRole.Description = *req.Description
	}

	err = c.RoleService.UpdateRole(existingRole)
	if err != nil {
		log.Printf("Error updating role %s: %v", id, err)
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to update role",
		})
	}

	return ctx.Status(http.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Role updated successfully",
		"data":    existingRole,
	})
}

// DeleteRole deletes a role by their ID.
func (c *RoleController) DeleteRole(ctx *fiber.Ctx) error {
	id := ctx.Params("id") // Get role ID from URL parameters

	err := c.RoleService.DeleteRole(id)
	if err != nil {
		log.Printf("Error deleting role by ID %s: %v", id, err)
		// Check for specific error types if needed, e.g., "role not found"
		if err.Error() == fmt.Sprintf("role with ID %s not found for deletion", id) {
			return ctx.Status(http.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Role not found",
			})
		}
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to delete role",
		})
	}

	return ctx.Status(http.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Role deleted successfully",
	})
}
