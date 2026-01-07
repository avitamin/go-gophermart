.PHONY: all build run test test-coverage lint fmt vet clean docker-build docker-up docker-down docker-logs migrate-up migrate-down help

# Variables
APP_NAME := gophermart
MAIN_PATH := ./cmd/gophermart
BUILD_DIR := ./bin
COVERAGE_FILE := coverage.out

# Go commands
GO := go
GOTEST := $(GO) test
GOBUILD := $(GO) build
GOVET := $(GO) vet
GOFMT := gofmt

# Docker commands
DOCKER_COMPOSE := docker compose

# Default target
all: lint test build

# Build the application
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -ldflags="-w -s" -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(APP_NAME)"

# Run the application locally
run: build
	@echo "Running $(APP_NAME)..."
	$(BUILD_DIR)/$(APP_NAME)

# Run with environment variables
run-dev:
	@echo "Running $(APP_NAME) in development mode..."
	RUN_ADDRESS=localhost:8080 \
	DATABASE_URI=postgres://gophermart:gophermart@localhost:5432/gophermart?sslmode=disable \
	LOG_LEVEL=debug \
	$(GO) run $(MAIN_PATH)

# Run all tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -race ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	$(GO) tool cover -func=$(COVERAGE_FILE)

# Show coverage in browser
test-coverage-html: test-coverage
	$(GO) tool cover -html=$(COVERAGE_FILE)

# Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	golangci-lint run ./...

# Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) -s -w .

# Run go vet
vet:
	@echo "Running go vet..."
	$(GOVET) ./...

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	$(GO) mod tidy

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f $(COVERAGE_FILE)
	@echo "Clean complete"

# Docker commands
docker-build:
	@echo "Building Docker image..."
	$(DOCKER_COMPOSE) build

docker-up:
	@echo "Starting services..."
	$(DOCKER_COMPOSE) up -d

docker-down:
	@echo "Stopping services..."
	$(DOCKER_COMPOSE) down

docker-down-v:
	@echo "Stopping services and removing volumes..."
	$(DOCKER_COMPOSE) down -v

docker-logs:
	$(DOCKER_COMPOSE) logs -f

docker-logs-app:
	$(DOCKER_COMPOSE) logs -f gophermart

docker-ps:
	$(DOCKER_COMPOSE) ps

docker-restart:
	@echo "Restarting services..."
	$(DOCKER_COMPOSE) restart

docker-rebuild: docker-down docker-build docker-up

# Database commands (requires running postgres)
db-shell:
	@echo "Connecting to database..."
	$(DOCKER_COMPOSE) exec postgres psql -U gophermart -d gophermart

# Generate mocks (requires mockgen)
mocks:
	@echo "Generating mocks..."
	@which mockgen > /dev/null || (echo "mockgen not installed. Run: go install github.com/golang/mock/mockgen@latest" && exit 1)
	mockgen -source=internal/domain/repository/user.go -destination=internal/domain/service/mocks/user_repository.go -package=mocks
	mockgen -source=internal/domain/repository/order.go -destination=internal/domain/service/mocks/order_repository.go -package=mocks
	mockgen -source=internal/domain/repository/balance.go -destination=internal/domain/service/mocks/balance_repository.go -package=mocks

# Help
help:
	@echo "Available targets:"
	@echo ""
	@echo "  Build & Run:"
	@echo "    build            - Build the application"
	@echo "    run              - Build and run the application"
	@echo "    run-dev          - Run in development mode with debug logging"
	@echo "    clean            - Remove build artifacts"
	@echo ""
	@echo "  Testing:"
	@echo "    test             - Run all tests"
	@echo "    test-coverage    - Run tests with coverage report"
	@echo "    test-coverage-html - Show coverage in browser"
	@echo ""
	@echo "  Code Quality:"
	@echo "    lint             - Run golangci-lint"
	@echo "    fmt              - Format code with gofmt"
	@echo "    vet              - Run go vet"
	@echo "    tidy             - Tidy go.mod dependencies"
	@echo ""
	@echo "  Docker:"
	@echo "    docker-build     - Build Docker image"
	@echo "    docker-up        - Start all services"
	@echo "    docker-down      - Stop all services"
	@echo "    docker-down-v    - Stop services and remove volumes"
	@echo "    docker-logs      - Follow logs for all services"
	@echo "    docker-logs-app  - Follow logs for gophermart only"
	@echo "    docker-ps        - Show running containers"
	@echo "    docker-restart   - Restart all services"
	@echo "    docker-rebuild   - Rebuild and restart services"
	@echo ""
	@echo "  Database:"
	@echo "    db-shell         - Connect to PostgreSQL shell"
	@echo ""
	@echo "  Code Generation:"
	@echo "    mocks            - Generate mock files"
