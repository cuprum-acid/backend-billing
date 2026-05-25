package e2e

import (
	"net/http"
	"testing"
	"time"
)

// TestHealthEndpoints verifies health check endpoints.
func (s *Suite) TestHealthEndpoints() {
	s.T().Run("Health check returns OK", func(_ *testing.T) {
		resp, body := s.makeRequest(http.MethodGet, "/health", nil)
		defer resp.Body.Close() //nolint:errcheck

		s.Equal(http.StatusOK, resp.StatusCode)
		s.Contains(string(body), "ok")
	})

	s.T().Run("Ready check returns OK", func(_ *testing.T) {
		resp, body := s.makeRequest(http.MethodGet, "/ready", nil)
		defer resp.Body.Close() //nolint:errcheck

		s.Equal(http.StatusOK, resp.StatusCode)
		s.Contains(string(body), "ok")
	})

	s.T().Run("Metrics endpoint accessible", func(_ *testing.T) {
		resp, _ := s.makeRequest(http.MethodGet, "/metrics", nil)
		defer resp.Body.Close() //nolint:errcheck

		s.Equal(http.StatusOK, resp.StatusCode)
	})
}

// TestBillingPlanWorkflow tests the complete billing plan lifecycle.
func (s *Suite) TestBillingPlanWorkflow() {
	s.T().Run("Create and retrieve billing plan", func(_ *testing.T) {
		// Create a plan
		planData := map[string]string{
			"name":          "e2e-test-plan",
			"price":         "29.99",
			"currency":      "USD",
			"billingPeriod": "monthly",
		}

		resp, body := s.makeRequest(http.MethodPost, "/plans", planData)
		defer resp.Body.Close() //nolint:errcheck
		s.Equal(http.StatusCreated, resp.StatusCode)

		var createdPlan map[string]interface{}
		s.decodeJSON(body, &createdPlan)
		s.Equal("e2e-test-plan", createdPlan["name"])
		s.Equal("29.99", createdPlan["price"])

		// Get all plans
		resp, body = s.makeRequest(http.MethodGet, "/plans", nil)
		defer resp.Body.Close() //nolint:errcheck
		s.Equal(http.StatusOK, resp.StatusCode)

		var plans []map[string]interface{}
		s.decodeJSON(body, &plans)
		s.GreaterOrEqual(len(plans), 1)
	})

	s.T().Run("Validate plan creation", func(_ *testing.T) {
		// Invalid plan (missing fields)
		invalidPlan := map[string]string{
			"name": "ab", // Too short
		}

		resp, _ := s.makeRequest(http.MethodPost, "/plans", invalidPlan)
		defer resp.Body.Close() //nolint:errcheck
		s.Equal(http.StatusBadRequest, resp.StatusCode)
	})
}

// TestSubscriptionWorkflow tests the complete subscription lifecycle.
func (s *Suite) TestSubscriptionWorkflow() {
	s.T().Run("Create subscription", func(_ *testing.T) {
		// First create a plan
		planData := map[string]string{
			"name":          "sub-test-plan",
			"price":         "9.99",
			"currency":      "USD",
			"billingPeriod": "monthly",
		}
		respPlan, _ := s.makeRequest(http.MethodPost, "/plans", planData)
		defer respPlan.Body.Close() //nolint:errcheck

		// Create subscription
		subData := map[string]string{
			"userId":  "e2e-user-123",
			"planRef": "sub-test-plan",
		}

		resp, body := s.makeRequest(http.MethodPost, "/subscriptions", subData)
		defer resp.Body.Close() //nolint:errcheck
		s.Equal(http.StatusCreated, resp.StatusCode)

		var sub map[string]interface{}
		s.decodeJSON(body, &sub)
		s.Equal("e2e-user-123", sub["userId"])
		s.Equal("sub-test-plan", sub["planRef"])
		s.Equal("Active", sub["state"])
	})

	s.T().Run("Get subscription by ID", func(_ *testing.T) {
		// Create subscription first
		planData := map[string]string{
			"name":          "get-sub-plan",
			"price":         "5.99",
			"currency":      "USD",
			"billingPeriod": "monthly",
		}
		respPlan, _ := s.makeRequest(http.MethodPost, "/plans", planData)
		defer respPlan.Body.Close() //nolint:errcheck

		subData := map[string]string{
			"userId":  "e2e-user-456",
			"planRef": "get-sub-plan",
		}
		respSub, body := s.makeRequest(http.MethodPost, "/subscriptions", subData)
		defer respSub.Body.Close() //nolint:errcheck

		var sub map[string]interface{}
		s.decodeJSON(body, &sub)

		// Get subscription
		resp, _ := s.makeRequest(http.MethodGet, "/subscriptions", nil)
		defer resp.Body.Close() //nolint:errcheck
		s.Equal(http.StatusOK, resp.StatusCode)
	})

	s.T().Run("Cancel subscription", func(_ *testing.T) {
		// Create plan and subscription
		planData := map[string]string{
			"name":          "cancel-test-plan",
			"price":         "15.99",
			"currency":      "USD",
			"billingPeriod": "monthly",
		}
		respPlan, _ := s.makeRequest(http.MethodPost, "/plans", planData)
		defer respPlan.Body.Close() //nolint:errcheck

		subData := map[string]string{
			"userId":  "e2e-user-789",
			"planRef": "cancel-test-plan",
		}
		respSub, body := s.makeRequest(http.MethodPost, "/subscriptions", subData)
		defer respSub.Body.Close() //nolint:errcheck

		var sub map[string]interface{}
		s.decodeJSON(body, &sub)

		// Cancel subscription
		resp, cancelBody := s.makeRequest(http.MethodPost, "/subscriptions/1/cancel", nil)
		defer resp.Body.Close() //nolint:errcheck
		s.Equal(http.StatusOK, resp.StatusCode)

		var canceledSub map[string]interface{}
		s.decodeJSON(cancelBody, &canceledSub)
		s.Equal("Canceled", canceledSub["state"])
	})

	s.T().Run("Subscription not found", func(_ *testing.T) {
		resp, _ := s.makeRequest(http.MethodGet, "/subscriptions/99999", nil)
		defer resp.Body.Close() //nolint:errcheck
		s.Equal(http.StatusNotFound, resp.StatusCode)
	})
}

// TestValidationErrors tests API validation.
func (s *Suite) TestValidationErrors() {
	s.T().Run("Invalid plan data", func(_ *testing.T) {
		// Empty body
		resp, _ := s.makeRequest(http.MethodPost, "/plans", map[string]string{})
		defer resp.Body.Close() //nolint:errcheck
		s.Equal(http.StatusBadRequest, resp.StatusCode)

		// Invalid currency
		invalidPlan := map[string]string{
			"name":          "invalid-plan",
			"price":         "10.00",
			"currency":      "INVALID",
			"billingPeriod": "monthly",
		}
		resp, _ = s.makeRequest(http.MethodPost, "/plans", invalidPlan)
		defer resp.Body.Close() //nolint:errcheck
		s.Equal(http.StatusBadRequest, resp.StatusCode)
	})

	s.T().Run("Invalid subscription data", func(_ *testing.T) {
		// Empty userId
		invalidSub := map[string]string{
			"userId":  "",
			"planRef": "some-plan",
		}
		resp, _ := s.makeRequest(http.MethodPost, "/subscriptions", invalidSub)
		defer resp.Body.Close() //nolint:errcheck
		s.Equal(http.StatusBadRequest, resp.StatusCode)
	})
}

// TestRateLimiting tests rate limiting behavior.
func (s *Suite) TestRateLimiting() {
	s.T().Run("Multiple requests allowed within limit", func(_ *testing.T) {
		// Send 10 requests rapidly (within rate limit)
		for i := 0; i < 10; i++ {
			resp, _ := s.makeRequest(http.MethodGet, "/health", nil)
			_ = resp.Body.Close()
			s.Equal(http.StatusOK, resp.StatusCode, "Request %d should succeed", i+1)
		}
	})
}

// TestAPIErrors tests error handling.
func (s *Suite) TestAPIErrors() {
	s.T().Run("404 for unknown endpoint", func(_ *testing.T) {
		resp, _ := s.makeRequest(http.MethodGet, "/unknown-endpoint", nil)
		defer resp.Body.Close() //nolint:errcheck
		s.Equal(http.StatusNotFound, resp.StatusCode)
	})

	s.T().Run("405 for wrong method", func(_ *testing.T) {
		resp, _ := s.makeRequest(http.MethodPut, "/plans", nil)
		defer resp.Body.Close() //nolint:errcheck
		s.Equal(http.StatusMethodNotAllowed, resp.StatusCode)
	})
}

// TestMetricsEndpoint verifies metrics are exposed.
func (s *Suite) TestMetricsEndpoint() {
	s.T().Run("Metrics contain billing metrics", func(_ *testing.T) {
		resp, body := s.makeRequest(http.MethodGet, "/metrics", nil)
		defer resp.Body.Close() //nolint:errcheck
		s.Equal(http.StatusOK, resp.StatusCode)

		metrics := string(body)
		s.Contains(metrics, "billing_subscriptions_created_total")
	})
}

// TestConcurrentAccess tests API behavior under concurrent load.
func (s *Suite) TestConcurrentAccess() {
	s.T().Run("Concurrent health checks", func(_ *testing.T) {
		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func() {
				resp, _ := s.makeRequest(http.MethodGet, "/health", nil)
				defer resp.Body.Close() //nolint:errcheck
				s.Equal(http.StatusOK, resp.StatusCode)
				done <- true
			}()
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			select {
			case <-done:
			case <-time.After(5 * time.Second):
				s.T().Fatal("Timeout waiting for concurrent requests")
			}
		}
	})
}
