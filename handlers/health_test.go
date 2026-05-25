package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"backend-billing/handlers"
	"backend-billing/repository/mocks"
)

func TestHealthCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "returns ok status",
			expectedStatus: http.StatusOK,
			expectedBody:   "ok",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			rr := httptest.NewRecorder()

			// Call handler
			handlers.HealthCheck(rr, req)

			// Assert status
			assert.Equal(t, tt.expectedStatus, rr.Code)

			// Assert body contains expected status
			body := rr.Body.String()
			assert.Contains(t, body, tt.expectedBody)
			assert.Contains(t, body, "timestamp")
		})
	}
}

func TestReadyCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		mockSetup      func(*mocks.MockDBTX)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "database is available",
			mockSetup: func(mock *mocks.MockDBTX) {
				// Mock will be tested through integration tests
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "ok",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: Full testing of ReadyCheck requires database integration
			// This is a placeholder for future mock-based testing
			t.Skip("Requires mock setup for DBTX interface")
		})
	}
}
