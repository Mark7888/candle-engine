package handlers

import (
	"net/http"
	"time"

	"candle-engine/pkg/webserver/api/models"

	"github.com/gin-gonic/gin"
)

// Health godoc
//
//	@Summary		Health check
//	@Description	Returns the operational status of the API server and the current server time.
//	@Tags			system
//	@Produce		json
//	@Success		200	{object}	models.HealthResponse
//	@Router			/api/v1/health [get]
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, models.HealthResponse{Status: "ok", ServerTime: time.Now().UTC()})
}
