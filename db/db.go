// Package db provides database connection and initialization.
package db

import (
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"

	"backend-billing/config"
	"backend-billing/models"
)

// Conn is the global database connection instance.
var Conn *gorm.DB

// InitDB initializes the database connection and runs auto migrations.
func InitDB(cfg *config.Config) {
	dsn := cfg.GetDSN()

	// Retry connection loop for docker-compose startup
	var err error
	for i := 0; i < 5; i++ {
		Conn, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		log.Printf("Failed to connect to database... retrying in 2 seconds (%v)\n", err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Auto Migrate the schema based on structs
	err = Conn.AutoMigrate(&models.BillingPlan{}, &models.Subscription{})
	if err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	// Attach OpenTelemetry to GORM
	if err := Conn.Use(tracing.NewPlugin()); err != nil {
		log.Printf("failed to set up opentelemetry tracing for db: %v", err)
	}
}
