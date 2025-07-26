package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"  // Standard Fiber import path
	"github.com/golang-jwt/jwt/v5" // Using v5 for JWT

	"github.com/anpsniper/anpbayu-be/config" // Import your config package
)

// GenerateJWT creates a new JWT token for the given user ID and roles.
// Make sure this function exists and is exported (starts with a capital G).
func GenerateJWT(userID string, email string, roles []string) (string, error) {
	// Define your claims
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"roles":   roles,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
		"iat":     time.Now().Unix(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with your secret key from the application configuration
	// This ensures consistency with the JWT middleware in main.go
	t, err := token.SignedString([]byte(config.AppConfig.JWTSecret)) // NOW USING config.AppConfig.JWTSecret
	if err != nil {
		return "", err
	}
	return t, nil
}

// GetUserIDFromJWT extracts the user ID from the JWT claims in the Fiber context.
// It assumes the jwtware.New middleware has already run and populated c.Locals("user").
func GetUserIDFromJWT(c *fiber.Ctx) (string, bool) { // Changed return type to string for UUID
	// The jwtware.New middleware puts the *jwt.Token object into c.Locals("user")
	user := c.Locals("user")
	if user == nil {
		return "", false // Token not found in locals, likely Auth middleware not run or failed
	}

	// Assert the type to *jwt.Token
	token, ok := user.(*jwt.Token)
	if !ok {
		return "", false // Not a *jwt.Token
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", false // Claims not of type jwt.MapClaims
	}

	// Extract user_id. It's a string (UUID) in your models.
	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", false // user_id not found or not string
	}
	return userID, true
}

// GetUserRolesFromJWT extracts the user roles from the JWT claims in the Fiber context.
// It assumes the jwtware.New middleware has already run and populated c.Locals("user").
// This is adapted for roles stored as a slice in the JWT.
func GetUserRolesFromJWT(c *fiber.Ctx) ([]string, bool) {
	user := c.Locals("user")
	if user == nil {
		return nil, false // Token not found in locals
	}

	token, ok := user.(*jwt.Token)
	if !ok {
		return nil, false // Not a *jwt.Token
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, false // Claims not of type jwt.MapClaims
	}

	// Extract roles. It's expected to be stored as []string in the JWT.
	// We need to handle the case where it might be []interface{} if Marshalled directly from a slice of strings.
	rolesInterface, ok := claims["roles"].([]interface{})
	if ok {
		var roles []string
		for _, role := range rolesInterface {
			if r, isString := role.(string); isString {
				roles = append(roles, r)
			}
		}
		return roles, true
	}

	// Fallback: If "roles" claim is a single string (less common, but good to handle)
	if singleRole, isString := claims["roles"].(string); isString {
		return []string{singleRole}, true
	}

	return nil, false // roles not found or not a slice/string
}

// JwtError is a custom error handler for the Fiber JWT middleware.
func JwtError(c *fiber.Ctx, err error) error {
	if err.Error() == "Missing or malformed JWT" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Missing or malformed JWT"})
	}
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "error", "message": "Invalid or expired JWT"})
}

// HasRole is a Fiber middleware that checks if the authenticated user has
// at least one of the required roles.
// It should be used AFTER the main JWT authentication middleware.
func HasRole(requiredRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRoles, ok := GetUserRolesFromJWT(c)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Forbidden: User roles not found in token or token invalid. (Ensure JWT middleware runs first)",
			})
		}

		for _, requiredRole := range requiredRoles {
			for _, userRole := range userRoles {
				if userRole == requiredRole {
					return c.Next() // User has at least one required role, proceed
				}
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Forbidden: Insufficient role permissions",
		})
	}
}
