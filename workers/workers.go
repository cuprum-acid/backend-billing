// Package workers provides background job processing for subscription management.
package workers

import (
	"context"
	"log/slog"
	"time"

	"backend-billing/db"
	"backend-billing/models"
)

// StartSubscriptionChecker runs in the background and periodically
// checks for subscriptions that have passed their NextBilling date
// and marks them as Expired. All log lines are emitted through slog
// so they share the JSON handler installed by observability.InitLogger.
func StartSubscriptionChecker() {
	go func() {
		// Run check every 1 minute (for demonstration purposes, typically daily or hourly)
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			<-ticker.C
			slog.Info("worker tick: checking for expired subscriptions",
				slog.String("component", "subscription_checker"))
			checkExpiredSubscriptions()
		}
	}()
}

func checkExpiredSubscriptions() {
	now := time.Now()

	ctx := context.Background()
	var expiredSubs []models.Subscription
	if err := db.Conn.WithContext(ctx).
		Where("state = ? AND next_billing < ?", "Active", now).
		Find(&expiredSubs).Error; err != nil {
		slog.Error("worker query failed",
			slog.String("component", "subscription_checker"),
			slog.Any("error", err))
		return
	}

	for _, sub := range expiredSubs {
		sub.State = "Expired"
		if err := db.Conn.WithContext(ctx).Save(&sub).Error; err != nil {
			slog.Error("worker failed to mark subscription expired",
				slog.String("component", "subscription_checker"),
				slog.Uint64("subscription_id", uint64(sub.ID)),
				slog.String("user_id", sub.UserID),
				slog.Any("error", err))
			continue
		}
		slog.Info("worker marked subscription expired",
			slog.String("component", "subscription_checker"),
			slog.Uint64("subscription_id", uint64(sub.ID)),
			slog.String("user_id", sub.UserID))
	}
}
