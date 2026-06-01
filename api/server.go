package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"inventory-transaction-service/config"
	"inventory-transaction-service/handlers"
	"inventory-transaction-service/middlewares"
	"inventory-transaction-service/observability"
	"inventory-transaction-service/processors"
	"inventory-transaction-service/repository"
	"inventory-transaction-service/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// Server wires Gin, dependencies, and the OTEL/Prometheus pipelines.
type Server struct {
	router            *gin.Engine
	cfg               config.Config
	repo              *repository.TransactionRepository
	producer          *processors.InventoryProducer
	otelShutdown      func(context.Context) error
	metrics           *observability.AppMetrics
	prometheusMetrics *observability.PrometheusMetrics
	httpServer        *http.Server
}

// NewServer constructs the HTTP server and all of its collaborators.
func NewServer(
	ctx context.Context,
	cfg config.Config,
	repo *repository.TransactionRepository,
) (*Server, error) {
	// Boot OTEL.
	otelShutdown, err := observability.SetupOTelSDK(
		ctx,
		cfg.ServiceName, "1.0.0",
		cfg.OTELExporterOTLPEndpoint, cfg.OTELExporterOTLPHeaders,
		cfg.TracingSamplingRatio,
	)
	if err != nil {
		slog.Error("Failed to setup OpenTelemetry", slog.Any("error", err))
		otelShutdown = func(context.Context) error { return nil }
	}

	otelMetrics, err := observability.CreateMetrics(cfg.ServiceName)
	if err != nil {
		slog.Error("Failed to create OTEL metrics", slog.Any("error", err))
	}

	promMetrics := observability.NewPrometheusMetrics(cfg.ServiceName)

	producer := processors.NewInventoryProducer(promMetrics)
	if err := producer.Init(cfg); err != nil {
		return nil, fmt.Errorf("kafka producer init: %w", err)
	}

	router := gin.New()
	router.Use(gin.Logger(), middlewares.Recovery())
	router.Use(promMetrics.PrometheusMiddleware())

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Bearer"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	observability.SetupPrometheusEndpoint(router)

	srv := &Server{
		router:            router,
		cfg:               cfg,
		repo:              repo,
		producer:          producer,
		otelShutdown:      otelShutdown,
		metrics:           otelMetrics,
		prometheusMetrics: promMetrics,
	}
	router.Use(srv.otelMetricsMiddleware())

	h := handlers.NewHandlers(repo, producer, promMetrics)
	hc := middlewares.NewHealthChecker(repo, promMetrics)
	r := routes.NewRoute(h, hc)
	r.AddHealthRoutes(router)
	r.AddTransactionRoutes(router)

	return srv, nil
}

// Run binds the underlying HTTP server to `addr` and blocks until it exits.
func (s *Server) Run(addr string) error {
	slog.Info("Starting HTTP server",
		slog.String("address", addr),
		slog.String("service", s.cfg.ServiceName))

	s.httpServer = &http.Server{
		Addr:              addr,
		Handler:           s.router,
		ReadHeaderTimeout: 5 * time.Second,
	}
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Shutdown gracefully tears down the HTTP server and observability pipelines.
func (s *Server) Shutdown(ctx context.Context) error {
	slog.Info("Shutting down server...")
	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			slog.Error("HTTP server shutdown error", slog.Any("error", err))
		}
	}
	if s.producer != nil {
		if err := s.producer.Close(); err != nil {
			slog.Error("Kafka producer close error", slog.Any("error", err))
		}
	}
	if s.otelShutdown != nil {
		if err := s.otelShutdown(ctx); err != nil {
			slog.Error("OTEL shutdown error", slog.Any("error", err))
		}
	}
	return nil
}

// otelMetricsMiddleware emits OTEL HTTP metrics in addition to the Prometheus
// middleware (different aggregation backends).
func (s *Server) otelMetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		if s.metrics == nil {
			return
		}
		duration := time.Since(start).Seconds()
		attrs := []attribute.KeyValue{
			attribute.String("method", c.Request.Method),
			attribute.String("route", c.FullPath()),
			attribute.Int("status_code", c.Writer.Status()),
		}
		s.metrics.RequestCounter.Add(c.Request.Context(), 1, metric.WithAttributes(attrs...))
		s.metrics.RequestDuration.Record(c.Request.Context(), duration, metric.WithAttributes(attrs...))
	}
}
