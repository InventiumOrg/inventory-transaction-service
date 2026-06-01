package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Health responds to GET /health (and /healthz). Matches the Java HealthController shape:
//
//	{ "status": "UP", "service": "inventory-transaction-service" }
func (h *Handlers) Health(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"status":    "UP",
		"service":   "inventory-transaction-service",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
