package metrics

import (
	"context"

	"github.com/aileron-projects/go/znet/zhttp"
)

type MetricsMiddleware interface {
	zhttp.ServerMiddleware
	zhttp.ClientMiddleware
	Finalize(context.Context)
}
