package controllers

import (
	"log"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5" // Import JWT library

	"github.com/anpsniper/anpbayu-be/config" // Import your config package
	"github.com/anpsniper/anpbayu-be/middleware"
	"github.com/anpsniper/anpbayu-be/services" // Import your services package
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
	Token   string `json:"token,omitempty"` // JWT token will be included here
}

// AuthController holds dependencies for authentication-related operations.
type AuthController struct {
	UserService services.UserService // Dependency on the UserService interface
}

// NewAuthController creates and returns a new instance of AuthController.
// It initializes with a concrete implementation of UserService.
func NewAuthController() *AuthController {
	return &AuthController{
		UserService: services.NewUserService(), // Instantiate the user service
	}
}

// Login handles user authentication.
// It expects an email and password in the request body, validates them
// against the database, and generates a JWT upon successful authentication.
func (ac *AuthController) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Error parsing login request: %v", err)
		return c.Status(http.StatusBadRequest).JSON(LoginResponse{
			Success: false,
			Message: "Invalid request body",
		})
	}

	// 1. Retrieve user by email from the database
	user, err := ac.UserService.GetUserByEmail(req.Email)
	if err != nil {
		log.Printf("Login failed for email '%s': %v", req.Email, err)
		// Return a generic "Invalid credentials" to avoid leaking user existence
		return c.Status(http.StatusUnauthorized).JSON(LoginResponse{
			Success: false,
			Message: "Invalid credentials",
		})
	}

	// 2. Compare the provided password with the hashed password from the database
	if !services.ComparePasswords([]byte(user.PasswordHash), []byte(req.Password)) {
		log.Printf("Login failed for email '%s': Incorrect password", req.Email)
		return c.Status(http.StatusUnauthorized).JSON(LoginResponse{
			Success: false,
			Message: "Invalid credentials",
		})
	}

	// 3. Generate JWT token
	// Define claims for the token
	claims := middleware.Claims{
		Email: req.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)), // Token expires in 24 hours
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	// Create a new token with the specified claims and signing method
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key from config
	tokenString, err := token.SignedString([]byte(config.AppConfig.JWTSecret))
	if err != nil {
		log.Printf("Error signing JWT token: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(LoginResponse{
			Success: false,
			Message: "Failed to generate token",
		})
	}

	log.Printf("User '%s' logged in successfully. Token generated.", req.Email)
	return c.Status(http.StatusOK).JSON(LoginResponse{
		Success: true,
		Message: "Login successful",
		Token:   tokenString, // Return the generated JWT token
	})
}

// You can add other authentication-related methods here, e.g., Register, Logout, RefreshToken
