package main

import (
	"log"
	"net/http"
	"time"

	"github.com/aileron-projects/aileron-observability/metrics/prom"
)

func main() {
	p, _ := prom.New(&prom.Config{})
	h := p.ServerMiddleware(p) // Apply server-side API call counting.

	log.Println("server listening on localhost:8080")
	svr := &http.Server{
		Addr: ":8080",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Println(r.Method, r.URL.Path) // Simple access log.
			h.ServeHTTP(w, r)
		}),
		ReadTimeout: 10 * time.Second,
	}
	if err := svr.ListenAndServe(); err != nil {
		panic(err)
	}
}
