package otel

import (
	"context"
	"net/http"

	"github.com/aileron-projects/go/znet/zhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

var (
	_ zhttp.ServerMiddleware = &Metrics{}
	_ zhttp.ClientMiddleware = &Metrics{}
)

type Metrics struct {
	provider *sdkmetric.MeterProvider
	// serverCounter is the api call counter for
	// the server-side middleware.
	serverCounter metric.Int64Counter
	// clientCounter is the api call counter for
	// the client-side middleware.
	clientCounter metric.Int64Counter
}

// MeterProvider return the opentelemetry metric provider.
// The registry can be used to register custom metrics.
func (m *Metrics) MeterProvider() *sdkmetric.MeterProvider {
	return m.provider
}

func (m *Metrics) ServerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := zhttp.WrapResponseWriter(w)
		defer func(ctx context.Context) {
			m.serverCounter.Add(ctx, 1,
				metric.WithAttributes(
					attribute.String("method", r.Method),
					attribute.String("host", r.Host),
					attribute.String("path", r.URL.Path),
					attribute.Int("code", ww.StatusCode()),
				),
			)
		}(r.Context())
		next.ServeHTTP(ww, r)
	})
}

func (m *Metrics) ClientMiddleware(next http.RoundTripper) http.RoundTripper {
	return zhttp.RoundTripperFunc(func(r *http.Request) (resp *http.Response, err error) {
		defer func() {
			status := 0
			if resp != nil {
				status = resp.StatusCode
			}
			m.clientCounter.Add(r.Context(), 1,
				metric.WithAttributes(
					attribute.String("method", r.Method),
					attribute.String("host", r.Host),
					attribute.String("path", r.URL.Path),
					attribute.Int("code", status),
				),
			)
		}()
		return next.RoundTrip(r)
	})
}

// Finalize closes internal meter provider.
func (m *Metrics) Finalize(ctx context.Context) error {
	return m.provider.Shutdown(ctx)
}
