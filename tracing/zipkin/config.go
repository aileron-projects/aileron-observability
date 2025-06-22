package zipkin

import (
	"net/http"

	"github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/reporter"
)

type Config struct {
	Reporter   reporter.Reporter
	TracerOpts []zipkin.TracerOption

	AddCaller bool

	ServerSpanHook func(span zipkin.Span, w http.ResponseWriter, r *http.Request)
	ClientSpanHook func(span zipkin.Span, w *http.Response, r *http.Request)
}

func New(c *Config) (*Tracer, error) {
	tracer, err := zipkin.NewTracer(c.Reporter, c.TracerOpts...)
	if err != nil {
		return nil, err
	}
	t := &Tracer{
		tracer:         tracer,
		reporter:       c.Reporter,
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
