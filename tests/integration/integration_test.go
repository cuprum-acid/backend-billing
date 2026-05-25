// Package integration_test provides integration tests using testcontainers.
package integration_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"backend-billing/models"
)

// setupTestContainer starts a PostgreSQL test container and returns the connection string.
func setupTestContainer(t *testing.T) (testcontainers.Container, string) {
	t.Helper()

	ctx := context.Background()

	// Start PostgreSQL container
	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:15",
		tcpostgres.WithDatabase("billing_test"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	require.NoError(t, err)

	// Get connection string
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	return pgContainer, connStr
}

// setupTestDatabase creates a new DB connection and runs migrations.
func setupTestDatabase(t *testing.T, connStr string) *gorm.DB {
	t.Helper()

	dbConn, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	require.NoError(t, err)

	// Auto Migrate the schema
	err = dbConn.AutoMigrate(&models.BillingPlan{}, &models.Subscription{})
	require.NoError(t, err)

	return dbConn
}

func TestPlansAPI(t *testing.T) {
	// Setup
	pgContainer, connStr := setupTestContainer(t)
	t.Cleanup(func() {
		if err := pgContainer.Terminate(context.Background()); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	})

	dbConn := setupTestDatabase(t, connStr)

	t.Run("CreatePlan and GetPlans", func(t *testing.T) {
		// Create test plan
		plan := models.BillingPlan{
			Name:          "test-pro",
			Price:         "19.99",
			Currency:      "USD",
			BillingPeriod: "monthly",
		}

		err := dbConn.Create(&plan).Error
		require.NoError(t, err)
		assert.NotZero(t, plan.ID)

		// Get all plans
		var plans []models.BillingPlan
		err = dbConn.Find(&plans).Error
		require.NoError(t, err)
		assert.Len(t, plans, 1)
		assert.Equal(t, "test-pro", plans[0].Name)
	})
}

func TestSubscriptionsAPI(t *testing.T) {
	// Setup
	pgContainer, connStr := setupTestContainer(t)
	t.Cleanup(func() {
		if err := pgContainer.Terminate(context.Background()); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	})

	dbConn := setupTestDatabase(t, connStr)

	// First create a plan
	plan := models.BillingPlan{
		Name:          "test-basic",
		Price:         "9.99",
		Currency:      "USD",
		BillingPeriod: "monthly",
	}
	err := dbConn.Create(&plan).Error
	require.NoError(t, err)

	t.Run("CreateSubscription", func(t *testing.T) {
		// Create subscription with explicit timestamps
		now := time.Now()
		nextBilling := now.AddDate(0, 1, 0)

		sub := models.Subscription{
			UserID:      "test-user-123",
			PlanRef:     "test-basic",
			State:       "Active",
			LastPayment: &now,
			NextBilling: &nextBilling,
		}

		err = dbConn.Create(&sub).Error
		require.NoError(t, err)
		assert.NotZero(t, sub.ID)
		assert.Equal(t, "Active", sub.State)
		assert.NotNil(t, sub.LastPayment)
		assert.NotNil(t, sub.NextBilling)
	})

	t.Run("CancelSubscription", func(t *testing.T) {
		// Create subscription
		sub := models.Subscription{
			UserID:  "test-user-456",
			PlanRef: "test-basic",
			State:   "Active",
		}
		err := dbConn.Create(&sub).Error
		require.NoError(t, err)

		// Cancel subscription
		sub.State = "Canceled"
		err = dbConn.Save(&sub).Error
		require.NoError(t, err)

		// Verify state changed
		var updatedSub models.Subscription
		err = dbConn.First(&updatedSub, sub.ID).Error
		require.NoError(t, err)
		assert.Equal(t, "Canceled", updatedSub.State)
	})

	t.Run("GetSubscriptionByID", func(t *testing.T) {
		// Create subscription
		sub := models.Subscription{
			UserID:  "test-user-789",
			PlanRef: "test-basic",
			State:   "Active",
		}
		err := dbConn.Create(&sub).Error
		require.NoError(t, err)

		// Get by ID
		var retrievedSub models.Subscription
		err = dbConn.First(&retrievedSub, sub.ID).Error
		require.NoError(t, err)
		assert.Equal(t, sub.ID, retrievedSub.ID)
		assert.Equal(t, "test-user-789", retrievedSub.UserID)
	})

	t.Run("SubscriptionNotFound", func(t *testing.T) {
		var sub models.Subscription
		err := dbConn.First(&sub, 99999).Error
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "record not found")
	})
}

func TestHealthChecks(t *testing.T) {
	// Setup
	pgContainer, connStr := setupTestContainer(t)
	t.Cleanup(func() {
		if err := pgContainer.Terminate(context.Background()); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	})

	dbConn := setupTestDatabase(t, connStr)

	t.Run("DatabaseConnection", func(t *testing.T) {
		// Test database connection
		var result int
		err := dbConn.Raw("SELECT 1").Scan(&result).Error
		require.NoError(t, err)
		assert.Equal(t, 1, result)
	})
}

func TestBackgroundWorker(t *testing.T) {
	// Setup
	pgContainer, connStr := setupTestContainer(t)
	t.Cleanup(func() {
		if err := pgContainer.Terminate(context.Background()); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	})

	dbConn := setupTestDatabase(t, connStr)

	// Create a plan first
	plan := models.BillingPlan{
		Name:          "test-basic",
		Price:         "9.99",
		Currency:      "USD",
		BillingPeriod: "monthly",
	}
	err := dbConn.Create(&plan).Error
	require.NoError(t, err)

	t.Run("ExpireOldSubscriptions", func(t *testing.T) {
		// Create subscription with past billing date
		pastTime := time.Now().AddDate(0, -1, 0) // 1 month ago
		sub := models.Subscription{
			UserID:      "test-user-expire",
			PlanRef:     "test-basic",
			State:       "Active",
			LastPayment: &pastTime,
			NextBilling: &pastTime,
		}

		err := dbConn.Create(&sub).Error
		require.NoError(t, err)

		// Simulate worker logic (check for expired)
		var expiredSubs []models.Subscription
		err = dbConn.Where("state = ? AND next_billing < ?", "Active", time.Now()).Find(&expiredSubs).Error
		require.NoError(t, err)
		assert.Len(t, expiredSubs, 1)
		assert.Equal(t, sub.ID, expiredSubs[0].ID)

		// Update state to Expired
		for i := range expiredSubs {
			expiredSubs[i].State = "Expired"
			err = dbConn.Save(&expiredSubs[i]).Error
			require.NoError(t, err)
		}

		// Verify state changed
		var updatedSub models.Subscription
		err = dbConn.First(&updatedSub, sub.ID).Error
		require.NoError(t, err)
		assert.Equal(t, "Expired", updatedSub.State)
	})
}

func TestMetrics(t *testing.T) {
	// Setup
	pgContainer, connStr := setupTestContainer(t)
	t.Cleanup(func() {
		if err := pgContainer.Terminate(context.Background()); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	})

	dbConn := setupTestDatabase(t, connStr)

	// Create a plan first
	plan := models.BillingPlan{
		Name:          "test-basic",
		Price:         "9.99",
		Currency:      "USD",
		BillingPeriod: "monthly",
	}
	err := dbConn.Create(&plan).Error
	require.NoError(t, err)

	t.Run("CountSubscriptions", func(t *testing.T) {
		// Create multiple subscriptions
		for i := 0; i < 5; i++ {
			sub := models.Subscription{
				UserID:  fmt.Sprintf("test-user-%d", i),
				PlanRef: "test-basic",
				State:   "Active",
			}
			err := dbConn.Create(&sub).Error
			require.NoError(t, err)
		}

		// Count active subscriptions
		var count int64
		err := dbConn.Model(&models.Subscription{}).Where("state = ?", "Active").Count(&count).Error
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(5))
	})
}
