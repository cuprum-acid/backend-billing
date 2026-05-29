.PHONY: build run test clean lint docker-build docker-up docker-down help generate-mocks bench bench-load

# Variables
BINARY_NAME=billing-api
GO=go
GOFMT=gofmt
GOLANGCI_LINT=golangci-lint
MOCKGEN=mockgen
DOCKER=docker
DOCKER_COMPOSE=docker-compose

# Go variables
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)
LDFLAGS=-ldflags "-s -w"

# Default target
all: lint test build

## build: Build the binary
build:
	$(GO) build $(LDFLAGS) -o $(BINARY_NAME) .

## run: Run the application locally
run:
	$(GO) run main.go

## test: Run tests
test:
	$(GO) test -v -race -cover ./...

## test-coverage: Run tests with coverage report
test-coverage:
	$(GO) test -v -race -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## bench: Run microbenchmarks (no DB) and end-to-end load harness against a running server
bench:
	@echo "=== Microbenchmarks (in-process, no DB) ==="
	$(GO) test -bench=. -benchmem -benchtime=2s -count=3 -run=^$$ ./handlers/
	@echo
	@echo "=== End-to-end load harness ==="
	@echo "Requires backend running at http://localhost:8080 with a 'basic' plan."
	@mkdir -p bench/results
	@for op in create-plan create-subscription get-subscription cancel-subscription; do \
		echo "-- op=$$op --"; \
		$(GO) run ./bench -base http://localhost:8080 -n 1000 -c 50 -warmup 5s -op $$op -plan basic -out bench/results/$$op.csv || true; \
	done

## bench-load: Run only the load harness (handy when the server is already running)
bench-load:
	@mkdir -p bench/results
	$(GO) run ./bench -base http://localhost:8080 -n 5000 -c 50 -warmup 60s \
		-op create-subscription -plan basic -out bench/results/create-subscription.csv

## clean: Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

## lint: Run linter
lint:
	@echo "Running linter..."
	$(GOLANGCI_LINT) run --fix

## lint-check: Check linting without fixing
lint-check:
	$(GOLANGCI_LINT) run

## fmt: Format code
fmt:
	$(GOFMT) -s -w .

## docker-build: Build Docker images
docker-build:
	$(DOCKER_COMPOSE) build

## docker-up: Start all services
docker-up:
	$(DOCKER_COMPOSE) up --build

## docker-down: Stop all services
docker-down:
	$(DOCKER_COMPOSE) down

## docker-logs: Show logs from all containers
docker-logs:
	$(DOCKER_COMPOSE) logs -f

## docker-clean: Remove all containers and volumes
docker-clean:
	$(DOCKER_COMPOSE) down -v

## install-deps: Install development dependencies
install-deps:
	@echo "Installing golangci-lint..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Installing mockgen..."
	go install go.uber.org/mock/mockgen@latest

## generate-mocks: Generate mocks for testing
generate-mocks:
	@echo "Generating mocks..."
	$(MOCKGEN) -source=repository/repository.go -destination=repository/mocks/mock_repository.go -package=mocks

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
