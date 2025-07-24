package controllers

import (
	"fmt"
	"log"
	"net/http"

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

// GetAllUsers retrieves all users from the database.
// This is an example and might need pagination/filtering for large datasets.
func (c *UserController) GetAllUsers(ctx *fiber.Ctx) error {
	// In a real application, you'd implement a method in UserService
	// to get all users. For now, let's return a placeholder.
	// You would typically query the database for all users here.
	log.Println("GetAllUsers endpoint hit.")
	return ctx.Status(http.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "This endpoint would return all users.",
		"data":    []string{"user1", "user2", "user3"}, // Placeholder data
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

	// Fetch existing user to get current values (especially password hash)
	existingUser, err := c.UserService.GetUserByID(id)
	if err != nil {
		log.Printf("Error fetching existing user for update %s: %v", id, err)
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to retrieve user for update",
		})
	}
	if existingUser == nil {
		return ctx.Status(http.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"message": "User not found for update",
		})
	}

	req := new(UpdateUserRequest)
	if err := ctx.BodyParser(req); err != nil {
		log.Printf("Error parsing update user request body for ID %s: %v", id, err)
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
		})
	}

	// Apply updates only if provided in the request
	if req.Username != nil {
		existingUser.Username = *req.Username
	}
	if req.Email != nil {
		existingUser.Email = *req.Email
	}
	if req.RoleID != nil {
		existingUser.RoleID = *req.RoleID
	}

	err = c.UserService.UpdateUser(existingUser)
	if err != nil {
		log.Printf("Error updating user %s: %v", id, err)
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to update user",
		})
	}

	// Do not return password hash
	existingUser.Password = ""
	return ctx.Status(http.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "User updated successfully",
		"data":    existingUser,
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
