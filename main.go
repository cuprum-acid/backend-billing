// Package main provides the entry point for the billing API server.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"

	"backend-billing/config"
	"backend-billing/db"
	"backend-billing/handlers"
	"backend-billing/middleware"
	"backend-billing/observability"
	"backend-billing/workers"
)

func main() {
	observability.InitLogger()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	ctx := context.Background()
	tp, err := observability.InitTracer(ctx, cfg.Tracing.Service)
	if err != nil {
		log.Fatalf("failed to init tracer: %v", err)
	}
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	log.Println("Initializing database...")
	db.InitDB(cfg)

	// Start background worker to check expiration
	log.Println("Starting background workers...")
	workers.StartSubscriptionChecker()

	r := mux.NewRouter()
	r.Use(otelmux.Middleware("billing-api"))

	// Rate limiting: 10 requests per second, burst of 20
	rateLimiter := middleware.NewRateLimiter(10, 20)
	rateLimiter.Cleanup(1 * time.Minute)
	r.Use(rateLimiter.Middleware)

	// Observability
	r.Handle("/metrics", promhttp.Handler()).Methods("GET")

	// Health check endpoints
	r.HandleFunc("/health", handlers.HealthCheck).Methods("GET")
	r.HandleFunc("/ready", handlers.ReadyCheck).Methods("GET")

	// Billing Plan endpoints
	r.HandleFunc("/plans", handlers.GetPlans).Methods("GET")
	r.HandleFunc("/plans", handlers.CreatePlan).Methods("POST")

	// Subscription endpoints
	r.HandleFunc("/subscriptions", handlers.GetSubscriptions).Methods("GET")
	r.HandleFunc("/subscriptions", handlers.CreateSubscription).Methods("POST")
	r.HandleFunc("/subscriptions/{id:[0-9]+}", handlers.GetSubscription).Methods("GET")
	r.HandleFunc("/subscriptions/{id:[0-9]+}/cancel", handlers.CancelSubscription).Methods("POST")

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server started on :%s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	if err := srv.Shutdown(shutdownCtx); err != nil {
		cancel()
		log.Printf("Server forced to shutdown: %v", err)
		os.Exit(1) //nolint:gocritic // graceful shutdown failure requires immediate exit
	}
	cancel()

	log.Println("Server exited gracefully")
}
