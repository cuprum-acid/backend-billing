// Package handlers provides HTTP handlers for the billing API.
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"backend-billing/db"
	"backend-billing/models"
)

var (
	subsCreatedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "billing_subscriptions_created_total",
		Help: "The total number of created subscriptions",
	})
	subsCanceledTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "billing_subscriptions_canceled_total",
		Help: "The total number of canceled subscriptions",
	})
)

// GetSubscriptions returns all subscriptions.
func GetSubscriptions(w http.ResponseWriter, r *http.Request) {
	var subs []models.Subscription
	db.Conn.WithContext(r.Context()).Find(&subs)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(subs); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetSubscription returns a subscription by ID.
func GetSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid subscription id", http.StatusBadRequest)
		return
	}

	var sub models.Subscription
	if err := db.Conn.WithContext(r.Context()).First(&sub, id).Error; err != nil {
		http.Error(w, "subscription not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(sub); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

// CreateSubscription creates a new subscription.
func CreateSubscription(w http.ResponseWriter, r *http.Request) {
	var sub models.Subscription
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Simple business logic: set state and billing cycle
	sub.State = "Active"
	now := time.Now()
	next := now.AddDate(0, 1, 0) // Assume monthly
	sub.LastPayment = &now
	sub.NextBilling = &next

	if err := db.Conn.WithContext(r.Context()).Create(&sub).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Increment Prometheus counter
	subsCreatedTotal.Inc()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(sub); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

// CancelSubscription cancels an existing subscription.
func CancelSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid subscription id", http.StatusBadRequest)
		return
	}

	var sub models.Subscription
	if err := db.Conn.WithContext(r.Context()).First(&sub, id).Error; err != nil {
		http.Error(w, "subscription not found", http.StatusNotFound)
		return
	}

	sub.State = "Canceled"

	if err := db.Conn.WithContext(r.Context()).Save(&sub).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Increment Prometheus counter
	subsCanceledTotal.Inc()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(sub); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
