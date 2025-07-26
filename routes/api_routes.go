package routes

import (
	"net/http" // For http.StatusOK etc.

	"github.com/anpsniper/anpbayu-be/controllers" // Import controllers package
	"github.com/anpsniper/anpbayu-be/middleware"  // Import your custom middleware for RBAC
	"github.com/anpsniper/anpbayu-be/services"    // Import services package
	"github.com/gofiber/fiber/v2"
)

// SetupAPIRoutes sets up all API routes for the Fiber application.
// IMPORTANT: This function is called AFTER the JWT authentication middleware
// in main.go. Therefore, all routes defined here will automatically require
// a valid JWT. Role-based access control is then applied on top of that.
func SetupAPIRoutes(app *fiber.App) {
	// Initialize services
	userService := services.NewUserService()
	roleService := services.NewRoleService() // Initialize RoleService

	// Initialize controllers with their dependencies.
	authController := controllers.NewAuthController(userService) // NEW: Initialize AuthController
	userController := controllers.NewUserController(userService)
	roleController := controllers.NewRoleController(roleService)

	// Public route for authentication (no JWT middleware applied to this specific route)
	app.Post("/login", authController.Login) // This should be outside the JWT-protected group

	// Group authenticated API routes under /api prefix.
	// All routes within this group will automatically require a valid JWT
	// because the JWT middleware is applied to `app` before this function is called.
	api := app.Group("/api")

	// NEW: Logout route (requires JWT, any authenticated user can logout)
	api.Post("/auth/logout", authController.Logout) // This will be protected by the global JWT middleware on `api` group

	// NEW: Route for listing roles for dropdown, accessible by admin
	// This is now directly under /api and has its own admin role check.
	api.Get("/lstroles", middleware.HasRole("admin"), userController.GetAllRoles)

	// --- Public Protected Routes (requires JWT, no specific role check here) ---
	// Accessible by any authenticated user (i.e., with a valid JWT).
	api.Get("/dashboard", func(c *fiber.Ctx) error {
		// You can optionally retrieve user ID/roles from context here if needed
		// for displaying personalized dashboard content.
		// userID, _ := middleware.GetUserIDFromJWT(c)
		// userRoles, _ := middleware.GetUserRolesFromJWT(c)
		return c.Status(http.StatusOK).JSON(fiber.Map{
			"message": "Welcome to the protected dashboard!",
			"data":    "This data is from GoFiber backend.",
		})
	})

	// Example route to get authenticated user's profile information
	// This route was previously in main.go's example.
	api.Get("/profile", func(c *fiber.Ctx) error {
		userID, _ := middleware.GetUserIDFromJWT(c)
		userRoles, _ := middleware.GetUserRolesFromJWT(c)
		return c.Status(http.StatusOK).JSON(fiber.Map{
			"message": "Welcome to your profile!",
			"user_id": userID,
			"roles":   userRoles,
		})
	})

	// --- User Management Routes (Requires 'admin' role) ---
	// All routes within this group will require the 'admin' role.
	userManagement := api.Group("/users")
	userManagement.Use(middleware.HasRole("admin")) // Apply role-based middleware for admin
	{
		userManagement.Get("/", userController.GetAllUsers)    // GET /api/users
		userManagement.Get("/:id", userController.GetUserByID) // GET /api/users/:id
		userManagement.Post("/", userController.CreateUser)    // POST /api/users
		// userManagement.Get("/lstroles", userController.GetAllRoles) // REMOVED: Moved to directly under /api
		userManagement.Put("/:id", userController.UpdateUser)    // PUT /api/users/:id
		userManagement.Delete("/:id", userController.DeleteUser) // DELETE /api/users/:id
	}

	// --- Role Management Routes (Requires 'admin' role) ---
	// All routes within this group will require the 'admin' role.
	roleManagement := api.Group("/roles")
	roleManagement.Use(middleware.HasRole("admin")) // Apply role-based middleware for admin
	{
		roleManagement.Get("/", roleController.GetAllRoles)      // GET /api/roles (with search, pagination)
		roleManagement.Get("/:id", roleController.GetRoleByID)   // GET /api/roles/:id
		roleManagement.Post("/", roleController.CreateRole)      // POST /api/roles
		roleManagement.Put("/:id", roleController.UpdateRole)    // PUT /api/roles/:id
		roleManagement.Delete("/:id", roleController.DeleteRole) // DELETE /api/roles/:id
	}

	// --- Example of a route accessible by multiple roles ---
	// For instance, a "premium content" route that "premium_user" and "admin" can access
	premiumContent := api.Group("/premium")
	premiumContent.Use(middleware.HasRole("admin", "premium_user"))
	{
		premiumContent.Get("/special-content", func(c *fiber.Ctx) error {
			return c.Status(http.StatusOK).JSON(fiber.Map{"message": "This is highly exclusive premium content!"})
		})
	}

	// --- Example of a route accessible by 'user' or 'admin' roles ---
	userSpecificData := api.Group("/my-data")
	userSpecificData.Use(middleware.HasRole("user", "admin"))
	{
		userSpecificData.Get("/", func(c *fiber.Ctx) error {
			userID, _ := middleware.GetUserIDFromJWT(c)
			return c.Status(http.StatusOK).JSON(fiber.Map{"message": "This is data specific to you, user ID:", "user_id": userID})
		})
	}
}
