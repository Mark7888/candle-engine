// Package api wires up all REST API routes under the /api/v1 prefix.
package api

import (
	"candle-engine/pkg/webserver/api/handlers"
	"candle-engine/pkg/webserver/api/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes attaches all REST API routes to the provided Gin router.
// corsOrigin is forwarded to the CORS middleware; pass "*" for local development
// or a specific origin string (e.g. "https://app.example.com") for production.
// A Swagger UI is also served at /swagger/.
func RegisterRoutes(router *gin.Engine, corsOrigin string) {
	// Apply CORS globally so browser preflight requests for /api/* succeed.
	router.Use(middleware.CORS(corsOrigin))

	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", handlers.Health)

		projects := v1.Group("/projects")
		{
			projects.GET("", handlers.ListProjects)
			projects.GET("/:id", handlers.GetProject)
			projects.GET("/:id/render", handlers.RenderProject)
		}
	}
}
