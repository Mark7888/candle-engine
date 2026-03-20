// Package middleware provides Gin middleware used by the REST API.
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS returns a Gin middleware that sets CORS headers for the given allowed
// origins. Pass "*" to allow all origins (suitable for local development only).
// For production, provide the exact frontend origin(s), e.g. "https://app.example.com".
func CORS(allowedOrigins ...string) gin.HandlerFunc {
	allowed := strings.Join(allowedOrigins, ", ")
	if allowed == "" {
		allowed = "*"
	}

	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", allowed)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept")

		// Handle preflight requests.
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
