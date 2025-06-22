package otel

import (
	"cmp"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

const (
	// ScopeName is the instrumentation scope name.
	ScopeName = "github.com/aileron-projects/aileron-observability/metrics/otel"
)

// Config is the OpenTelemetry middleware configurations.
type Config struct {
	// ServiceName is the application service name.
	// If empty, default "aileron" is used.
	ServiceName string
	// ProviderOpts is the options for MeterProvider.
	ProviderOpts []sdkmetric.Option
	// MeterOpts is the options used when creating
	// a meter from provider.
	MeterOpts []metric.MeterOption
}

func New(c *Config) (*Metrics, error) {
	service := sdkmetric.WithResource(resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceName(cmp.Or(c.ServiceName, "aileron"))))
	provider := sdkmetric.NewMeterProvider(append(c.ProviderOpts, service)...)
	_ = runtime.Start(
		runtime.WithMeterProvider(provider),
		runtime.WithMinimumReadMemStatsInterval(time.Second),
	)

	meter := provider.Meter(ScopeName, c.MeterOpts...)
	serverCounter, _ := meter.Int64Counter(
		"http_requests_total",
		metric.WithDescription("Total number of received http requests"),
	)
	clientCounter, _ := meter.Int64Counter(
		"http_client_requests_total",
		metric.WithDescription("Total number of sent http requests"),
	)

	return &Metrics{
		provider:      provider,
		serverCounter: serverCounter,
		clientCounter: clientCounter,
	}, nil
}
