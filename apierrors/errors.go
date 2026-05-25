// Package apierrors provides custom error types for the billing API.
package apierrors

import (
	"errors"
	"fmt"
	"net/http"
)

// APIError represents a structured API error.
type APIError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Details    string `json:"details,omitempty"`
	StatusCode int    `json:"-"`
}

// Error implements the error interface.
func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Common error codes
const (
	ErrValidation = "VALIDATION_ERROR"
	ErrNotFound   = "NOT_FOUND"
	ErrInternal   = "INTERNAL_ERROR"
	ErrBadRequest = "BAD_REQUEST"
	ErrDuplicate  = "DUPLICATE_ENTRY"
	ErrInvalidID  = "INVALID_ID"
	ErrDatabase   = "DATABASE_ERROR"
)

// ValidationErrors represents multiple validation errors.
type ValidationErrors struct {
	Fields map[string]string `json:"fields"`
}

// Error implements the error interface.
func (e *ValidationErrors) Error() string {
	return "validation failed"
}

// Error responses
var (
	ErrInvalidIDFormat      = &APIError{Code: ErrInvalidID, Message: "Invalid ID format", StatusCode: http.StatusBadRequest}
	ErrSubscriptionNotFound = &APIError{Code: ErrNotFound, Message: "Subscription not found", StatusCode: http.StatusNotFound}
	ErrPlanNotFound         = &APIError{Code: ErrNotFound, Message: "Plan not found", StatusCode: http.StatusNotFound}
	ErrInternalServer       = &APIError{Code: ErrInternal, Message: "Internal server error", StatusCode: http.StatusInternalServerError}
)

// NewValidationError creates a new validation error for a specific field.
func NewValidationError(field, message string) *APIError {
	return &APIError{
		Code:       ErrValidation,
		Message:    fmt.Sprintf("Field '%s' validation failed: %s", field, message),
		StatusCode: http.StatusBadRequest,
	}
}

// NewValidationErrors creates validation errors from field errors map.
func NewValidationErrors(fieldErrors map[string]string) *ValidationErrors {
	return &ValidationErrors{Fields: fieldErrors}
}

// Is returns true if target is an APIError with the same code.
func (e *APIError) Is(target error) bool {
	var apiErr *APIError
	if errors.As(target, &apiErr) {
		return e.Code == apiErr.Code
	}
	return false
}

// WrapError wraps an error with additional context.
func WrapError(err error, message string) *APIError {
	return &APIError{
		Code:       ErrInternal,
		Message:    message,
		Details:    err.Error(),
		StatusCode: http.StatusInternalServerError,
	}
}
