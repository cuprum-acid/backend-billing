// Package models defines the data models for the billing system.
package models

import (
	"time"

	"gorm.io/gorm"
)

// BillingPlan models a subscription tier.
type BillingPlan struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Name          string         `gorm:"uniqueIndex" json:"name"`
	Price         string         `json:"price"`
	Currency      string         `json:"currency"`
	BillingPeriod string         `json:"billingPeriod"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// Subscription models a user subscription to a plan.
type Subscription struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	UserID      string         `json:"userId"`
	PlanRef     string         `json:"planRef"` // Corresponds to BillingPlan Name
	State       string         `json:"state"`
	LastPayment *time.Time     `json:"lastPayment,omitempty"`
	NextBilling *time.Time     `json:"nextBilling,omitempty"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
