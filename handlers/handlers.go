package handlers

import (
	"inventory-transaction-service/observability"
	"inventory-transaction-service/processors"
	"inventory-transaction-service/repository"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// Handlers groups dependencies shared by all HTTP handlers.
type Handlers struct {
	Repo     *repository.TransactionRepository
	Producer *processors.InventoryProducer
	Metrics  *observability.PrometheusMetrics
	Tracer   trace.Tracer
}

// NewHandlers constructs a Handlers value with a tracer named after the package.
func NewHandlers(
	repo *repository.TransactionRepository,
	producer *processors.InventoryProducer,
	metrics *observability.PrometheusMetrics,
) *Handlers {
	return &Handlers{
		Repo:     repo,
		Producer: producer,
		Metrics:  metrics,
		Tracer:   otel.Tracer("inventory-transaction-service/handlers"),
	}
}
