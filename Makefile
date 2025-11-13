.PHONY: all build run vet lint install-lint

# Default target runs static analysis
all: vet lint

# Build the application binary
build:
	@go build -o api ./cmd/api/main.go

# Run the application
run: build
	@./api

test:
	@go test -v ./...

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

# Clean up build artifacts
clean:
	@echo "Cleaning up..."
	@rm -f api
