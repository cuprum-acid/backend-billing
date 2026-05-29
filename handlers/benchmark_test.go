// Package handlers provides HTTP handlers for the billing API.
//
// These benchmarks exercise components that have no I/O dependency so they
// can be re-run from any developer machine with `go test -bench=. ./handlers/`.
// Benchmarks that need a database (CreateSubscription, CreatePlan end-to-end)
// live in tests/integration and require Docker.
package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend-billing/apierrors"
	"backend-billing/validator"
)

// BenchmarkValidatePlanRequest measures validator.Struct on the real
// CreatePlanRequest type used by the CreatePlan handler. This exercises
// the custom currency / billing_period / price validators.
func BenchmarkValidatePlanRequest(b *testing.B) {
	req := CreatePlanRequest{
		Name:          "test-plan",
		Price:         "19.99",
		Currency:      "USD",
		BillingPeriod: "monthly",
	}
	v := validator.GetValidator()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := v.Struct(req); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkValidateSubscriptionRequest measures the validator on the
// CreateSubscriptionRequest type (only required/min/max tags).
func BenchmarkValidateSubscriptionRequest(b *testing.B) {
	req := CreateSubscriptionRequest{
		UserID:  "user-123",
		PlanRef: "pro",
	}
	v := validator.GetValidator()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := v.Struct(req); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkDecodeAndValidatePlan measures the JSON-decode + validate
// pipeline a handler runs before any database call. It approximates the
// non-database portion of CreatePlan's hot path.
func BenchmarkDecodeAndValidatePlan(b *testing.B) {
	body, _ := json.Marshal(map[string]string{
		"name":          "bench-plan",
		"price":         "19.99",
		"currency":      "USD",
		"billingPeriod": "monthly",
	})
	v := validator.GetValidator()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var req CreatePlanRequest
		if err := json.NewDecoder(bytes.NewReader(body)).Decode(&req); err != nil {
			b.Fatal(err)
		}
		if err := v.Struct(req); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkWriteError measures the structured-error response writer.
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

// BenchmarkHealthCheck exercises the full HealthCheck handler (no DB).
func BenchmarkHealthCheck(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()
		HealthCheck(w, req)
	}
}

// BenchmarkReadyCheck requires a database connection; the
// integration-test suite covers it end-to-end.
func BenchmarkReadyCheck(b *testing.B) {
	b.Skip("Requires database connection - see tests/integration")
}
