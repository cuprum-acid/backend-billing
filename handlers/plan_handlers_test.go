package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend-billing/apierrors"
)

func TestCreatePlan_Validation(t *testing.T) {
	// Note: Validation testing requires database connection
	// Full validation testing is done in tests/integration/
	t.Skip("Requires database connection - see integration tests")
}

func TestCreatePlan_ValidRequest(t *testing.T) {
	// Note: This test requires database connection
	// Full integration testing is done in tests/integration/
	t.Skip("Requires database connection - see integration tests")
}

func TestGetPlans_EmptyList(t *testing.T) {
	// Note: This test requires database connection
	t.Skip("Requires database connection - see integration tests")
}

func TestGetPlans_WithResults(t *testing.T) {
	// Note: This test requires database connection
	t.Skip("Requires database connection - see integration tests")
}

func TestWriteError_APIError(t *testing.T) {
	w := httptest.NewRecorder()

	err := &apierrors.APIError{
		Code:       apierrors.ErrNotFound,
		Message:    "Resource not found",
		StatusCode: http.StatusNotFound,
	}

	writeError(w, err)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}

	var resp apierrors.APIError
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Code != apierrors.ErrNotFound {
		t.Errorf("Expected code %s, got %s", apierrors.ErrNotFound, resp.Code)
	}

	if resp.Message != "Resource not found" {
		t.Errorf("Expected message 'Resource not found', got '%s'", resp.Message)
	}
}

func TestWriteError_GenericError(t *testing.T) {
	w := httptest.NewRecorder()

	err := apierrors.WrapError(testErrGeneric, "Generic error")

	writeError(w, err)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	var resp apierrors.APIError
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Code != apierrors.ErrInternal {
		t.Errorf("Expected code %s, got %s", apierrors.ErrInternal, resp.Code)
	}
}

var testErrGeneric = testErr("generic error")

type testErr string

func (t testErr) Error() string { return string(t) }

func TestIsDuplicateError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "duplicate key error",
			err:      &apierrors.APIError{Message: "duplicate key value violates unique constraint"},
			expected: true,
		},
		{
			name:     "unique constraint error",
			err:      &apierrors.APIError{Message: "unique constraint violated"},
			expected: true,
		},
		{
			name:     "generic error",
			err:      &apierrors.APIError{Message: "some other error"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isDuplicateError(tt.err)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
