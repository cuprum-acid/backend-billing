# Backend Billing

A Go-based billing system for managing subscription plans and customer subscriptions.

## Features

- **Billing Plans**: Define subscription plans with pricing, currency, and billing periods
- **Subscription Management**: Create, view, and cancel customer subscriptions
- **Auto-Expiration**: Automatic detection and status update for expired subscriptions
- **Rate Limiting**: Protection against abuse (10 requests/second per IP)
- **Monitoring**: Prometheus metrics and distributed tracing support

## Quick Start

### Running with Docker Compose

```bash
docker-compose up --build
```

Services available at:
- **API**: http://localhost:8080
- **PostgreSQL**: http://localhost:5432
- **Jaeger UI**: http://localhost:16686
- **Prometheus**: http://localhost:9090

### Running Locally

```bash
go mod download

export DATABASE_URL="host=localhost user=postgres password=password dbname=billing port=5432 sslmode=disable"

go run main.go
```

## Usage

### Create a Billing Plan

```bash
curl -X POST http://localhost:8080/plans \
  -H "Content-Type: application/json" \
  -d '{
    "name": "pro",
    "price": "19.99",
    "currency": "USD",
    "billingPeriod": "monthly"
  }'
```

### Create a Subscription

```bash
curl -X POST http://localhost:8080/subscriptions \
  -H "Content-Type: application/json" \
  -d '{
    "userId": "user-123",
    "planRef": "pro"
  }'
```

### View Subscriptions

```bash
# List all
curl http://localhost:8080/subscriptions

# Get by ID
curl http://localhost:8080/subscriptions/1
```

### Cancel a Subscription

```bash
curl -X POST http://localhost:8080/subscriptions/1/cancel
```

## Development

```bash
# Build
go build -o billing-api .

# Run tests
go test -v ./...

# Lint
golangci-lint run
```

See `Makefile` for available commands.

## API Documentation

Full documentation with OpenAPI spec:

- **Swagger UI**: http://localhost:8080/swagger
- **OpenAPI Spec**: http://localhost:8080/openapi.yaml

## License

MIT License
