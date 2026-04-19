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
	subsCancelledTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "billing_subscriptions_cancelled_total",
		Help: "The total number of cancelled subscriptions",
	})
)

func GetSubscriptions(w http.ResponseWriter, r *http.Request) {
	var subs []models.Subscription
	db.Conn.WithContext(r.Context()).Find(&subs)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subs)
}

func GetSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, _ := strconv.Atoi(idStr)

	var sub models.Subscription
	if err := db.Conn.WithContext(r.Context()).First(&sub, id).Error; err != nil {
		http.Error(w, "subscription not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sub)
}

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
	json.NewEncoder(w).Encode(sub)
}

func CancelSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, _ := strconv.Atoi(idStr)

	var sub models.Subscription
	if err := db.Conn.WithContext(r.Context()).First(&sub, id).Error; err != nil {
		http.Error(w, "subscription not found", http.StatusNotFound)
		return
	}

	sub.State = "Cancelled"

	if err := db.Conn.WithContext(r.Context()).Save(&sub).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Increment Prometheus counter
	subsCancelledTotal.Inc()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(sub)
}
