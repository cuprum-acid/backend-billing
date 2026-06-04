// Package middleware provides HTTP middleware for the billing API.
package middleware

import (
	"encoding/json"
	"log/slog"
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
// whole server. The recovered value and stack trace are emitted to the
// configured slog handler using structured fields so user-controlled
// inputs (method, path) cannot forge log lines.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}
			panicRecovered.WithLabelValues(r.URL.Path).Inc()
			slog.Error("http handler panic recovered",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Any("recovered", rec),
				slog.String("stack", string(debug.Stack())),
			)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"code":    "INTERNAL_ERROR",
				"message": "An unexpected error occurred",
			})
		}()
		next.ServeHTTP(w, r)
	})
}
