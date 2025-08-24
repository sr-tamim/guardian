.PHONY: build clean test run dev install deps lint fmt help

# Build variables
BINARY_NAME=guardian
VERSION=0.0.1
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS=-ldflags "-X github.com/sr-tamim/guardian/pkg/version.Version=$(VERSION) -X github.com/sr-tamim/guardian/pkg/version.GitCommit=$(GIT_COMMIT) -X github.com/sr-tamim/guardian/pkg/version.BuildTime=$(BUILD_TIME)"

# Go variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

deps: ## Download dependencies
	$(GOMOD) download
	$(GOMOD) tidy

build: deps ## Build the binary
	$(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/guardian

build-linux: deps ## Build for Linux
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 ./cmd/guardian

build-all: deps ## Build for all platforms
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 ./cmd/guardian
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-arm64 ./cmd/guardian
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-amd64 ./cmd/guardian
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-arm64 ./cmd/guardian
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-amd64.exe ./cmd/guardian

clean: ## Clean build artifacts
	$(GOCLEAN)
	rm -rf bin/
	rm -rf dist/

test: ## Run tests
	$(GOTEST) -v ./...

test-coverage: ## Run tests with coverage
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

lint: ## Run linter
	golangci-lint run

fmt: ## Format code
	$(GOCMD) fmt ./...

run: build ## Build and run
	./bin/$(BINARY_NAME) monitor

dev: build ## Run in development mode
	./bin/$(BINARY_NAME) --dev monitor

daemon: build ## Run as daemon
	sudo ./bin/$(BINARY_NAME) daemon

install: build ## Install to /usr/local/bin
	sudo cp bin/$(BINARY_NAME) /usr/local/bin/
	sudo chmod +x /usr/local/bin/$(BINARY_NAME)
	@echo "Guardian installed to /usr/local/bin/$(BINARY_NAME)"

uninstall: ## Uninstall from /usr/local/bin
	sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "Guardian uninstalled"

docker-build: ## Build Docker image
	docker build -t guardian:$(VERSION) .

docker-run: docker-build ## Run in Docker
	docker run --rm -it --privileged --network host -v /var/log:/var/log guardian:$(VERSION)

# Development targets
dev-setup: ## Set up development environment
	@echo "Setting up development environment..."
	@mkdir -p bin configs test/testdata
	@echo "Development environment ready!"

simulate: build ## Simulate attacks for testing
	./bin/$(BINARY_NAME) test simulate --count 20 --delay 500

status: build ## Show status
	./bin/$(BINARY_NAME) status

list: build ## List blocked IPs
	./bin/$(BINARY_NAME) list
