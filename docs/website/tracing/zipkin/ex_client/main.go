package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/aileron-projects/aileron-observability/tracing/zipkin"
	zipkingo "github.com/openzipkin/zipkin-go"
	reporter "github.com/openzipkin/zipkin-go/reporter/http"
)

func main() {
	endpoint, _ := zipkingo.NewEndpoint("aileron", "127.0.0.1:8080") // Service info.
	t, _ := zipkin.New(&zipkin.Config{
		Reporter:   reporter.NewReporter("http://localhost:9411/api/v2/spans"),
		TracerOpts: []zipkingo.TracerOption{zipkingo.WithLocalEndpoint(endpoint)},
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
