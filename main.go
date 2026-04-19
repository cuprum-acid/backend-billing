package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"

	"backend-billing/db"
	"backend-billing/handlers"
	"backend-billing/observability"
	"backend-billing/workers"
)

func main() {
	observability.InitLogger()

	ctx := context.Background()
	tp, err := observability.InitTracer(ctx, "billing-api")
	if err != nil {
		log.Fatalf("failed to init tracer: %v", err)
	}
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	log.Println("Initializing database...")
	db.InitDB()

	// Start background worker to check expiration
	log.Println("Starting background workers...")
	workers.StartSubscriptionChecker()

	r := mux.NewRouter()
	r.Use(otelmux.Middleware("billing-api"))

	// Observability
	r.Handle("/metrics", promhttp.Handler()).Methods("GET")

	// Billing Plan endpoints
	r.HandleFunc("/plans", handlers.GetPlans).Methods("GET")
	r.HandleFunc("/plans", handlers.CreatePlan).Methods("POST")

	// Subscription endpoints
	r.HandleFunc("/subscriptions", handlers.GetSubscriptions).Methods("GET")
	r.HandleFunc("/subscriptions", handlers.CreateSubscription).Methods("POST")
	r.HandleFunc("/subscriptions/{id:[0-9]+}", handlers.GetSubscription).Methods("GET")
	r.HandleFunc("/subscriptions/{id:[0-9]+}/cancel", handlers.CancelSubscription).Methods("POST")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server started on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
