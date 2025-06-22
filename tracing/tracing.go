package tracing

import (
	"context"

	"github.com/aileron-projects/go/znet/zhttp"
)

type contextKey struct{ string }

var (
	ServerCtxKey = contextKey{"server"}
	ClientCtxKey = contextKey{"client"}
)

type Tracer interface {
	Trace(ctx context.Context, name string, tags map[string]string) (spanCtx context.Context, finish func())
}

type TraceMiddleware interface {
	zhttp.ServerMiddleware
	zhttp.ClientMiddleware
	Finalize(context.Context)
	Trace(ctx context.Context, name string, tags map[string]string) (spanCtx context.Context, finish func())
}
