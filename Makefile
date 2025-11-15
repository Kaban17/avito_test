.PHONY: all build run vet lint install-lint test integration-test clean

# Default target runs static analysis
all: vet lint

# Build the application binary
build:
	@go build -o api ./cmd/api/main.go

# Run the application
run: build
	@./api

# Run all tests
test:
	@go test -v ./...

# Run integration tests
integration-test:
	@echo "Running integration tests..."
	@INTEGRATION_TESTS=1 go test -v ./tests/integration/...

# Run go vet to check for programmatic errors
vet:
	@echo "Running go vet..."
	@go vet ./...

# Run golangci-lint to check for style issues
lint:
	@echo "Running golangci-lint..."
	golangci-lint run

# Install golangci-lint
install-lint:
	@echo "Installing golangci-lint..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Database initialization
init-db: start-db
	@echo "Initializing database..."
	@./scripts/init-db.sh

# Start database using Docker Compose
start-db:
	@echo "Starting database..."
	@docker compose up -d postgres

# Stop database using Docker Compose
stop-db:
	@echo "Stopping database..."
	@docker compose down

# Run database and application
run-with-db: start-db init-db
	@make run

# Clean up database
clean-db:
	@echo "Cleaning up database..."
	@docker compose down -v

# Database migration
migrate-up:
	@echo "Running migrations..."
	@psql -h localhost -p 5433 -U postgres -d reviewer_service -f migrations/001_init.up.sql

# Clean up build artifacts
clean:
	@echo "Cleaning up..."
	@rm -f api
