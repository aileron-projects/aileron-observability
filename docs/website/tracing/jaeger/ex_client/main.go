package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/aileron-projects/aileron-observability/tracing/jaeger"
	"github.com/uber/jaeger-client-go/config"
)

func main() {
	t, _ := jaeger.New(&jaeger.Config{
		JaegerConfig: config.Configuration{
			ServiceName: "aileron",
			Sampler:     &config.SamplerConfig{Type: "const", Param: 1}, // Trace all.
		},
	})
	rt := http.DefaultTransport // A http client.
	rt = t.ClientMiddleware(rt) // Apply client-side API call counting.

	target, _ := url.Parse("http://httpbin.org")
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.Transport = rt

	log.Println("server listening on localhost:8080")
	svr := &http.Server{
		Addr:        ":8080",
		Handler:     proxy,
		ReadTimeout: 10 * time.Second,
	}
	if err := svr.ListenAndServe(); err != nil {
		panic(err)
	}
}
