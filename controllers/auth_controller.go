package controllers

import (
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt" // For password comparison

	// Import models package for User struct
	"github.com/anpsniper/anpbayu-be/services" // Import services package for UserService
)

// AuthController handles authentication-related requests.
type AuthController struct {
	UserService *services.UserService // UserService dependency
}

// NewAuthController creates and returns a new AuthController instance.
func NewAuthController(userService *services.UserService) *AuthController {
	return &AuthController{
		UserService: userService,
	}
}

// LoginRequest represents the expected structure of the login request body.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Login handles user login requests.
// It validates credentials and returns a success/failure response.
func (c *AuthController) Login(ctx *fiber.Ctx) error {
	// 1. Parse the request body
	req := new(LoginRequest)
	if err := ctx.BodyParser(req); err != nil {
		log.Printf("Error parsing login request body: %v", err)
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
		})
	}

	// Log received credentials (for debugging, remove in production)
	log.Printf("Login attempt for username: %s", req.Username)

	// 2. Fetch the user from the database by username (or email, depending on your login strategy)
	user, err := c.UserService.GetUserByEmail(req.Username) // Assuming username is email
	if err != nil {
		log.Printf("Error fetching user by email %s: %v", req.Username, err)
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Internal server error during user retrieval",
		})
	}
	if user == nil {
		log.Printf("Login failed: User with email %s not found.", req.Username)
		return ctx.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Invalid credentials", // Generic message for security
		})
	}

	// 3. Compare the provided password with the hashed password from the database
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		// If passwords do not match, bcrypt.CompareHashAndPassword returns an error
		log.Printf("Login failed for user %s: Invalid password.", req.Username)
		return ctx.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"message": "Invalid credentials", // Generic message for security
		})
	}

	// 4. If credentials are valid, return a success response
	// In a real application, you would generate a JWT token here and return it.
	// For NextAuth, the 'user' object returned here is used to create the session.
	// Ensure you DO NOT send the password_hash back to the frontend.
	log.Printf("User %s logged in successfully.", user.Email)
	return ctx.Status(http.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Login successful",
		"user": fiber.Map{ // Return user details (excluding password hash)
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role.Name, // Include role name
		},
		// "token": "your_generated_jwt_token_here", // Placeholder for JWT
	})
}
