// Package handlers provides HTTP handlers for the billing API.
package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"backend-billing/apierrors"
	"backend-billing/db"
	"backend-billing/models"
	"backend-billing/validator"
)

// CreatePlanRequest represents the request body for creating a plan.
type CreatePlanRequest struct {
	Name          string `json:"name" validate:"required,min=3,max=50"`
	Price         string `json:"price" validate:"required,price"`
	Currency      string `json:"currency" validate:"required,currency"`
	BillingPeriod string `json:"billingPeriod" validate:"required,billing_period"`
}

// GetPlans returns all billing plans.
func GetPlans(w http.ResponseWriter, r *http.Request) {
	var plans []models.BillingPlan
	if err := db.Conn.WithContext(r.Context()).Find(&plans).Error; err != nil {
		writeError(w, apierrors.WrapError(err, "Failed to retrieve plans"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(plans); err != nil {
		writeError(w, apierrors.WrapError(err, "Failed to encode response"))
		return
	}
}

// CreatePlan creates a new billing plan.
func CreatePlan(w http.ResponseWriter, r *http.Request) {
	var req CreatePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, &apierrors.APIError{
			Code:       apierrors.ErrBadRequest,
			Message:    "Invalid request body",
			Details:    err.Error(),
			StatusCode: http.StatusBadRequest,
		})
		return
	}

	// Validate request
	if err := validator.GetValidator().Struct(req); err != nil {
		fieldErrors := make(map[string]string)
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			for _, ve := range validationErrors {
				fieldErrors[ve.Field()] = getValidationMessage(ve)
			}
		}
		writeError(w, apierrors.NewValidationErrors(fieldErrors))
		return
	}

	plan := models.BillingPlan{
		Name:          req.Name,
		Price:         req.Price,
		Currency:      req.Currency,
		BillingPeriod: req.BillingPeriod,
	}

	if err := db.Conn.WithContext(r.Context()).Create(&plan).Error; err != nil {
		if isDuplicateError(err) {
			writeError(w, &apierrors.APIError{
				Code:       apierrors.ErrDuplicate,
				Message:    "Plan with this name already exists",
				StatusCode: http.StatusConflict,
			})
			return
		}
		writeError(w, apierrors.WrapError(err, "Failed to create plan"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(plan); err != nil {
		writeError(w, apierrors.WrapError(err, "Failed to encode response"))
		return
	}
}

// getValidationMessage returns a user-friendly validation error message.
func getValidationMessage(ve validator.FieldError) string {
	switch ve.Tag() {
	case "required":
		return "This field is required"
	case "min":
		return "Value is too short (minimum " + ve.Param() + " characters)"
	case "max":
		return "Value is too long (maximum " + ve.Param() + " characters)"
	case "currency":
		return "Invalid currency code (e.g., USD, EUR, RUB)"
	case "billing_period":
		return "Invalid billing period (must be 'monthly' or 'yearly')"
	case "price":
		return "Invalid price format"
	default:
		return "Invalid value"
	}
}

// isDuplicateError checks if the error is a unique constraint violation.
func isDuplicateError(err error) bool {
	// Check for GORM duplicate key error
	return err.Error() != "" && (contains(err.Error(), "duplicate key") || contains(err.Error(), "unique constraint"))
}

// contains checks if s contains substr.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// writeError writes a structured error response.
func writeError(w http.ResponseWriter, err error) {
	var apiErr *apierrors.APIError
	if errors.As(err, &apiErr) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(apiErr.StatusCode)
		json.NewEncoder(w).Encode(apiErr)
		return
	}

	// Fallback for non-API errors
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(&apierrors.APIError{
		Code:    apierrors.ErrInternal,
		Message: "An unexpected error occurred",
	})
}
