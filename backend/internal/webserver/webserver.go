// Package webserver initialises and starts the Gin HTTP server.
//
//	@title			Candle Engine API
//	@version		1.0
//	@description	REST API for the Candle Engine charting system.
//	@host			localhost:8080
//	@BasePath		/
//	@schemes		http
package webserver

import (
	"log"
	"os"

	"candle-engine/internal/webserver/api"
	sse_handler "candle-engine/internal/webserver/sse"

	_ "candle-engine/docs" // Swagger generated docs

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// InitServer configures and starts the Gin router on localhost:8080.
//
// Configuration via environment variables:
//   - CORS_ALLOWED_ORIGIN: allowed CORS origin for the REST API (default: "*").
//     Set to your frontend URL in production, e.g. "https://app.example.com".
func InitServer() {
	router := gin.Default()

	// REST API routes (JSON) at /api/v1/*.
	// Read the allowed CORS origin from an environment variable so that
	// production deployments can lock this down without a code change.
	corsOrigin := os.Getenv("CORS_ALLOWED_ORIGIN")
	if corsOrigin == "" {
		corsOrigin = "*" // permissive default for local development
	}
	api.RegisterRoutes(router, corsOrigin)

	// Swagger UI at /swagger/index.html.
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Server-Sent Events endpoint for real-time browser push.
	sse_handler.RegisterRoute(router, "/events/")

	log.Println("Webserver listening on localhost:8080")
	if err := router.Run("localhost:8080"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
