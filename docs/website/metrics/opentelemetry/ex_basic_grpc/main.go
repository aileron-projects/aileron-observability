package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/aileron-projects/aileron-observability/metrics/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/sdk/metric"
)

func main() {
	// Use gRPC exporter without TLS.
	exporter, _ := otlpmetricgrpc.New(context.Background(), otlpmetricgrpc.WithInsecure())
	_, _ = otel.New(&otel.Config{
		ProviderOpts: []metric.Option{
			metric.WithReader(metric.NewPeriodicReader(
				exporter,
				metric.WithInterval(time.Second),
				metric.WithTimeout(10*time.Second),
			)),
		},
	})

	log.Println("server listening on localhost:8080")
	svr := &http.Server{
		Addr: ":8080",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Println(r.Method, r.URL.Path) // Simple access log.
			_, _ = w.Write([]byte("Hello!!"))
		}),
		ReadTimeout: 10 * time.Second,
	}
	if err := svr.ListenAndServe(); err != nil {
		panic(err)
	}
}
