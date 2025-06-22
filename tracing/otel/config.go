package otel

import (
	"cmp"
	"net/http"

	"go.opentelemetry.io/contrib/propagators/autoprop"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	// ScopeName is the instrumentation scope name.
	ScopeName = "github.com/aileron-projects/aileron-observability/tracing/otel"
)

type Config struct {
	// ServiceName is the application service name.
	// If empty, default "aileron" is used.
	ServiceName string

	Props        []propagation.TextMapPropagator
	ProviderOpts []sdktrace.TracerProviderOption
	TracerOpts   []trace.TracerOption
	Attributes   []attribute.KeyValue

	AddCaller bool

	ServerSpanFunc func(span trace.Span, w http.ResponseWriter, r *http.Request)
	ClientSpanFunc func(span trace.Span, w *http.Response, r *http.Request)
}

func New(c *Config) (*Tracer, error) {
	c.Attributes = append(c.Attributes, semconv.ServiceName(cmp.Or(c.ServiceName, "aileron")))
	c.ProviderOpts = append(c.ProviderOpts, sdktrace.WithResource(resource.NewWithAttributes(semconv.SchemaURL, c.Attributes...)))
	tracerProvider := sdktrace.NewTracerProvider(c.ProviderOpts...)
	props := c.Props
	if len(props) == 0 {
		props = append(props, propagation.TraceContext{}, propagation.Baggage{})
	}

	tracer := tracerProvider.Tracer(ScopeName, c.TracerOpts...)
	t := &Tracer{
		tracer:         tracer,
		tp:             tracerProvider,
		pg:             autoprop.NewTextMapPropagator(props...),
		addCaller:      c.AddCaller,
		serverSpanHook: c.ServerSpanFunc,
		clientSpanHook: c.ClientSpanFunc,
	}
	if t.serverSpanHook == nil {
		t.serverSpanHook = serverSpanHook
	}
	if t.clientSpanHook == nil {
		t.clientSpanHook = clientSpanHook
	}
	return t, nil
}
