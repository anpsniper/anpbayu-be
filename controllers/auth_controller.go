package controllers

import (
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"

	"github.com/anpsniper/anpbayu-be/config"
)

// LoginRequest represents the expected payload for the login endpoint
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents the response for the login endpoint
type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Token   string `json:"token,omitempty"` // Placeholder for a real token
}

// AuthController holds dependencies for authentication-related operations.
type AuthController struct {
	// Add any dependencies here, e.g., a service for user management, database connection
}

// NewAuthController creates and returns a new instance of AuthController.
func NewAuthController() *AuthController {
	return &AuthController{}
}

// Login handles user authentication.
// It expects an email and password in the request body, validates them
// against the configured credentials, and returns a success/failure response.
func (ac *AuthController) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Error parsing login request: %v", err)
		return c.Status(http.StatusBadRequest).JSON(LoginResponse{
			Success: false,
			Message: "Invalid request body",
		})
	}

	// Authenticate using configured email and password from the global AppConfig
	if req.Email == config.AppConfig.AuthEmail && req.Password == config.AppConfig.AuthPassword {
		log.Printf("User '%s' logged in successfully.", req.Email)
		// In a real application, you would generate a JWT here using config.AppConfig.JWTSecret
		return c.Status(http.StatusOK).JSON(LoginResponse{
			Success: true,
			Message: "Login successful",
			Token:   "dummy-jwt-token-for-user", // Placeholder for a real JWT token
		})
	} else {
		log.Printf("Login failed for user '%s'. Invalid credentials.", req.Email)
		return c.Status(http.StatusUnauthorized).JSON(LoginResponse{
			Success: false,
			Message: "Invalid credentials",
		})
	}
}

// You can add other authentication-related methods here, e.g., Register, Logout, RefreshToken
