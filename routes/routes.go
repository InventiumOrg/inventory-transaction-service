package routes

import (
	"inventory-transaction-service/handlers"
	"inventory-transaction-service/middlewares"

	"github.com/gin-gonic/gin"
)

// Route exposes route-group helpers backed by a shared handlers value.
type Route struct {
	handlers      *handlers.Handlers
	healthChecker *middlewares.HealthChecker
}

func NewRoute(h *handlers.Handlers, hc *middlewares.HealthChecker) *Route {
	return &Route{handlers: h, healthChecker: hc}
}

// AddHealthRoutes registers liveness/readiness endpoints.
//
//   - GET /health   → simple liveness (matches Java HealthController)
//   - GET /healthz  → liveness probe  (Kubernetes convention)
//   - GET /readyz   → readiness probe: 200 only when DynamoDB is reachable
func (r *Route) AddHealthRoutes(router *gin.Engine) {
	router.GET("/health", r.handlers.Health)
	router.GET("/healthz", r.healthChecker.LivezHandler)
	router.GET("/readyz", r.healthChecker.ReadyzHandler)
}

// AddTransactionRoutes registers `/api/v1/transactions` endpoints.
func (r *Route) AddTransactionRoutes(router *gin.Engine) {
	v1 := router.Group("/api/v1")
	{
		tx := v1.Group("/transactions")
		{
			tx.POST("", r.handlers.CreateTransaction)
			tx.GET("/:inventoryId", r.handlers.ListTransactions)
			tx.GET("/:inventoryId/:id", r.handlers.GetTransaction)
			tx.PUT("/:inventoryId/:id", r.handlers.UpdateTransaction)
		}
	}
}
