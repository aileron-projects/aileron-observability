package otel

import (
	"context"
	"net/http"
	"path"
	"runtime"

	"github.com/aileron-projects/aileron-observability/tracing"
	"github.com/aileron-projects/go/znet/zhttp"
	"github.com/aileron-projects/go/zx/zuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

var (
	_ zhttp.ServerMiddleware = &Tracer{}
	_ zhttp.ClientMiddleware = &Tracer{}
	_ tracing.Tracer         = &Tracer{}
)

type Tracer struct {
	tracer trace.Tracer
	tp     *sdktrace.TracerProvider
	pg     propagation.TextMapPropagator

	addCaller bool

	serverSpanHook func(span trace.Span, w http.ResponseWriter, r *http.Request)
	clientSpanHook func(span trace.Span, w *http.Response, r *http.Request)
}

func (t *Tracer) ServerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := 1 // counter
		if v := r.Context().Value(tracing.ServerCtxKey); v != nil {
			c = v.(int) + 1
		}
		r = r.WithContext(context.WithValue(r.Context(), tracing.ServerCtxKey, c))

		span, ctx := t.spanContext(r, "server+"+r.URL.Path)
		defer span.End()
		r = r.WithContext(ctx)

		if t.addCaller {
			if ptr, file, _, ok := runtime.Caller(2); ok {
				f, _ := runtime.CallersFrames([]uintptr{ptr}).Next()
				span.SetAttributes(attribute.String("caller.file", path.Base(path.Dir(file))+"/"+path.Base(file)))
				span.SetAttributes(attribute.String("caller.func", f.Function))
				span.SetAttributes(attribute.Int("caller.line", f.Line))
			}
		}

		if c == 1 {
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
		defer span.End()
		r = r.WithContext(ctx)

		if t.addCaller {
			if ptr, file, _, ok := runtime.Caller(2); ok {
				f, _ := runtime.CallersFrames([]uintptr{ptr}).Next()
				span.SetAttributes(attribute.String("caller.file", path.Base(path.Dir(file))+"/"+path.Base(file)))
				span.SetAttributes(attribute.String("caller.func", f.Function))
				span.SetAttributes(attribute.Int("caller.line", f.Line))
			}
		}

		res, err := next.RoundTrip(r)
		if c == 1 {
			t.clientSpanHook(span, res, r)
		}
		return res, err
	})
}

// spanContext returns nre span and context for the request.
func (t *Tracer) spanContext(r *http.Request, name string) (trace.Span, context.Context) {
	var span trace.Span
	ctx := r.Context()
	if parentSpan := trace.SpanFromContext(ctx); parentSpan.SpanContext().IsValid() {
		ctx, span = t.tracer.Start(
			ctx,
			name,
			trace.WithSpanKind(trace.SpanKindInternal),
			trace.WithLinks(trace.Link{SpanContext: parentSpan.SpanContext()}),
		)
	} else {
		remoteCtx := t.pg.Extract(ctx, propagation.HeaderCarrier(r.Header))
		sc := trace.SpanFromContext(remoteCtx).SpanContext()
		if sc.IsValid() {
			ctx, span = t.tracer.Start(remoteCtx, name,
				trace.WithSpanKind(trace.SpanKindInternal),
				trace.WithLinks(trace.Link{SpanContext: sc}),
			)
		} else {
			ctx, span = t.tracer.Start(ctx, name,
				trace.WithSpanKind(trace.SpanKindServer),
			)
		}
	}
	return span, ctx
}

// Trace is the method than can be called from any types of resources.
// Callers must update their context with the returned one.
// The returned function with finishes spans must be called when finishing spans.
func (t *Tracer) Trace(ctx context.Context, name string, tags map[string]string) (spanCtx context.Context, finish func()) {
	var span trace.Span
	if parentSpan := trace.SpanFromContext(ctx); parentSpan.SpanContext().IsValid() {
		spanCtx, span = t.tracer.Start(ctx, name,
			trace.WithSpanKind(trace.SpanKindInternal),
			trace.WithLinks(trace.Link{SpanContext: parentSpan.SpanContext()}),
		)
	} else {
		spanCtx, span = t.tracer.Start(ctx, name,
			trace.WithSpanKind(trace.SpanKindServer),
		)
	}
	for k, v := range tags {
		span.SetAttributes(attribute.String(k, v))
	}
	return spanCtx, func() { span.End() }
}

// Finalize calls t.tp.Shutdown and flushes remaining data.
func (t *Tracer) Finalize(ctx context.Context) error {
	return t.tp.Shutdown(ctx)
}

func serverSpanHook(span trace.Span, w http.ResponseWriter, r *http.Request) {
	if id := zuid.FromContext(r.Context(), "context"); id != "" {
		span.SetAttributes(attribute.String("context", id))
	}
	proto := "http"
	if r.TLS != nil {
		proto = "https"
	}
	span.SetAttributes(attribute.String("http.schema", proto))
	span.SetAttributes(attribute.String("http.method", r.Method))
	span.SetAttributes(attribute.String("http.path", r.URL.Path))
	span.SetAttributes(attribute.String("http.query", r.URL.RawQuery))
	span.SetAttributes(attribute.String("http.referer", r.Referer()))
	span.SetAttributes(attribute.String("http.route", r.Pattern))
	span.SetAttributes(attribute.Int64("http.request_content_length", r.ContentLength))
	span.SetAttributes(attribute.String("net.addr", r.RemoteAddr))
	span.SetAttributes(attribute.String("net.host", r.Host))
	ww := w.(*zhttp.ResponseWrapper)
	span.SetAttributes(attribute.Int("http.status_code", ww.StatusCode()))
	span.SetAttributes(attribute.Int64("http.response_content_length", ww.Written()))
}

func clientSpanHook(span trace.Span, w *http.Response, r *http.Request) {
	if id := zuid.FromContext(r.Context(), "context"); id != "" {
		span.SetAttributes(attribute.String("context", id))
	}
	proto := "http"
	if r.TLS != nil {
		proto = "https"
	}
	span.SetAttributes(attribute.String("http.schema", proto))
	span.SetAttributes(attribute.String("http.method", r.Method))
	span.SetAttributes(attribute.String("http.path", r.URL.Path))
	span.SetAttributes(attribute.String("http.query", r.URL.RawQuery))
	span.SetAttributes(attribute.Int64("http.request_content_length", r.ContentLength))
	span.SetAttributes(attribute.String("net.addr", r.RemoteAddr))
	span.SetAttributes(attribute.String("peer.host", r.URL.Host))
	if w == nil {
		span.SetAttributes(attribute.Int("http.status_code", 0))
	} else {
		span.SetAttributes(attribute.Int("http.status_code", w.StatusCode))
		span.SetAttributes(attribute.Int64("http.response_content_length", w.ContentLength))
	}
}
