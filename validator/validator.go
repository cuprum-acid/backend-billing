// Package validator provides validation utilities for the billing API.
package validator

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

// pricePattern matches the same shape the kube-billing CRD enforces:
// a positive decimal with at most two fractional digits.
var pricePattern = regexp.MustCompile(`^\d+(\.\d{1,2})?$`)

var validate *validator.Validate

// userIDPattern matches the kube-billing CRD pattern for spec.userId.
var userIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// planRefPattern matches the kube-billing CRD pattern for spec.planRef
// (a DNS-1123 subdomain label, the convention shared with Kubernetes
// resource names).
var planRefPattern = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

func init() {
	validate = validator.New()

	// Register custom validation rules (ignore errors for simplicity)
	_ = validate.RegisterValidation("currency", validateCurrency)
	_ = validate.RegisterValidation("billing_period", validateBillingPeriod)
	_ = validate.RegisterValidation("price", validatePrice)
	_ = validate.RegisterValidation("user_id", validateUserID)
	_ = validate.RegisterValidation("plan_ref", validatePlanRef)
}

// validateUserID matches the same pattern the CRD enforces on
// Subscription.spec.userId.
func validateUserID(fl validator.FieldLevel) bool {
	return userIDPattern.MatchString(fl.Field().String())
}

// validatePlanRef matches the same pattern the CRD enforces on
// Subscription.spec.planRef.
func validatePlanRef(fl validator.FieldLevel) bool {
	return planRefPattern.MatchString(fl.Field().String())
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

// validatePrice validates that the price string is a non-negative decimal
// with at most two fractional digits, matching the kube-billing CRD pattern.
func validatePrice(fl validator.FieldLevel) bool {
	return pricePattern.MatchString(fl.Field().String())
}
