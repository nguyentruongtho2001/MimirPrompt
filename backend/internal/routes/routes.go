package routes

import (
	"database/sql"

	"mimirprompt/internal/handlers"
	"mimirprompt/internal/middleware"
	"mimirprompt/internal/repository"

	"github.com/gin-gonic/gin"
)

// SetupRouter configures and returns the Gin router
func SetupRouter(db *sql.DB, imagesDir string) *gin.Engine {
	router := gin.Default()

	// Apply global middleware
	router.Use(middleware.CORS())
	router.Use(middleware.Recovery())

	// Serve static files for images
	if imagesDir != "" {
		router.Static("/images", imagesDir)
	}

	// Initialize repositories
	promptRepo := repository.NewPromptRepository(db)
	tagRepo := repository.NewTagRepository(db)
	authRepo := repository.NewAuthRepository(db)
	authorRepo := repository.NewAuthorRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	auditRepo := repository.NewAuditRepository(db)

	// Initialize handlers
	promptHandler := handlers.NewPromptHandler(promptRepo)
	tagHandler := handlers.NewTagHandler(tagRepo, db)
	authHandler := handlers.NewAuthHandler(authRepo)
	uploadHandler := handlers.NewUploadHandler(imagesDir)
	authorHandler := handlers.NewAuthorHandler(authorRepo)
	categoryHandler := handlers.NewCategoryHandler(categoryRepo)
	auditHandler := handlers.NewAuditHandler(auditRepo)

	// API routes
	api := router.Group("/api")
	{
		// ===== Public routes (no auth required) =====

		// Auth (login only - registration disabled for security)
		api.POST("/auth/login", authHandler.Login)

		// Prompts (read only)
		api.GET("/prompts", promptHandler.List)
		api.GET("/prompts/:id", promptHandler.Get)

		// Tags (read only)
		api.GET("/tags", tagHandler.List)
		api.GET("/tags/:id", tagHandler.Get)
		api.GET("/tags/slug/:slug/prompts", promptHandler.ListByTag)

		// Authors (read only)
		api.GET("/authors", authorHandler.List)
		api.GET("/authors/:id", authorHandler.Get)

		// Categories (read only)
		api.GET("/categories", categoryHandler.List)
		api.GET("/categories/:id", categoryHandler.Get)

		// Stats & Search
		api.GET("/stats", tagHandler.GetStats)
		api.GET("/search", promptHandler.Search)

		// ===== Protected routes (auth required) =====
		protected := api.Group("")
		protected.Use(middleware.AuthRequired())
		{
			// Auth
			protected.GET("/auth/profile", authHandler.GetProfile)
			protected.PUT("/auth/password", authHandler.ChangePassword)

			// Upload
			protected.POST("/upload", uploadHandler.UploadImage)

			// Prompts (write operations)
			protected.POST("/prompts", promptHandler.Create)
			protected.PUT("/prompts/:id", promptHandler.Update)
			protected.DELETE("/prompts/:id", promptHandler.Delete)

			// Prompt bulk operations
			protected.POST("/prompts/bulk-delete", promptHandler.BulkDelete)
			protected.POST("/prompts/bulk-visibility", promptHandler.BulkUpdateVisibility)

			// Tags (write operations)
			protected.POST("/tags", tagHandler.Create)
			protected.PUT("/tags/:id", tagHandler.Update)
			protected.DELETE("/tags/:id", tagHandler.Delete)

			// Authors (write operations)
			protected.POST("/authors", authorHandler.Create)
			protected.PUT("/authors/:id", authorHandler.Update)
			protected.DELETE("/authors/:id", authorHandler.Delete)

			// Categories (write operations)
			protected.POST("/categories", categoryHandler.Create)
			protected.PUT("/categories/:id", categoryHandler.Update)
			protected.DELETE("/categories/:id", categoryHandler.Delete)

			// Audit logs (admin only)
			protected.GET("/audit-logs", auditHandler.List)
			protected.GET("/audit-logs/:entity_type/:entity_id", auditHandler.GetByEntity)
		}
	}

	return router
}
