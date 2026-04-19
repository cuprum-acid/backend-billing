package handlers

import (
	"encoding/json"
	"net/http"

	"backend-billing/db"
	"backend-billing/models"
)

func GetPlans(w http.ResponseWriter, r *http.Request) {
	var plans []models.BillingPlan
	db.Conn.WithContext(r.Context()).Find(&plans)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plans)
}

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
	json.NewEncoder(w).Encode(plan)
}
