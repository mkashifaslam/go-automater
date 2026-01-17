.PHONY: help build run clean test vet fmt lint install dev prod deps tidy vendor check-fmt check-vet all

# Variables
BINARY_NAME=automater
BUILD_DIR=bin
MAIN_PATH=.
GO=go
GOFLAGS=
LDFLAGS=-ldflags "-s -w"

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

all: clean deps build ## Clean, install dependencies, and build

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

run: build ## Build and run the application
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

dev: ## Run in development mode
	@echo "Running in development mode..."
	@ENV=development $(GO) run $(MAIN_PATH)

prod: build ## Run in production mode
	@echo "Running in production mode..."
	@ENV=production ./$(BUILD_DIR)/$(BINARY_NAME)

clean: ## Remove build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@$(GO) clean
	@echo "Clean complete"

test: ## Run tests
	@echo "Running tests..."
	$(GO) test -v -race -coverprofile=coverage.out ./...
	@echo "Tests complete"

test-coverage: test ## Run tests with coverage report
	@echo "Generating coverage report..."
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

bench: ## Run benchmarks
	@echo "Running benchmarks..."
	$(GO) test -bench=. -benchmem ./...

vet: ## Run go vet
	@echo "Running go vet..."
	$(GO) vet ./...

fmt: ## Format code
	@echo "Formatting code..."
	$(GO) fmt ./...

check-fmt: ## Check if code is formatted
	@echo "Checking code formatting..."
	@test -z "$$(gofmt -l .)" || (echo "Code is not formatted. Run 'make fmt'" && exit 1)

lint: ## Run golangci-lint (requires golangci-lint installed)
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Install from https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run ./...

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GO) mod download

tidy: ## Tidy dependencies
	@echo "Tidying dependencies..."
	$(GO) mod tidy

vendor: ## Vendor dependencies
	@echo "Vendoring dependencies..."
	$(GO) mod vendor

install: ## Install the binary to $GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	$(GO) install $(GOFLAGS) $(LDFLAGS) $(MAIN_PATH)
	@echo "Installed to $$(go env GOPATH)/bin/$(BINARY_NAME)"

upgrade: ## Upgrade all dependencies
	@echo "Upgrading dependencies..."
	$(GO) get -u ./...
	$(GO) mod tidy

check: check-fmt vet test ## Run all checks (format, vet, test)

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):latest .

docker-run: docker-build ## Run Docker container
	@echo "Running Docker container..."
	docker run -p 8080:8080 $(BINARY_NAME):latest

# Development helpers
watch: ## Watch for changes and rebuild (requires entr)
	@which entr > /dev/null || (echo "entr not installed. Install with: brew install entr" && exit 1)
	@echo "Watching for changes..."
	@find . -name '*.go' | entr -r make dev

info: ## Show project information
	@echo "Project: $(BINARY_NAME)"
	@echo "Go version: $$($(GO) version)"
	@echo "Build directory: $(BUILD_DIR)"
	@echo "Main path: $(MAIN_PATH)"
