package main

import (
	// Added for sql.ErrNoRows
	"log"
	"net/http"
	"time"

	jwtware "github.com/gofiber/contrib/jwt" // Use the base module path for jwtware (no /v5 here)
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

	// Import the JWT library itself
	"github.com/anpsniper/anpbayu-be/config"      // Your config package
	"github.com/anpsniper/anpbayu-be/controllers" // Import controllers package
	"github.com/anpsniper/anpbayu-be/database"    // Your database package
	"github.com/anpsniper/anpbayu-be/models"      // Your models package (User, Role, etc.)
	"github.com/anpsniper/anpbayu-be/routes"      // Your routes package
	"github.com/anpsniper/anpbayu-be/services"    // Import services package
)

func main() {
	// 1. Load application configurations from environment variables or .env file
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load application configuration: %v", err)
	}

	// 2. Initialize database connection (using database/sql as per your db.go)
	if err := database.InitDatabase(&config.AppConfig); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	// Ensure database connection is closed when the application exits
	defer database.CloseDatabase()

	// 3. Seed roles and example user
	// These functions (in models/modelseed.go) are now compatible with database/sql
	// and expect database.DB to be *sql.DB.
	log.Println("Seeding roles...")
	if err := models.SeedRoles(); err != nil {
		log.Fatalf("Failed to seed roles: %v", err)
	}
	log.Println("Roles seeded successfully.")

	log.Println("Seeding example user...")
	if err := models.SeedExampleUser(); err != nil {
		log.Fatalf("Failed to seed example user: %v", err)
	}
	log.Println("Example user seeded successfully.")

	// 4. Initialize Fiber app
	app := fiber.New()

	// 5. Configure CORS middleware using configuration from config package
	app.Use(cors.New(cors.Config{
		AllowOrigins:     config.AppConfig.FrontendOrigin,
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	}))

	// 6. Health Check Endpoint (publicly accessible)
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(http.StatusOK).JSON(fiber.Map{
			"status":  "healthy",
			"message": "GoFiber backend is running!",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	// Initialize UserService and AuthController
	userService := services.NewUserService()
	authController := controllers.NewAuthController(userService)

	// 7. Authentication Login Route (publicly accessible, handled by AuthController)
	// This replaces the manual login handler that was here.
	app.Post("/login", authController.Login) // Frontend should hit this endpoint directly

	// 8. JWT Middleware (Applies to all routes defined AFTER this point)
	// This middleware will protect all subsequent routes unless explicitly overridden.
	app.Use(jwtware.New(jwtware.Config{
		SigningKey: jwtware.SigningKey{Key: []byte(config.AppConfig.JWTSecret)},
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			if err.Error() == "Missing or malformed JWT" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing or malformed JWT"})
			}
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid or expired JWT"})
		},
	}))

	// 9. Setup all API routes (these will now be protected by the JWT middleware,
	// and some will have additional role-based checks via `middleware.HasRole`).
	// The /api/auth/logout route will also be handled by the authController within SetupAPIRoutes.
	routes.SetupAPIRoutes(app)

	// 10. Start the Fiber server
	log.Printf("GoFiber API server starting on port %s...", config.AppConfig.AppPort)
	log.Fatal(app.Listen(":" + config.AppConfig.AppPort))
}
