// Package handlers provides HTTP handlers for the billing API.
package handlers

import (
	"encoding/json"
	"net/http"

	"backend-billing/db"
	"backend-billing/models"
)

// GetPlans returns all billing plans.
func GetPlans(w http.ResponseWriter, r *http.Request) {
	var plans []models.BillingPlan
	db.Conn.WithContext(r.Context()).Find(&plans)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(plans); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

// CreatePlan creates a new billing plan.
func CreatePlan(w http.ResponseWriter, r *http.Request) {
	var plan models.BillingPlan
	if err := json.NewDecoder(r.Body).Decode(&plan); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := db.Conn.WithContext(r.Context()).Create(&plan).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(plan); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
