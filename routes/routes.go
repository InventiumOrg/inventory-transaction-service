package routes

import (
	"inventory-transaction-service/handlers"

	"github.com/gin-gonic/gin"
)

// Route exposes route-group helpers backed by a shared handlers value.
type Route struct {
	handlers *handlers.Handlers
}

func NewRoute(h *handlers.Handlers) *Route {
	return &Route{handlers: h}
}

// AddHealthRoutes registers liveness/health endpoints.
//
// `/health` matches the Java service. `/healthz` and `/readyz` are added for
// Kubernetes probe compatibility (matches the template).
func (r *Route) AddHealthRoutes(router *gin.Engine) {
	router.GET("/health", r.handlers.Health)
	router.GET("/healthz", r.handlers.Health)
	router.GET("/readyz", r.handlers.Health)
}

// AddTransactionRoutes registers `/api/v1/transactions` endpoints.
func (r *Route) AddTransactionRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1")
	{
		tx := v1.Group("/transactions")
		{
			tx.POST("", r.handlers.CreateTransaction)
			tx.GET("/:inventoryId/:id", r.handlers.GetTransaction)
			tx.PUT("/:inventoryId/:id", r.handlers.UpdateTransaction)
		}
	}
}
