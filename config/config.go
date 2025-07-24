// This is what your local config/config.go MUST contain for the Config struct
package config

import (
	"log"
	"os"

	"github.com/joho/godotenv" // For loading .env files
)

// Config holds all application-wide configurations.
type Config struct {
	AppPort        string
	FrontendOrigin string
	AuthEmail      string
	AuthPassword   string
	JWTSecret      string
	DBURL          string // <--- THIS LINE IS CRUCIAL AND MUST BE PRESENT
}

// AppConfig is a global instance of the Config struct.
// It will hold the loaded configuration values.
var AppConfig Config

// LoadConfig reads configuration from environment variables or .env file.
// It should be called once at the start of the application.
func LoadConfig() error {
	// Attempt to load .env file.
	// If it fails, log a message but continue, as variables might be set directly in the environment.
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found or error loading .env file, assuming environment variables are set externally.")
	}

	// Populate the AppConfig struct from environment variables.
	// Provide default values if environment variables are not set.

	AppConfig.AppPort = os.Getenv("APP_PORT")
	if AppConfig.AppPort == "" {
		AppConfig.AppPort = "8080" // Default port
		log.Printf("APP_PORT not set, defaulting to %s", AppConfig.AppPort)
	}

	AppConfig.FrontendOrigin = os.Getenv("FRONTEND_ORIGIN")
	if AppConfig.FrontendOrigin == "" {
		AppConfig.FrontendOrigin = "http://localhost:3000" // Default frontend origin for CORS
		log.Printf("FRONTEND_ORIGIN not set, defaulting to %s", AppConfig.FrontendOrigin)
	}

	AppConfig.AuthEmail = os.Getenv("AUTH_EMAIL")
	if AppConfig.AuthEmail == "" {
		AppConfig.AuthEmail = "user@example.com" // Default authentication email
		log.Printf("AUTH_EMAIL not set, defaulting to %s", AppConfig.AuthEmail)
	}

	AppConfig.AuthPassword = os.Getenv("AUTH_PASSWORD")
	if AppConfig.AuthPassword == "" {
		AppConfig.AuthPassword = "password" // Default authentication password
		log.Printf("AUTH_PASSWORD not set, defaulting to default password")
	}

	AppConfig.JWTSecret = os.Getenv("JWT_SECRET")
	if AppConfig.JWTSecret == "" {
		AppConfig.JWTSecret = "supersecretjwtkey" // Default JWT secret
		log.Printf("JWT_SECRET not set, defaulting to default secret")
	}

	// Load database URL
	AppConfig.DBURL = os.Getenv("DB_URL")
	if AppConfig.DBURL == "" {
		// Provide a default DB_URL for local development if not set
		AppConfig.DBURL = "host=localhost user=postgres password=password dbname=mydatabase port=5432 sslmode=disable"
		log.Printf("DB_URL not set, defaulting to: %s", AppConfig.DBURL)
	}

	log.Println("Configuration loaded successfully.")
	return nil
}
