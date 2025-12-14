.PHONY: help build run test clean docker-build docker-run deps fmt lint

# Variables
BINARY_NAME=api-gateway
DOCKER_IMAGE=api-gateway
DOCKER_TAG=latest
PORT=8060

help: ## Display this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

deps: ## Download Go dependencies
	go mod download
	go mod tidy

build: ## Build the binary
	go build -o $(BINARY_NAME) .

run: ## Run the application locally
	go run main.go

test: ## Run tests
	go test -v ./...

test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

fmt: ## Format Go code
	go fmt ./...

lint: ## Run linters
	golangci-lint run ./...

clean: ## Clean build artifacts
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	go clean

docker-build: ## Build Docker image
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-run: ## Run Docker container
	docker run -p $(PORT):$(PORT) \
		-e ENVIRONMENT=development \
		-e PORT=$(PORT) \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

docker-clean: ## Remove Docker images
	docker rmi $(DOCKER_IMAGE):$(DOCKER_TAG)

dev: ## Run in development mode with hot reload (requires air)
	air

install-tools: ## Install development tools
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
