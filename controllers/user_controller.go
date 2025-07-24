package controllers

import (
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	// No direct config import needed here for now, as it's not used in this dummy example.
	// If you were to fetch real user data from a DB, you might need config for DB connection.
)

// UserProfile represents the structure of a user's profile data.
type UserProfile struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// UserController holds dependencies for user-related operations.
type UserController struct {
	// Add any dependencies here, e.g., a service for user data retrieval, database connection
}

// NewUserController creates and returns a new instance of UserController.
func NewUserController() *UserController {
	return &UserController{}
}

// GetProfile handles fetching a user's profile.
// This is a protected endpoint that would typically require authentication middleware
// to ensure only logged-in users can access their profile data.
func (uc *UserController) GetProfile(c *fiber.Ctx) error {
	// In a real application, you would extract the user ID from the JWT token
	// (which would be validated by a middleware before this handler is called).
	// Then, you would fetch the user's profile from a database.

	log.Println("Accessing user profile endpoint.")

	// Simulate fetching a user profile
	dummyProfile := UserProfile{
		ID:    "user-123",
		Email: "user@example.com", // This would typically come from the authenticated user's session
		Name:  "Demo User",
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "User profile fetched successfully",
		"profile": dummyProfile,
	})
}
