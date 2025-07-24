package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// User represents a simple user structure for authentication
type User struct {
	Username string `json:"username"` // Note: Frontend now uses 'email', backend still expects 'username' for this example
	Password string `json:"password"`
}

// LoginRequest represents the expected payload for the login endpoint
type LoginRequest struct {
	Email    string `json:"email"` // Changed to 'email' to match frontend
	Password string `json:"password"`
}

// LoginResponse represents the response for the login endpoint
type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Token   string `json:"token,omitempty"` // Placeholder for a real token
}

func main() {
	// Initialize Fiber app
	app := fiber.New()

	// Configure CORS middleware
	// This allows the Next.js frontend (running on a different port) to communicate with the backend.
	// In a production environment, you would restrict AllowedOrigins to your specific frontend domain.
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000", // Allow requests from Next.js development server
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	}))

	// --- API Endpoints ---

	// Health Check Endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(http.StatusOK).JSON(fiber.Map{
			"status":  "healthy",
			"message": "GoFiber backend is running!",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	// Login Endpoint
	// This endpoint handles user authentication.
	// For this example, it checks against a hardcoded email and password.
	app.Post("/api/login", func(c *fiber.Ctx) error {
		var req LoginRequest
		if err := c.BodyParser(&req); err != nil {
			log.Printf("Error parsing login request: %v", err)
			return c.Status(http.StatusBadRequest).JSON(LoginResponse{
				Success: false,
				Message: "Invalid request body",
			})
		}

		// Simulate user authentication
		// In a real application, you would query a database and hash passwords.
		// The hardcoded user is now 'user@example.com' with password 'password'
		if req.Email == "user@example.com" && req.Password == "password" {
			log.Printf("User '%s' logged in successfully.", req.Email)
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
	})

	// Protected Dashboard Endpoint
	// This endpoint is meant to be accessed only by authenticated users.
	// For this simple example, it doesn't actually check for a token,
	// but in a real app, you'd add middleware here to validate JWTs.
	app.Get("/api/dashboard", func(c *fiber.Ctx) error {
		// In a real application, you would typically have a middleware
		// that validates an Authorization header (e.g., JWT token)
		// before reaching this handler.
		log.Println("Accessing protected dashboard endpoint.")
		return c.Status(http.StatusOK).JSON(fiber.Map{
			"message": "Welcome to the protected dashboard! This data came from the GoFiber backend.",
			"data":    "Some sensitive data only for logged-in users.",
		})
	})

	// Start the server
	log.Fatal(app.Listen(":8080")) // Listen on port 8080
}
