package jaeger

import (
	"cmp"
	"net/http"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go/config"
)

// Config is the configuration for the [Tracer].
// Use [New] to create a new instance of the [Tracer].
type Config struct {
	// JaegerConfig is the configuration for the jaeger client.
	// If not set, a zero value of the Configuration is used.
	JaegerConfig config.Configuration
	// AddCaller, if true, add caller info
	// to the root span tag.
	AddCaller bool
	// ServerSpanHook intercept spans in the server-side midleware before finishing it.
	// If not set, default hook function is used and adds default span tags.
	// Users can use this function to add custom tags.
	ServerSpanHook func(span opentracing.Span, w http.ResponseWriter, r *http.Request)
	// ClientSpanHook intercept spans in the client-side midleware before finishing it.
	// If not set, default hook function is used and adds default span tags.
	// Users can use this function to add custom tags.
	ClientSpanHook func(span opentracing.Span, w *http.Response, r *http.Request)
}

// New creates a new tracer from the [Config].
func New(c *Config) (*Tracer, error) {
	jc := c.JaegerConfig
	jc.Sampler = cmp.Or(jc.Sampler, &config.SamplerConfig{Type: "const", Param: 1})
	jc.ServiceName = cmp.Or(jc.ServiceName, "aileron")
	tracer, closer, err := jc.NewTracer()
	if err != nil {
		return nil, err
	}
	t := &Tracer{
		tracer:         tracer,
		closer:         closer,
		addCaller:      c.AddCaller,
		serverSpanHook: c.ServerSpanHook,
		clientSpanHook: c.ClientSpanHook,
	}
	if t.serverSpanHook == nil {
		t.serverSpanHook = serverSpanHook
	}
	if t.clientSpanHook == nil {
		t.clientSpanHook = clientSpanHook
	}
	return t, nil
}
