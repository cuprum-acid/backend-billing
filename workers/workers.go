package workers

import (
	"context"
	"log"
	"time"

	"backend-billing/db"
	"backend-billing/models"
)

// StartSubscriptionChecker runs in the background and periodically
// checks for subscriptions that have passed their NextBilling date
// and marks them as Expired.
func StartSubscriptionChecker() {
	go func() {
		// Run check every 1 minute (for demonstration purposes, typically daily or hourly)
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			<-ticker.C
			log.Println("[Worker] Checking for expired subscriptions...")
			checkExpiredSubscriptions()
		}
	}()
}

func checkExpiredSubscriptions() {
	now := time.Now()

	ctx := context.Background()
	var expiredSubs []models.Subscription
	if err := db.Conn.WithContext(ctx).Where("state = ? AND next_billing < ?", "Active", now).Find(&expiredSubs).Error; err != nil {
		log.Printf("[Worker] Failed to query expired subscriptions: %v", err)
		return
	}

	for _, sub := range expiredSubs {
		log.Printf("[Worker] Subscription ID %d for User %s has expired.", sub.ID, sub.UserID)
		sub.State = "Expired"
		if err := db.Conn.WithContext(ctx).Save(&sub).Error; err != nil {
			log.Printf("[Worker] Failed to update subscription %d to expired: %v", sub.ID, err)
		} else {
			log.Printf("[Worker] Successfully marked subscription %d as Expired", sub.ID)
		}
	}
}
