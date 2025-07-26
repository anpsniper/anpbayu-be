package controllers

import (
	"log"
	"net/http"

	"github.com/anpsniper/anpbayu-be/middleware" // Assuming JWT generation is here
	"github.com/anpsniper/anpbayu-be/services"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

// AuthController handles user authentication and session management.
type AuthController struct {
	UserService services.UserServiceInterface
}

// NewAuthController creates and returns a new AuthController instance.
func NewAuthController(userService services.UserServiceInterface) *AuthController {
	return &AuthController{
		UserService: userService,
	}
}

// LoginRequest defines the structure for login requests.
type LoginRequest struct {
	Email    string `json:"email"` // Changed to email as per your service layer
	Password string `json:"password"`
}

// Login handles user authentication and generates a JWT.
func (c *AuthController) Login(ctx *fiber.Ctx) error {
	req := new(LoginRequest)
	if err := ctx.BodyParser(req); err != nil {
		log.Printf("Error parsing login request body: %v", err)
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Invalid request body"})
	}

	user, err := c.UserService.GetUserByEmail(req.Email) // Fetch by email
	if err != nil {
		log.Printf("Error getting user by email %s: %v", req.Email, err)
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Internal server error"})
	}
	if user == nil {
		log.Printf("Login failed: User with email %s not found.", req.Email)
		return ctx.Status(http.StatusUnauthorized).JSON(fiber.Map{"status": "error", "message": "Invalid credentials"})
	}

	// Compare the provided password with the hashed password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		log.Printf("Login failed for user %s: Invalid password.", req.Email)
		return ctx.Status(http.StatusUnauthorized).JSON(fiber.Map{"status": "error", "message": "Invalid credentials"})
	}

	// Generate JWT token
	// Assuming GenerateJWT takes userID and a slice of roles (strings)
	token, err := middleware.GenerateJWT(user.ID, []string{user.RoleName}) // Pass user.RoleName as a slice
	if err != nil {
		log.Printf("Error generating JWT: %v", err)
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Failed to generate token"})
	}

	// NEW DEBUG LOG: Confirming the role name before sending to frontend
	log.Printf("DEBUG: User %s (ID: %s) has RoleName: '%s'", user.Email, user.ID, user.RoleName)

	// NEW: Log the login event and get the log ID
	logID, err := c.UserService.CreateUserLoginLog(user.ID)
	if err != nil {
		log.Printf("Warning: Failed to create login log for user %s: %v", user.ID, err)
		// Do not return error to client, as login itself was successful
	}

	return ctx.Status(http.StatusOK).JSON(fiber.Map{
		"status":            "success",
		"message":           "Login successful",
		"token":             token,
		"user_id":           user.ID,
		"username":          user.Username, // Ensure this is populated if you want it in session.user.name
		"email":             user.Email,
		"role_id":           user.RoleID,
		"roles":             []string{user.RoleName}, // Return roles as array for frontend
		"last_login_log_id": logID,                   // Return log ID to frontend for logout tracking
	})
}

// LogoutRequest defines the structure for logout requests.
type LogoutRequest struct {
	LastLoginLogID int `json:"last_login_log_id"` // The ID of the login log to update
}

// Logout handles user logout and updates the logout timestamp in the log.
// This endpoint assumes the frontend sends the last_login_log_id obtained during login.
func (c *AuthController) Logout(ctx *fiber.Ctx) error {
	req := new(LogoutRequest)
	if err := ctx.BodyParser(req); err != nil {
		log.Printf("Error parsing logout request body: %v", err)
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Invalid request body"})
	}

	// NEW DEBUG LOG: Log the received LastLoginLogID
	log.Printf("DEBUG: Received logout request for LastLoginLogID: %d", req.LastLoginLogID)

	if req.LastLoginLogID == 0 {
		return ctx.Status(http.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "last_login_log_id is required for logout logging"})
	}

	err := c.UserService.UpdateUserLogoutLog(req.LastLoginLogID)
	if err != nil {
		log.Printf("Warning: Failed to update logout log for ID %d: %v", req.LastLoginLogID, err)
		// Log the error but still return success to the client for logout
		return ctx.Status(http.StatusOK).JSON(fiber.Map{"status": "success", "message": "Logout successful, but log update failed."})
	}

	return ctx.Status(http.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Logout successful and log updated.",
	})
}
