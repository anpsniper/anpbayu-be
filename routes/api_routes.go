package routes

import (
	"github.com/anpsniper/anpbayu-be/controllers" // Import your controllers package
	"github.com/anpsniper/anpbayu-be/middleware"  // Import your middleware package
	"github.com/gofiber/fiber/v2"
)

// SetupAPIRoutes configures all the API endpoints for the application.
// It takes a Fiber app instance and attaches the routes to it.
func SetupAPIRoutes(app *fiber.App) {
	// Initialize controllers
	authController := controllers.NewAuthController()
	userController := controllers.NewUserController()
	// Add other controllers as they are created (e.g., productController := controllers.NewProductController())

	// Define a group for API routes (e.g., /api/v1)
	api := app.Group("/api") // You can add /v1 here if you want versioning: app.Group("/api/v1")

	// --- Public Routes ---
	// Login route: accessible without authentication
	api.Post("/login", authController.Login)

	// --- Protected Routes ---
	// Create a sub-group for routes that require authentication
	protected := api.Group("/", middleware.AuthRequired()) // Apply AuthRequired middleware to this group

	// Dashboard route: requires authentication
	protected.Get("/dashboard", func(c *fiber.Ctx) error {
		// This handler is now protected by AuthRequired middleware.
		// You can access user data stored by the middleware, e.g., c.Locals("userEmail")
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message":                "Welcome to the protected dashboard! This data came from the GoFiber backend (via protected route).",
			"data":                   "Some sensitive data only for logged-in users.",
			"authenticatedUserEmail": c.Locals("userEmail"), // Example: Accessing user email from middleware
		})
	})

	// User profile route: requires authentication
	protected.Get("/profile", userController.GetProfile)

	// Example: Product routes (if you add a ProductController later)
	// productController := controllers.NewProductController()
	// protected.Get("/products", productController.GetAllProducts)
	// protected.Post("/products", productController.CreateProduct)
	// protected.Get("/products/:id", productController.GetProductByID)
	// protected.Put("/products/:id", productController.UpdateProduct)
	// protected.Delete("/products/:id", productController.DeleteProduct)
}
