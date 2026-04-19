# backend-billing

This provides the exact same high-level models as the `kube-billing` Kubernetes CRDs but built as a traditional Go web application with a PostgreSQL database.

## Architecture
- **Language**: Go
- **Framework**: Gorilla Mux for HTTP routing
- **ORM**: GORM (gorm.io/gorm)
- **Database**: PostgreSQL 15

## Getting Started

Using Docker Compose:

```bash
cd backend-billing
docker-compose up --build
```
This spins up:
- The PostgreSQL database (`:5432`)
- The Application API (`localhost:8080`)

## API Endpoints

### Billing Plans

- `GET /plans` - List all plans
- `POST /plans` - Create a plan
  ```bash
  curl -X POST http://localhost:8080/plans \
  -H "Content-Type: application/json" \
  -d '{"name": "pro", "price": "19.99", "currency": "USD", "billingPeriod": "monthly"}'
  ```

### Subscriptions

- `GET /subscriptions` - List all subscriptions
- `GET /subscriptions/{id}` - Get subscription details
- `POST /subscriptions` - Create a new subscription
  ```bash
  curl -X POST http://localhost:8080/subscriptions \
  -H "Content-Type: application/json" \
  -d '{"userId": "user-123", "planRef": "pro"}'
  ```
- `POST /subscriptions/{id}/cancel` - Cancel a subscription
  ```bash
  curl -X POST http://localhost:8080/subscriptions/1/cancel
  ```

This replaces the Reconcile loops in controllers with explicit HTTP route handlers for standard create/read/cancel CRUD operations, backed strictly relying on SQL records instead of etcd Custom Resources.
