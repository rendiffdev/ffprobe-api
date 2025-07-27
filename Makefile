# FFprobe API Makefile

.PHONY: help build test test-coverage test-integration clean run dev docker-build docker-run install deploy validate

# Default target
help:
	@echo "🎬 FFprobe API - Available targets:"
	@echo ""
	@echo "📦 BUILD & TEST:"
	@echo "  build            - Build the application"
	@echo "  test             - Run unit tests"
	@echo "  test-coverage    - Run tests with coverage"
	@echo "  test-integration - Run integration tests"
	@echo "  test-all         - Run all tests"
	@echo "  clean            - Clean build artifacts"
	@echo ""
	@echo "🚀 RUN & DEPLOY:"
	@echo "  run              - Run the application"
	@echo "  dev              - Run in development mode"
	@echo "  docker-build     - Build Docker image"
	@echo "  docker-run       - Run with Docker Compose"
	@echo ""
	@echo "⚙️ SETUP & INSTALL:"
	@echo "  install          - Interactive installer"
	@echo "  quick-setup      - Quick setup (3 modes)"
	@echo "  validate         - Validate configuration"
	@echo "  setup-ollama     - Setup Ollama LLM models"
	@echo ""
	@echo "🛠️ DEVELOPMENT:"
	@echo "  lint             - Run linter"
	@echo "  fmt              - Format code"
	@echo "  deploy           - Deploy to production"

# Build the application
build:
	@echo "Building ffprobe-api..."
	go build -o bin/ffprobe-api ./cmd/ffprobe-api

# Run unit tests
test:
	@echo "Running unit tests..."
	go test -v ./tests/... -run "Test[^I].*" -short

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	go test -v ./tests/... -run "TestIntegration.*"

# Run all tests
test-all:
	@echo "Running all tests..."
	go test -v ./tests/...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./tests/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	rm -rf test_storage/ test_storage_service/ test_storage_service_error/

# Run the application
run: build
	@echo "Running ffprobe-api..."
	./bin/ffprobe-api

# Run in development mode
dev:
	@echo "Running in development mode..."
	go run ./cmd/ffprobe-api

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t ffprobe-api:latest .

# Run with Docker Compose
docker-run:
	@echo "Starting with Docker Compose..."
	docker-compose up --build

# Run with Docker Compose (development)
docker-dev:
	@echo "Starting development environment with Docker Compose..."
	docker-compose -f docker-compose.yml -f docker-compose.dev.yml up --build

# Run with Docker Compose (production)
docker-prod:
	@echo "Starting production environment with Docker Compose..."
	docker-compose -f docker-compose.yml -f docker-compose.prod.yml up --build

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run ./...

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Generate mocks (if using mockery)
mocks:
	@echo "Generating mocks..."
	mockery --all --output tests/mocks

# Install development tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/vektra/mockery/v2@latest

# Database migrations (if using migrate)
migrate-up:
	@echo "Running database migrations..."
	migrate -path migrations -database "postgres://localhost/ffprobe_api?sslmode=disable" up

migrate-down:
	@echo "Rolling back database migrations..."
	migrate -path migrations -database "postgres://localhost/ffprobe_api?sslmode=disable" down

# Generate API documentation
docs:
	@echo "Generating API documentation..."
	swag init -g cmd/ffprobe-api/main.go -o docs/

# Security scan
security:
	@echo "Running security scan..."
	gosec ./...

# Benchmark tests
benchmark:
	@echo "Running benchmark tests..."
	go test -bench=. -benchmem ./tests/...

# Profile application
profile:
	@echo "Running with profiling..."
	go run -tags profile ./cmd/ffprobe-api

# Check dependencies for vulnerabilities
vuln-check:
	@echo "Checking for vulnerabilities..."
	govulncheck ./...

# Generate code coverage badge
coverage-badge: test-coverage
	@echo "Generating coverage badge..."
	go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//' > coverage.txt

# =============================================================================
# INSTALLATION & DEPLOYMENT TARGETS
# =============================================================================

# Interactive installer
install:
	@echo "🎬 Starting FFprobe API Interactive Installer..."
	./scripts/setup/install.sh

# Quick setup
quick-setup:
	@echo "⚡ Starting FFprobe API Quick Setup..."
	./scripts/setup/quick-setup.sh

# Validate configuration
validate:
	@echo "✅ Validating configuration..."
	./scripts/setup/validate-config.sh

# Deploy to production
deploy:
	@echo "🚀 Deploying to production..."
	./scripts/deployment/deploy.sh deploy production latest

# Deploy to staging
deploy-staging:
	@echo "🧪 Deploying to staging..."
	./scripts/deployment/deploy.sh deploy staging latest

# Check deployment health
health-check:
	@echo "🏥 Checking deployment health..."
	./scripts/deployment/healthcheck.sh

# Create backup
backup:
	@echo "💾 Creating backup..."
	./scripts/maintenance/backup.sh

# Setup Ollama models
setup-ollama:
	@echo "🦙 Setting up Ollama models..."
	./scripts/setup/setup-ollama.sh

# Update Docker Compose files to new syntax
docker-update:
	@echo "🐳 Updating Docker Compose syntax..."
	docker compose -f compose.yml config > /dev/null && echo "✅ Base config valid"
	docker compose -f compose.yml -f compose.dev.yml config > /dev/null && echo "✅ Dev config valid"
	docker compose -f compose.yml -f compose.prod.yml config > /dev/null && echo "✅ Prod config valid"

# Complete setup workflow
setup-complete: install validate docker-update
	@echo "🎉 Complete setup workflow finished!"
	@echo "Your FFprobe API is ready to deploy!"

# Development workflow
dev-workflow: deps install-tools fmt lint test
	@echo "🔧 Development workflow complete!"

# CI/CD pipeline simulation
ci: deps fmt lint test test-integration security vuln-check
	@echo "🚦 CI pipeline simulation complete!"

# Production readiness check
prod-ready: validate docker-update security vuln-check test-all
	@echo "🏭 Production readiness check complete!"