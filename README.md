# backend-billing

Traditional Go backend for billing system — Bachelor's thesis project.

This provides the exact same high-level models as the `kube-billing` Kubernetes CRDs but built as a traditional Go web application with a PostgreSQL database.

## 📚 Documentation

- **[API Documentation](docs/API.md)** — Full API reference with examples
- **[OpenAPI Spec](docs/openapi.yaml)** — OpenAPI 3.0 specification

## 🏗️ Architecture

| Component | Technology |
|-----------|------------|
| **Language** | Go 1.25 |
| **HTTP Framework** | Gorilla Mux |
| **ORM** | GORM |
| **Database** | PostgreSQL 15 |
| **Observability** | OpenTelemetry + Prometheus |
| **Rate Limiting** | Token bucket (golang.org/x/time/rate) |

## 🚀 Quick Start

### Using Docker Compose

```bash
cd backend-billing
docker-compose up --build
```

This spins up:
- **PostgreSQL**: `localhost:5432`
- **API**: `http://localhost:8080`
- **Jaeger UI**: `http://localhost:16686`
- **Prometheus**: `http://localhost:9090`

### Local Development

```bash
# Install dependencies
go mod download

# Run with PostgreSQL
export DATABASE_URL="host=localhost user=postgres password=password dbname=billing port=5432 sslmode=disable"
go run main.go
```

## 📡 API Endpoints

### Health Checks

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Basic health check |
| GET | `/ready` | Readiness check (includes DB check) |

### Billing Plans

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/plans` | List all plans |
| POST | `/plans` | Create a new plan |

**Example:**
```bash
curl -X POST http://localhost:8080/plans \
  -H "Content-Type: application/json" \
  -d '{"name": "pro", "price": "19.99", "currency": "USD", "billingPeriod": "monthly"}'
```

### Subscriptions

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/subscriptions` | List all subscriptions |
| GET | `/subscriptions/{id}` | Get subscription by ID |
| POST | `/subscriptions` | Create a new subscription |
| POST | `/subscriptions/{id}/cancel` | Cancel a subscription |

**Example:**
```bash
curl -X POST http://localhost:8080/subscriptions \
  -H "Content-Type: application/json" \
  -d '{"userId": "user-123", "planRef": "pro"}'
```

### Observability

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/metrics` | Prometheus metrics |

## 🛠️ Development

### Build

```bash
go build -o billing-api .
```

### Run Tests

```bash
# Unit + Integration tests
go test -v ./...

# Integration tests only
go test -v ./tests/integration/...
```

### Lint

```bash
golangci-lint run
```

### Make Commands

```bash
make build          # Build binary
make run            # Run locally
make test           # Run tests
make lint           # Run linter
make docker-up      # Start Docker Compose
make docker-down    # Stop Docker Compose
make help           # Show all commands
```

## 🔒 Security

### Rate Limiting

- **Limit**: 10 requests/second
- **Burst**: 20 requests
- **Algorithm**: Token bucket (per-IP)

### Input Validation

- Plan name: 3-50 characters
- Price: positive decimal string
- Currency: USD, EUR, RUB, GBP, KZT
- Billing period: monthly, yearly
- User ID: 1-255 characters

## 📊 Observability

### Metrics

- `billing_subscriptions_created_total` — Total created subscriptions
- `billing_subscriptions_canceled_total` — Total canceled subscriptions
- `http_rate_limit_requests_total` — Total requests tracked
- `http_rate_limit_hits_total` — Rate limit hits

### Tracing

OpenTelemetry tracing is enabled for:
- HTTP requests (via `otelmux` middleware)
- Database queries (via GORM tracing plugin)

Exported to Jaeger at `http://localhost:16686`.

## 🧪 Testing

### Integration Tests

Tests use `testcontainers-go` to spin up isolated PostgreSQL containers:

```bash
go test -v ./tests/integration/... -timeout 10m
```

### Test Coverage

```bash
# Generate coverage report
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## 📝 Project Structure

```
backend-billing/
├── main.go                     # Application entry point
├── config/
│   ├── config.go              # Viper-based configuration
│   └── prometheus.yml         # Prometheus scrape config
├── db/
│   └── db.go                  # Database initialization
├── handlers/
│   ├── plan_handlers.go       # HTTP handlers for plans
│   ├── subscription_handlers.go # HTTP handlers for subscriptions
│   └── health.go              # Health check handlers
├── middleware/
│   └── rate_limiter.go        # Rate limiting middleware
├── models/
│   └── models.go              # GORM models
├── observability/
│   └── setup.go               # Logging and tracing setup
├── workers/
│   └── workers.go             # Background subscription checker
├── repository/
│   ├── repository.go          # Repository interfaces
│   └── mocks/
│       └── mock_repository.go # Generated mocks
├── testhelpers/
│   └── http_helpers.go        # Test utilities
├── tests/
│   └── integration/
│       └── integration_test.go # Integration tests
├── validator/
│   └── validator.go           # Custom validation rules
├── apierrors/
│   └── errors.go              # Custom error types
└── docs/
    ├── API.md                 # API documentation
    └── openapi.yaml           # OpenAPI 3.0 specification
```
