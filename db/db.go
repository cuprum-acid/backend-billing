package db

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/opentelemetry/tracing"
	
	"backend-billing/models"
)

var Conn *gorm.DB

func InitDB() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=password dbname=billing port=5432 sslmode=disable"
	}

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
