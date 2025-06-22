package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/aileron-projects/aileron-observability/tracing/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/trace"
)

func main() {
	// Use gRPC exporter without TLS.
	exporter, _ := otlptracegrpc.New(context.Background(), otlptracegrpc.WithInsecure())
	t, _ := otel.New(&otel.Config{
		ProviderOpts: []trace.TracerProviderOption{
			trace.WithSampler(trace.TraceIDRatioBased(1.0)),
			trace.WithBatcher(
				exporter,
				trace.WithBatchTimeout(time.Second),
				trace.WithExportTimeout(10*time.Second),
				trace.WithBlocking(),
			),
		},
	})

	target, _ := url.Parse("http://httpbin.org")
	proxy := httputil.NewSingleHostReverseProxy(target)

	h := t.ServerMiddleware(proxy) // Apply tracing.
	h = t.ServerMiddleware(h)      // Tracer can be used multiple times.
	h = t.ServerMiddleware(h)      // Tracer can be used multiple times.

	log.Println("server listening on localhost:8080")
	svr := &http.Server{
		Addr:        ":8080",
		Handler:     h,
		ReadTimeout: 10 * time.Second,
	}
	if err := svr.ListenAndServe(); err != nil {
		panic(err)
	}
}
