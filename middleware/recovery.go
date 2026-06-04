// Package middleware provides HTTP middleware for the billing API.
package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var panicRecovered = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_panics_recovered_total",
		Help: "Total number of HTTP handler panics recovered by the panic middleware",
	},
	[]string{"endpoint"},
)

// Recovery returns a middleware that turns panics inside an HTTP handler
// into a 500 response and a metric increment, instead of tearing down the
// whole server. The recovered value and stack trace are logged.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				panicRecovered.WithLabelValues(r.URL.Path).Inc()
				log.Printf("[Recovery] panic on %s %s: %v\n%s", r.Method, r.URL.Path, rec, debug.Stack())
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"code":    "INTERNAL_ERROR",
					"message": "An unexpected error occurred",
				})
			}
		}()
		next.ServeHTTP(w, r)
	})
}
