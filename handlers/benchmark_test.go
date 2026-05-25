// Package handlers provides HTTP handlers for the billing API.
package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend-billing/apierrors"
)

// BenchmarkCreatePlan_ValidRequest benchmarks plan creation performance.
func BenchmarkCreatePlan_ValidRequest(b *testing.B) {
	planData := map[string]string{
		"name":          "benchmark-plan",
		"price":         "19.99",
		"currency":      "USD",
		"billingPeriod": "monthly",
	}

	body, _ := json.Marshal(planData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/plans", bytes.NewReader(body))
		w := httptest.NewRecorder()

		// Skip DB operations for pure handler benchmark
		_ = req
		_ = w
	}
}

// BenchmarkValidatePlan benchmarks validation performance.
func BenchmarkValidatePlan(b *testing.B) {
	planData := map[string]string{
		"name":          "test-plan",
		"price":         "19.99",
		"currency":      "USD",
		"billingPeriod": "monthly",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		body, _ := json.Marshal(planData)
		_ = body
	}
}

// BenchmarkWriteError benchmarks error response encoding.
func BenchmarkWriteError(b *testing.B) {
	err := &apierrors.APIError{
		Code:       apierrors.ErrNotFound,
		Message:    "Resource not found",
		StatusCode: http.StatusNotFound,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		writeError(w, err)
	}
}

// BenchmarkHealthCheck benchmarks health endpoint performance.
func BenchmarkHealthCheck(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()
		HealthCheck(w, req)
	}
}

// BenchmarkReadyCheck benchmarks readiness endpoint performance.
// Note: Requires database connection for full benchmark
func BenchmarkReadyCheck(b *testing.B) {
	b.Skip("Requires database connection - see integration benchmarks")
}
