package observability

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// SetupOTelSDK bootstraps OpenTelemetry tracing+metrics pipelines pointed at
// the configured OTLP HTTP collector. Returns a shutdown function that should
// be invoked on application exit.
func SetupOTelSDK(
	ctx context.Context,
	serviceName, serviceVersion, otelCollectorEndpoint, otelHeaders string,
	samplingRatio float64,
) (func(context.Context) error, error) {
	var shutdownFuncs []func(context.Context) error

	shutdown := func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	handleErr := func(inErr error) error {
		return errors.Join(inErr, shutdown(ctx))
	}

	res := newResource(serviceName, serviceVersion)

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	tp, err := newTracerProvider(ctx, res, otelCollectorEndpoint, otelHeaders, samplingRatio)
	if err != nil {
		return shutdown, handleErr(err)
	}
	shutdownFuncs = append(shutdownFuncs, tp.Shutdown)
	otel.SetTracerProvider(tp)

	mp := newMeterProvider(res)
	shutdownFuncs = append(shutdownFuncs, mp.Shutdown)
	otel.SetMeterProvider(mp)

	return shutdown, nil
}

func newResource(serviceName, serviceVersion string) *resource.Resource {
	attrs := []attribute.KeyValue{
		semconv.ServiceName(serviceName),
		semconv.ServiceVersion(serviceVersion),
		semconv.ServiceInstanceID(serviceName),
	}

	// OTEL_RESOURCE_ATTRIBUTES=key1=value1,key2=value2
	if extra := os.Getenv("OTEL_RESOURCE_ATTRIBUTES"); extra != "" {
		for _, pair := range strings.Split(extra, ",") {
			if kv := strings.SplitN(strings.TrimSpace(pair), "=", 2); len(kv) == 2 {
				attrs = append(attrs, attribute.String(strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1])))
			}
		}
	}

	return resource.NewWithAttributes(semconv.SchemaURL, attrs...)
}

func newTracerProvider(
	ctx context.Context,
	res *resource.Resource,
	endpoint, headers string,
	samplingRatio float64,
) (*trace.TracerProvider, error) {
	if endpoint == "" {
		slog.Info("OTLP endpoint not set; using a no-op tracer provider")
		return trace.NewTracerProvider(trace.WithResource(res)), nil
	}

	headerMap := map[string]string{}
	for _, pair := range strings.Split(headers, ",") {
		if kv := strings.SplitN(pair, "=", 2); len(kv) == 2 {
			headerMap[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}

	opts := []otlptracehttp.Option{
		otlptracehttp.WithInsecure(),
		otlptracehttp.WithEndpoint(endpoint),
	}
	if len(headerMap) > 0 {
		opts = append(opts, otlptracehttp.WithHeaders(headerMap))
	}

	exporter, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, err
	}

	sampler := trace.ParentBased(trace.TraceIDRatioBased(samplingRatio))
	if samplingRatio <= 0 {
		sampler = trace.NeverSample()
	} else if samplingRatio >= 1 {
		sampler = trace.AlwaysSample()
	}

	return trace.NewTracerProvider(
		trace.WithBatcher(exporter,
			trace.WithBatchTimeout(5*time.Second),
			trace.WithMaxExportBatchSize(512),
		),
		trace.WithResource(res),
		trace.WithSampler(sampler),
	), nil
}

func newMeterProvider(res *resource.Resource) *sdkmetric.MeterProvider {
	// Push exporter intentionally omitted; Prometheus pull endpoint at /metrics
	// remains the primary metrics path (matches the Java service's design).
	return sdkmetric.NewMeterProvider(sdkmetric.WithResource(res))
}

// AppMetrics holds the application-level metrics emitted via OTEL.
type AppMetrics struct {
	RequestCounter  metric.Int64Counter
	RequestDuration metric.Float64Histogram
}

// CreateMetrics constructs the common HTTP request OTEL metrics.
func CreateMetrics(serviceName string) (*AppMetrics, error) {
	meter := otel.Meter(serviceName)

	counter, err := meter.Int64Counter("http_requests_total",
		metric.WithDescription("Total number of HTTP requests"))
	if err != nil {
		return nil, err
	}

	hist, err := meter.Float64Histogram("http_request_duration_seconds",
		metric.WithDescription("HTTP request duration in seconds"))
	if err != nil {
		return nil, err
	}

	return &AppMetrics{RequestCounter: counter, RequestDuration: hist}, nil
}
