package main

import (
	"log"
	"net/http"
	"time" // Still needed for time.Now() in health check

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

	"github.com/anpsniper/anpbayu-be/config"   // Import your config package
	"github.com/anpsniper/anpbayu-be/database" // Import your database package
	"github.com/anpsniper/anpbayu-be/routes"   // Import your routes package
)

func main() {
	// Load application configurations from environment variables or .env file
	// This should be the first thing you do in main.
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load application configuration: %v", err)
	}

	// Initialize database connection
	if err := database.InitDatabase(&config.AppConfig); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	// Ensure database connection is closed when the application exits
	defer database.CloseDatabase()

	// Initialize Fiber app
	app := fiber.New()

	// Configure CORS middleware using configuration from config package
	app.Use(cors.New(cors.Config{
		AllowOrigins:     config.AppConfig.FrontendOrigin, // Use configured frontend origin
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	}))

	// --- Health Check Endpoint ---
	// This endpoint can remain in main.go or be moved to a dedicated controller
	// if you prefer to keep main.go strictly for setup.
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(http.StatusOK).JSON(fiber.Map{
			"status":  "healthy",
			"message": "GoFiber backend is running!",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	// Setup all API routes by calling the function from the routes package
	routes.SetupAPIRoutes(app)

	// Start the server using the configured port
	log.Fatal(app.Listen(":" + config.AppConfig.AppPort))
}
