package controllers

import (
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"

	"github.com/anpsniper/anpbayu-be/models"   // Import your models package
	"github.com/anpsniper/anpbayu-be/services" // Import your services package
)

// UserController holds dependencies for user-related operations.
type UserController struct {
	UserService services.UserService // Dependency on the UserService interface
}

// NewUserController creates and returns a new instance of UserController.
func NewUserController() *UserController {
	return &UserController{
		UserService: services.NewUserService(), // Instantiate the user service
	}
}

// GetProfile handles fetching a user's profile.
// This is a protected endpoint that requires the AuthRequired middleware
// to ensure only logged-in users can access their profile data.
func (uc *UserController) GetProfile(c *fiber.Ctx) error {
	// 1. Extract the authenticated user's email from Fiber's locals context.
	// This value is set by the AuthRequired middleware after successful JWT validation.
	userEmail, ok := c.Locals("userEmail").(string)
	if !ok || userEmail == "" {
		log.Println("User email not found in context (AuthRequired middleware likely failed or not applied).")
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized: User context missing",
		})
	}

	log.Printf("Fetching profile for user: %s", userEmail)

	// 2. Retrieve the user's actual profile data from the database using the UserService.
	user, err := uc.UserService.GetUserByEmail(userEmail)
	if err != nil {
		log.Printf("Error fetching user profile for email '%s': %v", userEmail, err)
		// Return a generic error to the client
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to retrieve user profile",
		})
	}

	// 3. Construct the UserResponse to send back to the client.
	// This typically excludes sensitive fields like PasswordHash.
	userResponse := models.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name, // This will be empty if not set in DB
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "User profile fetched successfully",
		"profile": userResponse,
	})
}
