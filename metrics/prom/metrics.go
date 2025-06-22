package prom

import (
	"net/http"
	"strconv"

	"github.com/aileron-projects/go/znet/zhttp"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	_ http.Handler           = &Metrics{}
	_ zhttp.ServerMiddleware = &Metrics{}
	_ zhttp.ClientMiddleware = &Metrics{}
)

// Metrics collects metrics and export them as prometheus format.
// Metrics implements [http.Handler], ServerMiddleware and ClientMiddleware interface.
type Metrics struct {
	metrics http.Handler         // prometheus metrics handler.
	reg     *prometheus.Registry // prometheus registry.
	// serverCounter is the api call counter for
	// the server-side middleware.
	serverCounter *prometheus.CounterVec
	// clientCounter is the api call counter for
	// the client-side middleware.
	clientCounter *prometheus.CounterVec
}

// Registry return the prometheus registry.
func (m *Metrics) Registry() *prometheus.Registry {
	return m.reg
}

func (m *Metrics) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.metrics.ServeHTTP(w, r)
}

func (m *Metrics) ServerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := zhttp.WrapResponseWriter(w)
		defer func() {
			m.serverCounter.With(prometheus.Labels{
				"method": r.Method,
				"host":   r.Host,
				"path":   r.URL.Path,
				"code":   strconv.Itoa(ww.StatusCode()),
			}).Inc()
		}()
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
			m.clientCounter.With(prometheus.Labels{
				"method": r.Method,
				"host":   r.URL.Host,
				"path":   r.URL.Path,
				"code":   strconv.Itoa(status),
			}).Inc()
		}()
		return next.RoundTrip(r)
	})
}
