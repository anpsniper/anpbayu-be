package utils

import (
	"log"
	"time"
)

// ExampleHelperFunction is a placeholder for a general utility function.
// You can replace this with functions for things like:
// - String manipulation
// - Data formatting
// - Common validation logic (if not handled by specific libraries like Zod on frontend)
// - Error handling utilities
func ExampleHelperFunction(message string) {
	log.Printf("Helper function called at %s with message: %s", time.Now().Format(time.RFC3339), message)
}

// GenerateRandomString can be another helper function for generating random strings,
// useful for IDs, temporary passwords, etc.
// func GenerateRandomString(length int) (string, error) {
// 	// Implementation for generating a random string
// 	return "", nil
// }

// ValidateEmailFormat could be a server-side email validation if not using a library
// func ValidateEmailFormat(email string) bool {
// 	// Simple regex check or use a library
// 	return true // Placeholder
// }
