// Package repository provides database repository interfaces for testability.
package repository

import (
	"context"

	"gorm.io/gorm"

	"backend-billing/models"
)

// DBTX defines the interface for database operations, allowing for easy mocking in tests.
type DBTX interface {
	// Create creates a new record
	Create(value interface{}) *gorm.DB
	// First finds the first record matching the condition
	First(dest interface{}, conds ...interface{}) *gorm.DB
	// Find finds all records matching the condition
	Find(dest interface{}, conds ...interface{}) *gorm.DB
	// Save updates a record
	Save(value interface{}) *gorm.DB
	// Delete deletes a record
	Delete(value interface{}, conds ...interface{}) *gorm.DB
	// Raw executes raw SQL
	Raw(sql string, values ...interface{}) *gorm.DB
	// WithContext returns a DB with context
	WithContext(ctx context.Context) *gorm.DB
	// Error returns the error
	Error() error
	// Scan scans the result into dest
	Scan(dest interface{}) *gorm.DB
}

// PlanRepository provides database operations for BillingPlan.
type PlanRepository struct {
	db DBTX
}

// NewPlanRepository creates a new PlanRepository.
func NewPlanRepository(db DBTX) *PlanRepository {
	return &PlanRepository{db: db}
}

// GetAll returns all billing plans.
func (r *PlanRepository) GetAll(ctx context.Context) ([]models.BillingPlan, error) {
	var plans []models.BillingPlan
	err := r.db.WithContext(ctx).Find(&plans).Error
	return plans, err
}

// Create creates a new billing plan.
func (r *PlanRepository) Create(ctx context.Context, plan *models.BillingPlan) error {
	return r.db.WithContext(ctx).Create(plan).Error
}

// SubscriptionRepository provides database operations for Subscription.
type SubscriptionRepository struct {
	db DBTX
}

// NewSubscriptionRepository creates a new SubscriptionRepository.
func NewSubscriptionRepository(db DBTX) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

// GetAll returns all subscriptions.
func (r *SubscriptionRepository) GetAll(ctx context.Context) ([]models.Subscription, error) {
	var subs []models.Subscription
	err := r.db.WithContext(ctx).Find(&subs).Error
	return subs, err
}

// GetByID returns a subscription by ID.
func (r *SubscriptionRepository) GetByID(ctx context.Context, id uint) (*models.Subscription, error) {
	var sub models.Subscription
	err := r.db.WithContext(ctx).First(&sub, id).Error
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

// Create creates a new subscription.
func (r *SubscriptionRepository) Create(ctx context.Context, sub *models.Subscription) error {
	return r.db.WithContext(ctx).Create(sub).Error
}

// Save updates a subscription.
func (r *SubscriptionRepository) Save(ctx context.Context, sub *models.Subscription) error {
	return r.db.WithContext(ctx).Save(sub).Error
}

// HealthRepository provides database operations for health checks.
type HealthRepository struct {
	db DBTX
}

// NewHealthRepository creates a new HealthRepository.
func NewHealthRepository(db DBTX) *HealthRepository {
	return &HealthRepository{db: db}
}

// CheckDBConnection checks if the database is accessible.
func (r *HealthRepository) CheckDBConnection(ctx context.Context) error {
	var result int
	return r.db.WithContext(ctx).Raw("SELECT 1").Scan(&result).Error
}
