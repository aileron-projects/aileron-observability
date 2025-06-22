package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/aileron-projects/aileron-observability/metrics/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/sdk/metric"
)

func main() {
	// Use HTTP exporter without TLS.
	exporter, _ := otlpmetrichttp.New(context.Background(), otlpmetrichttp.WithInsecure())
	m, _ := otel.New(&otel.Config{
		ProviderOpts: []metric.Option{
			metric.WithReader(metric.NewPeriodicReader(
				exporter,
				metric.WithInterval(time.Second),
				metric.WithTimeout(10*time.Second),
			)),
		},
	})

	target, _ := url.Parse("http://httpbin.org")
	proxy := httputil.NewSingleHostReverseProxy(target)

	log.Println("server listening on localhost:8080")
	svr := &http.Server{
		Addr:        ":8080",
		Handler:     m.ServerMiddleware(proxy),
		ReadTimeout: 10 * time.Second,
	}
	if err := svr.ListenAndServe(); err != nil {
		panic(err)
	}
}
