package middleware

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time" // For token expiration checks

	"github.com/anpsniper/anpbayu-be/config" // Import your config package for JWT secret
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5" // JWT library
)

// Claims defines the structure of the JWT claims.
// You can add more fields here like 'UserID', 'Roles', etc.
type Claims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// AuthRequired is a Fiber middleware that checks for JWT authentication.
// It parses and validates a JWT token from the Authorization header.
func AuthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		log.Println("AuthRequired middleware triggered.")

		// 1. Get the Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			log.Println("Authorization header missing.")
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"message": "Unauthorized: Missing token",
			})
		}

		// 2. Check if it's a Bearer token and extract it
		if !strings.HasPrefix(authHeader, "Bearer ") {
			log.Println("Authorization header malformed: not a Bearer token.")
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"message": "Unauthorized: Invalid token format",
			})
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// 3. Parse and validate the token
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			// Return the JWT secret from your config
			return []byte(config.AppConfig.JWTSecret), nil
		})

		if err != nil {
			log.Printf("Token parsing or validation failed: %v", err)
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"message": fmt.Sprintf("Unauthorized: Invalid token (%v)", err),
			})
		}

		// 4. Check if the token is valid
		if !token.Valid {
			log.Println("Invalid token.")
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"message": "Unauthorized: Invalid token",
			})
		}

		// 5. Extract claims and store them in Fiber's Locals for subsequent handlers
		claims, ok := token.Claims.(*Claims)
		if !ok {
			log.Println("Could not extract claims from token.")
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"message": "Unauthorized: Invalid token claims",
			})
		}

		// Optional: Check token expiration (jwt.RegisteredClaims already handles this, but explicit check can be useful)
		if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
			log.Println("Token expired.")
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"message": "Unauthorized: Token expired",
			})
		}

		// Store user email (or other relevant claims) in Fiber's locals context
		// This makes user data accessible to subsequent handlers in the route chain.
		c.Locals("userEmail", claims.Email)
		log.Printf("Authentication successful for user: %s. Proceeding to next handler.", claims.Email)

		return c.Next() // Proceed to the next handler in the chain
	}
}
