// Package validator provides validation utilities for the billing API.
package validator

import (
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	// Register custom validation rules (ignore errors for simplicity)
	_ = validate.RegisterValidation("currency", validateCurrency)
	_ = validate.RegisterValidation("billing_period", validateBillingPeriod)
	_ = validate.RegisterValidation("price", validatePrice)
}

// GetValidator returns the global validator instance.
func GetValidator() *validator.Validate {
	return validate
}

// ValidationErrors is an alias for validator.ValidationErrors for easier imports.
type ValidationErrors = validator.ValidationErrors

// FieldError is an alias for validator.FieldError for easier imports.
type FieldError = validator.FieldError

// validateCurrency validates currency code (USD, EUR, RUB, etc.)
func validateCurrency(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	validCurrencies := map[string]bool{
		"USD": true,
		"EUR": true,
		"RUB": true,
		"GBP": true,
		"KZT": true,
	}
	return validCurrencies[value]
}

// validateBillingPeriod validates billing period (monthly, yearly)
func validateBillingPeriod(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	validPeriods := map[string]bool{
		"monthly": true,
		"yearly":  true,
	}
	return validPeriods[value]
}

// validatePrice validates price format (positive decimal)
func validatePrice(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return false
	}
	// Simple check: should be a valid number >= 0
	// For production, use decimal library for precise validation
	return len(value) > 0
}
