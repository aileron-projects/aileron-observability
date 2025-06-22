package jaeger

import (
	"context"
	"errors"
	"io"
	"net/http"
	"path"
	"runtime"

	"github.com/aileron-projects/aileron-observability/tracing"
	"github.com/aileron-projects/go/znet/zhttp"
	"github.com/aileron-projects/go/zx/zuid"
	"github.com/opentracing/opentracing-go"
)

var (
	_ zhttp.ServerMiddleware = &Tracer{}
	_ zhttp.ClientMiddleware = &Tracer{}
	_ tracing.Tracer         = &Tracer{}
)

// Middleware is a
type Tracer struct {
	// tracer is a jaeger tracer.
	tracer opentracing.Tracer
	// closer closes the tracer.
	// Close method should be called before shutting down
	// the application to flush all tracing data
	// in the internal buffer.
	closer io.Closer

	// addCaller, if true, add caller info
	// to the root span tag.
	addCaller bool

	serverSpanHook func(span opentracing.Span, w http.ResponseWriter, r *http.Request)
	clientSpanHook func(span opentracing.Span, w *http.Response, r *http.Request)
}

func (t *Tracer) ServerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := 1 // counter
		if v := r.Context().Value(tracing.ServerCtxKey); v != nil {
			c = v.(int) + 1
		}
		r = r.WithContext(context.WithValue(r.Context(), tracing.ServerCtxKey, c))

		span, ctx := t.spanContext(r, "server+"+r.URL.Path)
		defer span.Finish()
		r = r.WithContext(ctx)

		if t.addCaller {
			if ptr, file, _, ok := runtime.Caller(2); ok {
				f, _ := runtime.CallersFrames([]uintptr{ptr}).Next()
				span.SetTag("caller.file", path.Base(path.Dir(file))+"/"+path.Base(file))
				span.SetTag("caller.func", f.Function)
				span.SetTag("caller.line", f.Line)
			}
		}

		if c == 1 { // Only for root span.
			ww := zhttp.WrapResponseWriter(w)
			w = ww
			defer t.serverSpanHook(span, ww, r)
		}
		next.ServeHTTP(w, r)
	})
}

func (t *Tracer) ClientMiddleware(next http.RoundTripper) http.RoundTripper {
	return zhttp.RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
		c := 1 // counter
		if v := r.Context().Value(tracing.ClientCtxKey); v != nil {
			c = v.(int) + 1
		}
		r = r.WithContext(context.WithValue(r.Context(), tracing.ClientCtxKey, c))

		span, ctx := t.spanContext(r, "client+"+r.URL.Path)
		defer span.Finish()
		r = r.WithContext(ctx)

		if t.addCaller {
			if ptr, file, _, ok := runtime.Caller(2); ok {
				f, _ := runtime.CallersFrames([]uintptr{ptr}).Next()
				span.SetTag("caller.file", path.Base(path.Dir(file))+"/"+path.Base(file))
				span.SetTag("caller.func", f.Function)
				span.SetTag("caller.line", f.Line)
			}
		}

		res, err := next.RoundTrip(r)
		if c == 1 { // Only for root span.
			t.clientSpanHook(span, res, r)
		}
		return res, err
	})
}

// spanContext returns a new span.
func (t *Tracer) spanContext(r *http.Request, name string) (opentracing.Span, context.Context) {
	var span opentracing.Span
	if parentSpan := opentracing.SpanFromContext(r.Context()); parentSpan != nil {
		span = t.tracer.StartSpan(name, opentracing.ChildOf(parentSpan.Context()))
	} else {
		carrier := opentracing.HTTPHeadersCarrier(r.Header)
		sc, err := t.tracer.Extract(opentracing.HTTPHeaders, carrier)
		if errors.Is(err, opentracing.ErrSpanContextNotFound) {
			span = t.tracer.StartSpan(name)
		} else {
			span = t.tracer.StartSpan(name, opentracing.ChildOf(sc))
		}
	}
	return span, opentracing.ContextWithSpan(r.Context(), span)
}

// Trace is the method that can be called from any types of resources.
// Callers must update their context with the returned one.
// The returned function with finishes spans must be called when finishing spans.
func (t *Tracer) Trace(ctx context.Context, name string, tags map[string]string) (spanCtx context.Context, finish func()) {
	var span opentracing.Span
	if parentSpan := opentracing.SpanFromContext(ctx); parentSpan != nil {
		span = t.tracer.StartSpan(name, opentracing.ChildOf(parentSpan.Context()))
	} else {
		span = t.tracer.StartSpan(name)
	}
	for k, v := range tags {
		span.SetTag(k, v)
	}
	return opentracing.ContextWithSpan(ctx, span), span.Finish
}

// Finalize closes internal tracer flushing remained trace data.
func (t *Tracer) Finalize(_ context.Context) error {
	return t.closer.Close()
}

func serverSpanHook(span opentracing.Span, w http.ResponseWriter, r *http.Request) {
	if id := zuid.FromContext(r.Context(), "context"); id != "" {
		span.SetTag("context", id)
	}
	proto := "http"
	if r.TLS != nil {
		proto = "https"
	}
	span.SetTag("http.schema", proto)
	span.SetTag("http.method", r.Method)
	span.SetTag("http.path", r.URL.Path)
	span.SetTag("http.query", r.URL.RawQuery)
	span.SetTag("http.referer", r.Referer())
	span.SetTag("http.route", r.Pattern)
	span.SetTag("http.request_content_length", r.ContentLength)
	span.SetTag("net.addr", r.RemoteAddr)
	span.SetTag("net.host", r.Host)
	ww := w.(*zhttp.ResponseWrapper)
	span.SetTag("http.status_code", ww.StatusCode())
	span.SetTag("http.response_content_length", ww.Written())
}

func clientSpanHook(span opentracing.Span, w *http.Response, r *http.Request) {
	if id := zuid.FromContext(r.Context(), "context"); id != "" {
		span.SetTag("context", id)
	}
	proto := "http"
	if r.TLS != nil {
		proto = "https"
	}
	span.SetTag("http.schema", proto)
	span.SetTag("http.method", r.Method)
	span.SetTag("http.path", r.URL.Path)
	span.SetTag("http.query", r.URL.RawQuery)
	span.SetTag("http.request_content_length", r.ContentLength)
	span.SetTag("net.addr", r.RemoteAddr)
	span.SetTag("peer.host", r.URL.Host)
	if w == nil {
		span.SetTag("http.status_code", 0)
	} else {
		span.SetTag("http.status_code", w.StatusCode)
		span.SetTag("http.response_content_length", w.ContentLength)
	}
}
