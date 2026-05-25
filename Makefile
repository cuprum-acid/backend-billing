.PHONY: build run test clean lint docker-build docker-up docker-down help generate-mocks

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
