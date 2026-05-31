package observability

import (
	"log/slog"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// PrometheusMetrics holds all Prometheus metrics for the inventory-transaction-service.
type PrometheusMetrics struct {
	HTTPRequestsTotal       *prometheus.CounterVec
	HTTPRequestDuration     *prometheus.HistogramVec
	HTTPRequestsInFlight    prometheus.Gauge
	HTTPResponseStatusTotal *prometheus.CounterVec

	DBOperationDuration *prometheus.HistogramVec
	DBOperationErrors   *prometheus.CounterVec

	TransactionOperationsTotal *prometheus.CounterVec
	KafkaPublishesTotal        *prometheus.CounterVec
}

// NewPrometheusMetrics registers and returns the metrics for the service.
func NewPrometheusMetrics(serviceName string) *PrometheusMetrics {
	m := &PrometheusMetrics{
		HTTPRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{Name: "http_requests_total", Help: "Total HTTP requests"},
			[]string{"method", "endpoint", "status_code"},
		),
		HTTPRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),
		HTTPRequestsInFlight: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current number of HTTP requests being processed",
		}),
		HTTPResponseStatusTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_response_status_total",
				Help: "Total HTTP responses by status class",
			},
			[]string{"method", "endpoint", "status_class"},
		),

		DBOperationDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "database_operation_duration_seconds",
				Help:    "Database operation duration in seconds",
				Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
			},
			[]string{"operation", "table"},
		),
		DBOperationErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "database_operation_errors_total",
				Help: "Total number of database operation errors",
			},
			[]string{"operation", "table", "error_type"},
		),

		TransactionOperationsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "transaction_operations_total",
				Help: "Total transaction record operations",
			},
			[]string{"operation", "type"},
		),
		KafkaPublishesTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kafka_publishes_total",
				Help: "Total Kafka inventory status publishes",
			},
			[]string{"topic", "status"},
		),
	}

	prometheus.MustRegister(
		m.HTTPRequestsTotal,
		m.HTTPRequestDuration,
		m.HTTPRequestsInFlight,
		m.HTTPResponseStatusTotal,
		m.DBOperationDuration,
		m.DBOperationErrors,
		m.TransactionOperationsTotal,
		m.KafkaPublishesTotal,
	)

	slog.Info("Prometheus metrics registered", slog.String("service", serviceName))
	return m
}

func statusClass(code int) string {
	switch {
	case code >= 200 && code < 300:
		return "2xx"
	case code >= 300 && code < 400:
		return "3xx"
	case code >= 400 && code < 500:
		return "4xx"
	case code >= 500:
		return "5xx"
	default:
		return "1xx"
	}
}

// PrometheusMiddleware collects HTTP metrics for every request.
func (m *PrometheusMetrics) PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}
		start := time.Now()
		m.HTTPRequestsInFlight.Inc()
		defer m.HTTPRequestsInFlight.Dec()

		c.Next()

		route := c.FullPath()
		if route == "" {
			route = "unknown"
		}
		duration := time.Since(start).Seconds()
		code := c.Writer.Status()

		m.HTTPRequestsTotal.WithLabelValues(c.Request.Method, route, strconv.Itoa(code)).Inc()
		m.HTTPRequestDuration.WithLabelValues(c.Request.Method, route).Observe(duration)
		m.HTTPResponseStatusTotal.WithLabelValues(c.Request.Method, route, statusClass(code)).Inc()
	}
}

// RecordDBOperation records duration & error metrics for a DB call.
func (m *PrometheusMetrics) RecordDBOperation(operation, table string, duration time.Duration, err error) {
	m.DBOperationDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
	if err != nil {
		m.DBOperationErrors.WithLabelValues(operation, table, "unknown").Inc()
	}
}

// RecordTransactionOperation records a business-level transaction operation.
func (m *PrometheusMetrics) RecordTransactionOperation(operation, txType string) {
	m.TransactionOperationsTotal.WithLabelValues(operation, txType).Inc()
}

// RecordKafkaPublish records a Kafka publish attempt and its result.
func (m *PrometheusMetrics) RecordKafkaPublish(topic, status string) {
	m.KafkaPublishesTotal.WithLabelValues(topic, status).Inc()
}

// SetupPrometheusEndpoint exposes /metrics on the supplied Gin router.
func SetupPrometheusEndpoint(router *gin.Engine) {
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	slog.Info("Prometheus metrics endpoint configured at /metrics")
}
