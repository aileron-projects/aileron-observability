package prom

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Confis is the configuration for the [Metrics].
// Use [New] to create a new instance of the [Metrics].
type Config struct {
	// HandlerOpts is the option for prometheus handler.
	HandlerOpts promhttp.HandlerOpts
	// Collectors is the list of additional
	// prometheus collectors.
	Collectors []prometheus.Collector
}

// New returns a new instance of the [Metrics] from c.
func New(c *Config) (*Metrics, error) {
	reg := prometheus.NewRegistry()
	collectors.NewProcessCollector(collectors.ProcessCollectorOpts{})
	reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	reg.MustRegister(collectors.NewGoCollector())
	for _, c := range c.Collectors {
		if err := reg.Register(c); err != nil {
			return nil, err
		}
	}
	handler := promhttp.InstrumentMetricHandler(reg, promhttp.HandlerFor(reg, c.HandlerOpts))

	serverCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of received http requests",
		},
		[]string{"host", "path", "code", "method"},
	)
	reg.MustRegister(serverCounter)

	clientCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_client_requests_total",
			Help: "Total number of sent http requests",
		},
		[]string{"host", "path", "code", "method"},
	)
	reg.MustRegister(clientCounter)

	return &Metrics{
		metrics:       handler,
		serverCounter: serverCounter,
		clientCounter: clientCounter,
	}, nil
}
