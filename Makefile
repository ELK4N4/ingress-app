# Makefile for ingress-app

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=ingress-app
BINARY_UNIX=$(BINARY_NAME)_unix

# Test parameters
TEST_PACKAGES=./...
COVERAGE_OUT=coverage.out
COVERAGE_HTML=coverage.html

.PHONY: all build clean test test-short test-coverage test-integration test-unit lint deps help

# Default target
all: deps lint test build

# Build the application
build:
	@echo "Building application..."
	$(GOBUILD) -o $(BINARY_NAME) -v .

# Build for Linux
build-linux:
	@echo "Building for Linux..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v .

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
	rm -f $(COVERAGE_OUT)
	rm -f $(COVERAGE_HTML)

# Install dependencies
deps:
	@echo "Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Run all tests
test:
	@echo "Running all tests..."
	$(GOTEST) -v $(TEST_PACKAGES)

# Run unit tests only (short mode, skips integration tests)
test-unit:
	@echo "Running unit tests..."
	$(GOTEST) -v -short $(TEST_PACKAGES)

# Run integration tests only
test-integration:
	@echo "Running integration tests..."
	$(GOTEST) -v -run Integration $(TEST_PACKAGES)

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=$(COVERAGE_OUT) $(TEST_PACKAGES)
	$(GOCMD) tool cover -html=$(COVERAGE_OUT) -o $(COVERAGE_HTML)
	@echo "Coverage report generated: $(COVERAGE_HTML)"

# Run tests in watch mode (requires entr)
test-watch:
	@echo "Running tests in watch mode..."
	find . -name "*.go" | entr -r make test-unit

# Run benchmarks
benchmark:
	@echo "Running benchmarks..."
	$(GOTEST) -v -bench=. -benchmem $(TEST_PACKAGES)

# Lint the code (requires golangci-lint)
lint:
	@echo "Linting code..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install with: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.50.1" && exit 1)
	golangci-lint run

# Format the code
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt $(TEST_PACKAGES)

# Vet the code
vet:
	@echo "Vetting code..."
	$(GOCMD) vet $(TEST_PACKAGES)

# Security check (requires gosec)
security:
	@echo "Running security checks..."
	@which gosec > /dev/null || (echo "gosec not found. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest" && exit 1)
	gosec $(TEST_PACKAGES)

# Start development environment
dev-up:
	@echo "Starting development environment..."
	docker-compose up -d

# Stop development environment
dev-down:
	@echo "Stopping development environment..."
	docker-compose down

# View logs from development environment
dev-logs:
	@echo "Viewing development logs..."
	docker-compose logs -f

# Run the application
run: build
	@echo "Running application..."
	./$(BINARY_NAME)

# Run with hot reload (requires air)
dev:
	@echo "Starting development server with hot reload..."
	@which air > /dev/null || (echo "air not found. Install with: go install github.com/cosmtrek/air@latest" && exit 1)
	air

# Check test coverage threshold
test-coverage-check:
	@echo "Checking test coverage threshold..."
	$(GOTEST) -coverprofile=$(COVERAGE_OUT) $(TEST_PACKAGES)
	@coverage=$$($(GOCMD) tool cover -func=$(COVERAGE_OUT) | grep total | awk '{print $$3}' | sed 's/%//'); \
	if [ $${coverage%.*} -lt 80 ]; then \
		echo "Coverage is $${coverage}%, which is below the 80% threshold"; \
		exit 1; \
	else \
		echo "Coverage is $${coverage}%, which meets the 80% threshold"; \
	fi

# Generate mocks (requires mockery)
mocks:
	@echo "Generating mocks..."
	@which mockery > /dev/null || (echo "mockery not found. Install with: go install github.com/vektra/mockery/v2@latest" && exit 1)
	mockery --all --output=./mocks

# Install development tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	go install github.com/cosmtrek/air@latest
	go install github.com/vektra/mockery/v2@latest

# Help
help:
	@echo "Available targets:"
	@echo "  build              Build the application"
	@echo "  build-linux        Build for Linux"
	@echo "  clean              Clean build artifacts"
	@echo "  deps               Install dependencies"
	@echo "  test               Run all tests"
	@echo "  test-unit          Run unit tests only"
	@echo "  test-integration   Run integration tests only"
	@echo "  test-coverage      Run tests with coverage report"
	@echo "  test-watch         Run tests in watch mode"
	@echo "  benchmark          Run benchmarks"
	@echo "  lint               Lint the code"
	@echo "  fmt                Format the code"
	@echo "  vet                Vet the code"
	@echo "  security           Run security checks"
	@echo "  dev-up             Start development environment"
	@echo "  dev-down           Stop development environment"
	@echo "  dev-logs           View development logs"
	@echo "  run                Run the application"
	@echo "  dev                Run with hot reload"
	@echo "  mocks              Generate mocks"
	@echo "  install-tools      Install development tools"
	@echo "  help               Show this help message"
