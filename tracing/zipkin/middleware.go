package zipkin

import (
	"context"
	"errors"
	"net/http"
	"path"
	"runtime"
	"strconv"

	"github.com/aileron-projects/aileron-observability/tracing"
	"github.com/aileron-projects/go/znet/zhttp"
	"github.com/aileron-projects/go/zx/zuid"
	"github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/propagation/b3"
	"github.com/openzipkin/zipkin-go/reporter"
)

var (
	_ zhttp.ServerMiddleware = &Tracer{}
	_ zhttp.ClientMiddleware = &Tracer{}
	_ tracing.Tracer         = &Tracer{}
)

type Tracer struct {
	// tracer is a zipkin tracer.
	tracer   *zipkin.Tracer
	reporter reporter.Reporter

	addCaller bool

	serverSpanHook func(span zipkin.Span, w http.ResponseWriter, r *http.Request)
	clientSpanHook func(span zipkin.Span, w *http.Response, r *http.Request)
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
				span.Tag("caller.file", path.Base(path.Dir(file))+"/"+path.Base(file))
				span.Tag("caller.func", f.Function)
				span.Tag("caller.line", strconv.Itoa(f.Line))
			}
		}

		if c == 1 {
			ww := zhttp.WrapResponseWriter(w)
			w = ww
			defer t.serverSpanHook(span, w, r)
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
				span.Tag("caller.file", path.Base(path.Dir(file))+"/"+path.Base(file))
				span.Tag("caller.func", f.Function)
				span.Tag("caller.line", strconv.Itoa(f.Line))
			}
		}

		res, err := next.RoundTrip(r)
		if c == 1 {
			t.clientSpanHook(span, res, r)
		}
		return res, err
	})
}

// spanContext returns a new span.
func (t *Tracer) spanContext(r *http.Request, name string) (zipkin.Span, context.Context) {
	var span zipkin.Span
	ctx := r.Context()
	if parent := zipkin.SpanFromContext(ctx); parent != nil {
		span = t.tracer.StartSpan(name, zipkin.Parent(parent.Context()))
	} else {
		sc := t.tracer.Extract(b3.ExtractHTTP(r))
		if errors.Is(sc.Err, b3.ErrEmptyContext) {
			span = t.tracer.StartSpan(name)
		} else {
			span = t.tracer.StartSpan(name, zipkin.Parent(sc))
		}
	}
	return span, zipkin.NewContext(ctx, span)
}

// Trace is the method that can be called from any types of resources.
// Callers must update their context with the returned one.
// The returned function with finishes spans must be called when finishing spans.
func (t *Tracer) Trace(ctx context.Context, name string, tags map[string]string) (spanCtx context.Context, finish func()) {
	var span zipkin.Span
	if parent := zipkin.SpanFromContext(ctx); parent != nil {
		span = t.tracer.StartSpan(name, zipkin.Parent(parent.Context()))
	} else {
		span = t.tracer.StartSpan(name)
	}
	for k, v := range tags {
		span.Tag(k, v)
	}
	return zipkin.NewContext(ctx, span), span.Finish
}

// Finalize closes internal tracer flushing remained trace data.
func (t *Tracer) Finalize(_ context.Context) error {
	return t.reporter.Close()
}

func serverSpanHook(span zipkin.Span, w http.ResponseWriter, r *http.Request) {
	if id := zuid.FromContext(r.Context(), "context"); id != "" {
		span.Tag("context", id)
	}
	proto := "http"
	if r.TLS != nil {
		proto = "https"
	}
	span.Tag("http.schema", proto)
	span.Tag("http.method", r.Method)
	span.Tag("http.path", r.URL.Path)
	span.Tag("http.query", r.URL.RawQuery)
	span.Tag("http.referer", r.Referer())
	span.Tag("http.route", r.Pattern)
	span.Tag("http.request_content_length", strconv.FormatInt(r.ContentLength, 10))
	span.Tag("net.addr", r.RemoteAddr)
	span.Tag("net.host", r.Host)
	ww := w.(*zhttp.ResponseWrapper)
	span.Tag("http.status_code", strconv.Itoa(ww.StatusCode()))
	span.Tag("http.response_content_length", strconv.FormatInt(ww.Written(), 10))
}

func clientSpanHook(span zipkin.Span, w *http.Response, r *http.Request) {
	if id := zuid.FromContext(r.Context(), "context"); id != "" {
		span.Tag("context", id)
	}
	proto := "http"
	if r.TLS != nil {
		proto = "https"
	}
	span.Tag("http.schema", proto)
	span.Tag("http.method", r.Method)
	span.Tag("http.path", r.URL.Path)
	span.Tag("http.query", r.URL.RawQuery)
	span.Tag("http.request_content_length", strconv.FormatInt(r.ContentLength, 10))
	span.Tag("net.addr", r.RemoteAddr)
	span.Tag("peer.host", r.URL.Host)
	if w == nil {
		span.Tag("http.status_code", "0")
	} else {
		span.Tag("http.status_code", strconv.Itoa(w.StatusCode))
		span.Tag("http.response_content_length", strconv.FormatInt(w.ContentLength, 10))
	}
}
