package apierrors

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	err := &APIError{
		Code:    ErrNotFound,
		Message: "Resource not found",
	}

	expected := "NOT_FOUND: Resource not found"
	if err.Error() != expected {
		t.Errorf("Expected '%s', got '%s'", expected, err.Error())
	}
}

func TestAPIError_JSON(t *testing.T) {
	err := &APIError{
		Code:       ErrValidation,
		Message:    "Invalid input",
		Details:    "Field 'email' is required",
		StatusCode: http.StatusBadRequest,
	}

	data, marshalErr := json.Marshal(err)
	if marshalErr != nil {
		t.Fatalf("Failed to marshal error: %v", marshalErr)
	}

	var result map[string]interface{}
	if unmarshalErr := json.Unmarshal(data, &result); unmarshalErr != nil {
		t.Fatalf("Failed to unmarshal error: %v", unmarshalErr)
	}

	if result["code"] != ErrValidation {
		t.Errorf("Expected code '%s', got '%v'", ErrValidation, result["code"])
	}

	if result["message"] != "Invalid input" {
		t.Errorf("Expected message 'Invalid input', got '%v'", result["message"])
	}

	if result["details"] != "Field 'email' is required" {
		t.Errorf("Expected details 'Field 'email' is required', got '%v'", result["details"])
	}
}

func TestAPIError_Is(t *testing.T) {
	err1 := &APIError{Code: ErrNotFound, Message: "Not found"}
	err2 := &APIError{Code: ErrNotFound, Message: "Different message"}
	err3 := &APIError{Code: ErrInternal, Message: "Internal error"}

	if !err1.Is(err2) {
		t.Error("Expected errors with same code to be equal")
	}

	if err1.Is(err3) {
		t.Error("Expected errors with different codes to be not equal")
	}

	// Test with non-APIError
	standardErr := errors.New("standard error")
	if err1.Is(standardErr) {
		t.Error("Expected APIError.Is to return false for non-APIError")
	}
}

func TestWrapError(t *testing.T) {
	originalErr := errors.New("database connection failed")
	wrapped := WrapError(originalErr, "Failed to connect to database")

	if wrapped.Code != ErrInternal {
		t.Errorf("Expected code '%s', got '%s'", ErrInternal, wrapped.Code)
	}

	if wrapped.Message != "Failed to connect to database" {
		t.Errorf("Expected message 'Failed to connect to database', got '%s'", wrapped.Message)
	}

	if wrapped.Details != "database connection failed" {
		t.Errorf("Expected details 'database connection failed', got '%s'", wrapped.Details)
	}

	if wrapped.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, wrapped.StatusCode)
	}
}

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("email", "is required")

	if err.Code != ErrValidation {
		t.Errorf("Expected code '%s', got '%s'", ErrValidation, err.Code)
	}

	if err.Message != "Field 'email' validation failed: is required" {
		t.Errorf("Unexpected message: %s", err.Message)
	}

	if err.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, err.StatusCode)
	}
}

func TestNewValidationErrors(t *testing.T) {
	fieldErrors := map[string]string{
		"name":  "is required",
		"email": "is invalid",
	}

	err := NewValidationErrors(fieldErrors)

	if err.Error() != "validation failed" {
		t.Errorf("Expected 'validation failed', got '%s'", err.Error())
	}

	if len(err.Fields) != 2 {
		t.Errorf("Expected 2 field errors, got %d", len(err.Fields))
	}

	if err.Fields["name"] != "is required" {
		t.Errorf("Expected name error 'is required', got '%s'", err.Fields["name"])
	}
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      *APIError
		expected string
		code     string
		status   int
	}{
		{
			name:     "ErrInvalidIDFormat",
			err:      ErrInvalidIDFormat,
			expected: "Invalid ID format",
			code:     ErrInvalidID,
			status:   http.StatusBadRequest,
		},
		{
			name:     "ErrSubscriptionNotFound",
			err:      ErrSubscriptionNotFound,
			expected: "Subscription not found",
			code:     ErrNotFound,
			status:   http.StatusNotFound,
		},
		{
			name:     "ErrPlanNotFound",
			err:      ErrPlanNotFound,
			expected: "Plan not found",
			code:     ErrNotFound,
			status:   http.StatusNotFound,
		},
		{
			name:     "ErrInternalServer",
			err:      ErrInternalServer,
			expected: "Internal server error",
			code:     ErrInternal,
			status:   http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Message != tt.expected {
				t.Errorf("Expected message '%s', got '%s'", tt.expected, tt.err.Message)
			}

			if tt.err.Code != tt.code {
				t.Errorf("Expected code '%s', got '%s'", tt.code, tt.err.Code)
			}

			if tt.err.StatusCode != tt.status {
				t.Errorf("Expected status %d, got %d", tt.status, tt.err.StatusCode)
			}
		})
	}
}

func TestAPIError_WithDetails(t *testing.T) {
	err := &APIError{
		Code:       ErrDatabase,
		Message:    "Database error",
		Details:    "connection refused",
		StatusCode: http.StatusInternalServerError,
	}

	data, _ := json.Marshal(err)
	var result map[string]interface{}
	json.Unmarshal(data, &result)

	// Details should be present
	if result["details"] != "connection refused" {
		t.Errorf("Expected details 'connection refused', got '%v'", result["details"])
	}
}

func TestAPIError_WithoutDetails(t *testing.T) {
	err := &APIError{
		Code:       ErrNotFound,
		Message:    "Not found",
		StatusCode: http.StatusNotFound,
	}

	data, _ := json.Marshal(err)
	var result map[string]interface{}
	json.Unmarshal(data, &result)

	// Details should be omitted when empty
	_, exists := result["details"]
	if exists {
		t.Error("Expected details to be omitted when empty")
	}
}
