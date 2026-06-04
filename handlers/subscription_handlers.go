// Package handlers provides HTTP handlers for the billing API.
package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gorm.io/gorm"

	"backend-billing/apierrors"
	"backend-billing/db"
	"backend-billing/models"
	"backend-billing/validator"
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

// CreateSubscriptionRequest represents the request body for creating a subscription.
//
// The user_id and plan_ref validators enforce the same patterns the
// kube-billing CRD applies at admission time, so both implementations
// reject the same set of bad inputs.
type CreateSubscriptionRequest struct {
	UserID  string `json:"userId"  validate:"required,min=1,max=256,user_id"`
	PlanRef string `json:"planRef" validate:"required,min=1,max=253,plan_ref"`
}

// GetSubscriptions returns all subscriptions.
func GetSubscriptions(w http.ResponseWriter, r *http.Request) {
	var subs []models.Subscription
	if err := db.Conn.WithContext(r.Context()).Find(&subs).Error; err != nil {
		writeError(w, apierrors.WrapError(err, "Failed to retrieve subscriptions"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(subs); err != nil {
		writeError(w, apierrors.WrapError(err, "Failed to encode response"))
		return
	}
}

// GetSubscription returns a subscription by ID.
func GetSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, apierrors.ErrInvalidIDFormat)
		return
	}

	var sub models.Subscription
	if err := db.Conn.WithContext(r.Context()).First(&sub, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, apierrors.ErrSubscriptionNotFound)
			return
		}
		writeError(w, apierrors.WrapError(err, "Failed to retrieve subscription"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(sub); err != nil {
		writeError(w, apierrors.WrapError(err, "Failed to encode response"))
		return
	}
}

// CreateSubscription creates a new subscription.
func CreateSubscription(w http.ResponseWriter, r *http.Request) {
	var req CreateSubscriptionRequest
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

	// Verify plan exists
	var plan models.BillingPlan
	if err := db.Conn.WithContext(r.Context()).Where("name = ?", req.PlanRef).First(&plan).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, apierrors.ErrPlanNotFound)
			return
		}
		writeError(w, apierrors.WrapError(err, "Failed to verify plan"))
		return
	}

	// Create subscription with business logic
	sub := models.Subscription{
		UserID:  req.UserID,
		PlanRef: req.PlanRef,
		State:   "Active",
	}
	now := time.Now()
	next := now.AddDate(0, 1, 0) // Default to monthly billing
	sub.LastPayment = &now
	sub.NextBilling = &next

	if err := db.Conn.WithContext(r.Context()).Create(&sub).Error; err != nil {
		writeError(w, apierrors.WrapError(err, "Failed to create subscription"))
		return
	}

	// Increment Prometheus counter
	subsCreatedTotal.Inc()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(sub); err != nil {
		writeError(w, apierrors.WrapError(err, "Failed to encode response"))
		return
	}
}

// CancelSubscription cancels an existing subscription.
func CancelSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, apierrors.ErrInvalidIDFormat)
		return
	}

	var sub models.Subscription
	if err := db.Conn.WithContext(r.Context()).First(&sub, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			writeError(w, apierrors.ErrSubscriptionNotFound)
			return
		}
		writeError(w, apierrors.WrapError(err, "Failed to retrieve subscription"))
		return
	}

	sub.State = "Canceled"
	if err := db.Conn.WithContext(r.Context()).Save(&sub).Error; err != nil {
		writeError(w, apierrors.WrapError(err, "Failed to cancel subscription"))
		return
	}

	// Increment Prometheus counter
	subsCanceledTotal.Inc()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(sub); err != nil {
		writeError(w, apierrors.WrapError(err, "Failed to encode response"))
		return
	}
}
