package controllers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	// For time.Now()
	"github.com/gofiber/fiber/v2" // For generating UUIDs
	"golang.org/x/crypto/bcrypt"  // For password hashing

	"github.com/anpsniper/anpbayu-be/models"   // Import models package for User struct
	"github.com/anpsniper/anpbayu-be/services" // Import services package for UserServiceInterface
)

// UserController handles user-related requests.
type UserController struct {
	UserService services.UserServiceInterface // UserService dependency (interface)
}

// NewUserController creates and returns a new UserController instance.
func NewUserController(userService services.UserServiceInterface) *UserController {
	return &UserController{
		UserService: userService,
	}
}

func (c *UserController) GetAllUsers(ctx *fiber.Ctx) error {
	log.Println("GetAllUsers endpoint hit.")

	search := ctx.Query("search", "")                 // Get search term, default to empty string
	page, err := strconv.Atoi(ctx.Query("page", "1")) // Get page number, default to 1
	if err != nil || page < 1 {
		page = 1
	}
	limit, err := strconv.Atoi(ctx.Query("limit", "10")) // Get limit per page, default to 10
	if err != nil || limit < 1 {
		limit = 10
	}

	users, totalPages, totalItems, err := c.UserService.GetAllUsers(search, page, limit)
	if err != nil {
		log.Printf("Error fetching all users: %v", err)
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to retrieve users",
		})
	}

	// --- START OF REQUIRED MAPPING ---
	// Map models.User to models.UserResponse for the client
	userResponses := make([]models.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = models.UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			RoleID:    user.RoleID,
			RoleName:  user.RoleName,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}
	}
	// --- END OF REQUIRED MAPPING ---

	return ctx.Status(http.StatusOK).JSON(fiber.Map{
		"success":     true,
		"message":     "Users retrieved successfully",
		"data":        userResponses, // <-- Now returning userResponses
		"currentPage": page,
		"totalPages":  totalPages,
		"totalItems":  totalItems,
	})
}

func (c *UserController) GetAllRoles(ctx *fiber.Ctx) error {
	log.Println("GetAllRoles endpoint hit.")

	roles, err := c.UserService.GetAllRoles()
	if err != nil {
		log.Printf("Error fetching all roles: %v", err)
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to retrieve roles due to an internal error.",
		})
	}

	// Map models.Role to models.RoleResponse for API output
	// Jika Anda hanya ingin ID dan Name, gunakan models.RoleResponse
	// Jika Anda ingin semua field dari Role, cukup gunakan `roles` langsung
	roleResponses := make([]models.LstRole, len(roles))
	for i, role := range roles {
		roleResponses[i] = models.LstRole{
			ID:   role.ID,
			Name: role.Name, // Akses role.Name, bukan role.Username
		}
	}

	return ctx.Status(http.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Roles retrieved successfully.",
		"data":    roleResponses, // Mengembalikan DTO yang dimapping
	})
}

// GetUserByID retrieves a single user by their ID.
func (c *UserController) GetUserByID(ctx *fiber.Ctx) error {
	id := ctx.Params("id") // Get user ID from URL parameters

	user, err := c.UserService.GetUserByID(id)
	if err != nil {
		log.Printf("Error fetching user by ID %s: %v", id, err)
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to retrieve user",
		})
	}

	if user == nil {
		return ctx.Status(http.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "User not found",
		})
	}

	// Do not return password hash
	user.Password = ""
	return ctx.Status(http.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "User retrieved successfully",
		"data":    user,
	})
}

// CreateUserRequest represents the expected structure for creating a new user.
type CreateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	RoleID   string `json:"role_id"` // Expecting role ID from frontend
}

// CreateUser creates a new user in the database.
func (c *UserController) CreateUser(ctx *fiber.Ctx) error {
	req := new(CreateUserRequest)
	if err := ctx.BodyParser(req); err != nil {
		log.Printf("Error parsing create user request body: %v", err)
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
		})
	}

	// Hash the password before storing it
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password for new user %s: %v", req.Email, err)
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to process password",
		})
	}

	// Create a new User model instance
	newUser := models.NewUser(req.Username, req.Email, string(hashedPassword), req.RoleID)
	// The ID, CreatedAt, UpdatedAt will be set by the service/database layer

	err = c.UserService.CreateUser(newUser)
	if err != nil {
		log.Printf("Error creating user %s: %v", req.Email, err)
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to create user",
		})
	}

	// Do not return password hash
	newUser.Password = ""
	return ctx.Status(http.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "User created successfully",
		"data":    newUser,
	})
}

// UpdateUserRequest represents the expected structure for updating an existing user.
type UpdateUserRequest struct {
	Username *string `json:"username"` // Use pointer to differentiate between zero value and not provided
	Email    *string `json:"email"`
	RoleID   *string `json:"role_id"`
	// Password updates should be handled by a separate endpoint for security
}

// UpdateUser updates an existing user's information.
func (c *UserController) UpdateUser(ctx *fiber.Ctx) error {
	id := ctx.Params("id") // Get user ID from URL parameters
	if id == "" {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "User ID is required",
		})
	}

	req := new(models.UpdateUserRequest) // Use models.UpdateUserRequest
	if err := ctx.BodyParser(req); err != nil {
		log.Printf("Error parsing update user request body for ID %s: %v", id, err)
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
		})
	}

	req.ID = id // Set the ID from the URL parameter to the request struct

	// Basic validation (ensure required fields for update are present)
	if req.Username == "" || req.Email == "" || req.RoleID == "" {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Username, email, and role ID are required for update",
		})
	}

	// Pass the request directly to the service; conditional password update is handled in service.
	err := c.UserService.UpdateUser(req)
	if err != nil {
		log.Printf("Error updating user %s: %v", id, err)
		if err.Error() == fmt.Sprintf("user with ID %s not found for update", id) {
			return ctx.Status(http.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "User not found",
				"error":   err.Error(),
			})
		}
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to update user",
			"error":   err.Error(),
		})
	}

	return ctx.Status(http.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "User updated successfully",
	})
}

// DeleteUser deletes a user by their ID.
func (c *UserController) DeleteUser(ctx *fiber.Ctx) error {
	id := ctx.Params("id") // Get user ID from URL parameters

	err := c.UserService.DeleteUser(id)
	if err != nil {
		log.Printf("Error deleting user by ID %s: %v", id, err)
		// Check for specific error types if needed, e.g., "user not found"
		if err.Error() == fmt.Sprintf("user with ID %s not found for deletion", id) {
			return ctx.Status(http.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "User not found",
			})
		}
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to delete user",
		})
	}

	return ctx.Status(http.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "User deleted successfully",
	})
}
