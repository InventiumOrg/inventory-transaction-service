package middlewares

import (
	"context"
	"net/http"
	"time"

	"inventory-transaction-service/observability"
	"inventory-transaction-service/repository"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// HealthChecker provides gin handlers for liveness and readiness probes.
//
//   - LivezHandler  → GET /healthz  — fast liveness: always 200 while the
//     process is running.
//   - ReadyzHandler → GET /readyz   — readiness: 200 only when DynamoDB is
//     reachable, 503 otherwise.
type HealthChecker struct {
	repo    *repository.TransactionRepository
	metrics *observability.PrometheusMetrics
	tracer  trace.Tracer
}

// NewHealthChecker wires a HealthChecker with its dependencies.
func NewHealthChecker(
	repo *repository.TransactionRepository,
	metrics *observability.PrometheusMetrics,
) *HealthChecker {
	return &HealthChecker{
		repo:    repo,
		metrics: metrics,
		tracer:  otel.Tracer("inventory-transaction-service/health"),
	}
}

// LivezHandler is a lightweight liveness probe — it only confirms the process
// is alive and the HTTP server is responding.
func (hc *HealthChecker) LivezHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "inventory-transaction-service",
	})
}

// ReadyzHandler is a readiness probe that verifies DynamoDB connectivity
// before reporting the service as ready to receive traffic.
func (hc *HealthChecker) ReadyzHandler(c *gin.Context) {
	_, span := hc.tracer.Start(c.Request.Context(), "readyz")
	defer span.End()

	start := time.Now()
	err := hc.repo.Ping(context.Background())
	duration := time.Since(start)

	if hc.metrics != nil {
		hc.metrics.RecordDBOperation("ping", "connection", duration, err)
	}

	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("health.status", "not_ready"))

		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "not ready",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"service":   "inventory-transaction-service",
			"checks": gin.H{
				"database": "failed: " + err.Error(),
			},
		})
		return
	}

	span.SetAttributes(attribute.String("health.status", "ready"))

	c.JSON(http.StatusOK, gin.H{
		"status":    "ready",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "inventory-transaction-service",
		"checks": gin.H{
			"database": "ok",
		},
	})
}
