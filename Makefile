.PHONY: build test test-coverage clean install lint fmt run auth verify help ci

# Variables
BINARY_NAME=whoop-mcp
BUILD_DIR=.

# Build
build:
	@echo "Building ${BINARY_NAME}..."
	go build -o ${BUILD_DIR}/${BINARY_NAME} .

# Run tests
test:
	@echo "Running tests..."
	go test -v -race ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f ${BUILD_DIR}/${BINARY_NAME}
	rm -f coverage.out coverage.html
	go clean

# Install dependencies
install:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Lint code
lint:
	@echo "Linting code..."
	go vet ./...
	@which golangci-lint > /dev/null && golangci-lint run || echo "golangci-lint not installed, skipping"

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@which goimports > /dev/null && goimports -w . || true

# Run (for testing)
run: build
	@echo "Running ${BINARY_NAME}..."
	./${BINARY_NAME}

# OAuth authorization (auto-loads .env.local if exists)
auth:
	@echo "Starting OAuth helper..."
	@if [ -f .env.local ]; then \
		echo "Loading credentials from .env.local"; \
		. ./.env.local && go run cmd/auth/main.go; \
	else \
		echo "No .env.local found. Set WHOOP_CLIENT_ID and WHOOP_CLIENT_SECRET manually."; \
		go run cmd/auth/main.go; \
	fi

# Verify token (auto-loads .env.local if exists)
verify:
	@echo "Verifying WHOOP token..."
	@if [ -f .env.local ]; then \
		. ./.env.local && go run cmd/verify/main.go; \
	else \
		go run cmd/verify/main.go; \
	fi

# CI pipeline (run all checks)
ci: fmt lint test build
	@echo "All CI checks passed!"

# Help
help:
	@echo "Available commands:"
	@echo "  make build         - Build the project"
	@echo "  make test          - Run tests"
	@echo "  make test-coverage - Run tests with coverage report"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make install       - Install dependencies"
	@echo "  make lint          - Lint code"
	@echo "  make fmt           - Format code"
	@echo "  make run           - Build and run"
	@echo "  make auth          - Get WHOOP access token via OAuth"
	@echo "  make verify        - Verify WHOOP access token"
	@echo "  make ci            - Run all CI checks"
