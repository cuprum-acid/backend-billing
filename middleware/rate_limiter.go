// Package middleware provides HTTP middleware for the billing API.
package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/time/rate"
)

var (
	rateLimitHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_rate_limit_hits_total",
			Help: "Total number of rate limit hits",
		},
		[]string{"endpoint", "ip"},
	)
	rateLimitRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_rate_limit_requests_total",
			Help: "Total number of requests tracked by rate limiter",
		},
		[]string{"endpoint"},
	)
)

// RateLimiter middleware for rate limiting requests.
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// NewRateLimiter creates a new rate limiter middleware.
// rate: requests per second
// burst: maximum burst size
func NewRateLimiter(rateLimit float64, burst int) *RateLimiter {
	return &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate.Limit(rateLimit),
		burst:    burst,
	}
}

// getLimiter returns the rate limiter for a specific IP address.
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(rl.rate, rl.burst)
		v = &visitor{limiter: limiter, lastSeen: time.Now()}
		rl.visitors[ip] = v
	} else {
		v.lastSeen = time.Now()
	}

	return v.limiter
}

// Cleanup removes old visitors that haven't been seen recently.
func (rl *RateLimiter) Cleanup(interval time.Duration) {
	go func() {
		for range time.Tick(interval) {
			rl.mu.Lock()
			for ip, v := range rl.visitors {
				if time.Since(v.lastSeen) > 3*time.Minute {
					delete(rl.visitors, ip)
				}
			}
			rl.mu.Unlock()
		}
	}()
}

// Middleware returns the HTTP middleware handler.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)
		endpoint := r.URL.Path
		limiter := rl.getLimiter(ip)

		// Track request
		rateLimitRequests.WithLabelValues(endpoint).Inc()

		if !limiter.Allow() {
			// Track rate limit hit
			rateLimitHits.WithLabelValues(endpoint, ip).Inc()
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getClientIP extracts the client IP from the request.
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		return xff
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fallback to RemoteAddr
	return r.RemoteAddr
}
