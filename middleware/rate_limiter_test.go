package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestNewRateLimiter(t *testing.T) {
	rl := NewRateLimiter(10, 20)

	if rl == nil {
		t.Fatal("NewRateLimiter returned nil")
	}

	if rl.rate != 10 {
		t.Errorf("Expected rate 10, got %v", rl.rate)
	}

	if rl.burst != 20 {
		t.Errorf("Expected burst 20, got %v", rl.burst)
	}

	if rl.visitors == nil {
		t.Error("Expected visitors map to be initialized")
	}
}

func TestGetLimiter(t *testing.T) {
	rl := NewRateLimiter(10, 20)

	// First call should create new limiter
	limiter1 := rl.getLimiter("192.168.1.1")
	if limiter1 == nil {
		t.Fatal("Expected non-nil limiter")
	}

	// Second call should return same limiter
	limiter2 := rl.getLimiter("192.168.1.1")
	if limiter1 != limiter2 {
		t.Error("Expected same limiter for same IP")
	}

	// Different IP should get different limiter
	limiter3 := rl.getLimiter("192.168.1.2")
	if limiter1 == limiter3 {
		t.Error("Expected different limiters for different IPs")
	}
}

func TestMiddleware_Allow(t *testing.T) {
	rl := NewRateLimiter(100, 50) // High rate for testing

	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	middleware := rl.Middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("Expected handler to be called")
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestMiddleware_RateLimit(t *testing.T) {
	rl := NewRateLimiter(1, 2) // 1 req/sec, burst of 2

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := rl.Middleware(handler)

	// First 2 requests should succeed (burst)
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()
		middleware.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d: Expected status 200, got %d", i+1, w.Code)
		}
	}

	// Third request should be rate limited
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()
	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d", w.Code)
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		headers    map[string]string
		expected   string
	}{
		{
			name:       "X-Forwarded-For takes precedence",
			remoteAddr: "192.168.1.1:12345",
			headers: map[string]string{
				"X-Forwarded-For": "203.0.113.195",
				"X-Real-IP":       "203.0.113.196",
			},
			expected: "203.0.113.195",
		},
		{
			name:       "X-Real-IP when X-Forwarded-For empty",
			remoteAddr: "192.168.1.1:12345",
			headers: map[string]string{
				"X-Real-IP": "203.0.113.196",
			},
			expected: "203.0.113.196",
		},
		{
			name:       "RemoteAddr fallback",
			remoteAddr: "192.168.1.1:12345",
			headers:    map[string]string{},
			expected:   "192.168.1.1:12345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = tt.remoteAddr

			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			ip := getClientIP(req)
			if ip != tt.expected {
				t.Errorf("Expected IP %s, got %s", tt.expected, ip)
			}
		})
	}
}

func TestCleanup(t *testing.T) {
	rl := NewRateLimiter(10, 20)

	// Add a visitor
	rl.getLimiter("192.168.1.1")

	if len(rl.visitors) != 1 {
		t.Errorf("Expected 1 visitor, got %d", len(rl.visitors))
	}

	// Start cleanup with very short interval
	rl.Cleanup(10 * time.Millisecond)

	// Wait for cleanup to run
	time.Sleep(50 * time.Millisecond)

	// Visitor should be removed after 3 minutes of inactivity
	// But we can't test this without mocking time
	// This is a limitation of the current implementation
	t.Log("Cleanup goroutine started (cannot fully test without time mocking)")
}

func TestRateLimiter_Concurrency(t *testing.T) {
	rl := NewRateLimiter(1000, 100)

	var wg sync.WaitGroup
	numGoroutines := 10
	requestsPerGoroutine := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			ip := "192.168.1." + string(rune(id))
			for j := 0; j < requestsPerGoroutine; j++ {
				limiter := rl.getLimiter(ip)
				if limiter == nil {
					t.Errorf("Goroutine %d: Got nil limiter", id)
				}
			}
		}(i)
	}

	wg.Wait()

	// Should have created limiters for each goroutine
	rl.mu.RLock()
	numVisitors := len(rl.visitors)
	rl.mu.RUnlock()

	if numVisitors != numGoroutines {
		t.Errorf("Expected %d visitors, got %d", numGoroutines, numVisitors)
	}
}

func TestMiddleware_DifferentIPs(t *testing.T) {
	rl := NewRateLimiter(1, 1) // Very restrictive for testing

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := rl.Middleware(handler)

	// First IP uses its burst
	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req1.RemoteAddr = "192.168.1.1:12345"
	w1 := httptest.NewRecorder()
	middleware.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Errorf("IP 1, Request 1: Expected 200, got %d", w1.Code)
	}

	// Second IP should have its own limiter
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.RemoteAddr = "192.168.1.2:12345"
	w2 := httptest.NewRecorder()
	middleware.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("IP 2, Request 1: Expected 200, got %d", w2.Code)
	}
}
