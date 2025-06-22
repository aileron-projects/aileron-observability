package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/aileron-projects/aileron-observability/metrics/prom"
)

func main() {
	p, _ := prom.New(&prom.Config{})
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.URL.Path) // Simple access log.
		p.ServeHTTP(w, r)
	})

	rt := http.DefaultTransport // A http client.
	rt = p.ClientMiddleware(rt) // Apply client-side API call counting.

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
