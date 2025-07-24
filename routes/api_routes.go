package routes

import (
	"github.com/anpsniper/anpbayu-be/controllers" // Import controllers package
	"github.com/anpsniper/anpbayu-be/services"    // Import services package
	"github.com/gofiber/fiber/v2"
)

// SetupAPIRoutes sets up all API routes for the Fiber application.
func SetupAPIRoutes(app *fiber.App) {
	// Initialize services
	userService := services.NewUserService()
	roleService := services.NewRoleService() // Initialize RoleService

	// Initialize controllers with their dependencies.
	authController := controllers.NewAuthController(userService)
	userController := controllers.NewUserController(userService)
	roleController := controllers.NewRoleController(roleService) // Initialize RoleController

	// Group API routes under /api prefix.
	api := app.Group("/api")

	// --- Authentication Routes ---
	api.Post("/login", authController.Login)

	// --- Protected Routes (Example) ---
	api.Get("/dashboard", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Welcome to the protected dashboard!",
			"data":    "This data is from GoFiber backend.",
		})
	})

	// --- User Management Routes ---
	api.Get("/users", userController.GetAllUsers)       // GET /api/users
	api.Get("/users/:id", userController.GetUserByID)   // GET /api/users/:id
	api.Post("/users", userController.CreateUser)       // POST /api/users
	api.Put("/users/:id", userController.UpdateUser)    // PUT /api/users/:id
	api.Delete("/users/:id", userController.DeleteUser) // DELETE /api/users/:id

	// --- Role Management Routes ---
	api.Get("/roles", roleController.GetAllRoles)       // GET /api/roles (with search, pagination)
	api.Get("/roles/:id", roleController.GetRoleByID)   // GET /api/roles/:id
	api.Post("/roles", roleController.CreateRole)       // POST /api/roles
	api.Put("/roles/:id", roleController.UpdateRole)    // PUT /api/roles/:id
	api.Delete("/roles/:id", roleController.DeleteRole) // DELETE /api/roles/:id
}
