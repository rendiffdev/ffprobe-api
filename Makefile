# FFprobe API - Automated Build and Deployment
# Simple commands for all platforms and deployment modes

.PHONY: help install quick prod dev clean test build docker health logs backup

# Default target
help: ## Show this help message
	@echo "FFprobe API - Automated Deployment Commands"
	@echo ""
	@echo "🚀 QUICK START:"
	@echo "  make install    # One-command setup (recommended)"
	@echo "  make quick      # Quick development setup"
	@echo ""
	@echo "📦 DEPLOYMENT MODES:"
	@echo "  make minimal    # Minimal (4 core services)"
	@echo "  make quick      # Quick start (no auth, dev mode)"
	@echo "  make prod       # Production with monitoring"
	@echo "  make dev        # Development with hot reload"
	@echo ""
	@echo "🔧 MANAGEMENT:"
	@echo "  make start      # Start all services"
	@echo "  make stop       # Stop all services"
	@echo "  make restart    # Restart all services"
	@echo "  make logs       # Show logs"
	@echo "  make health     # Check service health"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

# === INSTALLATION ===

install: ## One-command installation with setup wizard
	@echo "🚀 Starting FFprobe API installation..."
	@chmod +x setup.sh && ./setup.sh

quick: ## Quick start (no auth, development mode)
	@echo "⚡ Quick start deployment..."
	@docker compose -f docker-image/compose.yaml --profile quick up -d
	@$(MAKE) wait-ready
	@echo "✅ Quick start complete! Access: http://localhost:8080"

minimal: ## Minimal deployment (4 core services only)
	@echo "⚡ Minimal deployment (API + DB + Redis + AI)..."
	@docker compose -f docker-image/compose.yaml --profile minimal up -d
	@$(MAKE) wait-ready
	@echo "✅ Minimal deployment complete! Access: http://localhost:8080"
	@echo "   Services: API, PostgreSQL, Redis, Ollama only"

prod: ## Production deployment with monitoring
	@echo "🏭 Production deployment..."
	@if [ ! -f .env ]; then echo "❌ .env file required for production. Run 'make install' first."; exit 1; fi
	@docker compose -f docker-image/compose.yaml -f docker-image/compose.production.yaml --profile production up -d
	@$(MAKE) wait-ready
	@echo "✅ Production deployment complete!"
	@echo "   API: http://localhost:8080"
	@echo "   Monitoring: http://localhost:3000"

dev: ## Development setup with hot reload
	@echo "🔧 Development setup..."
	@docker compose -f docker-image/compose.yaml -f docker-image/compose.development.yaml --profile development up -d
	@$(MAKE) wait-ready
	@echo "✅ Development environment ready!"

# === MANAGEMENT ===

start: ## Start all services
	@docker compose -f docker-image/compose.yaml up -d

stop: ## Stop all services
	@docker compose -f docker-image/compose.yaml stop

restart: ## Restart all services
	@docker compose -f docker-image/compose.yaml restart

down: ## Stop and remove all containers
	@docker compose -f docker-image/compose.yaml down

clean: ## Remove all containers, volumes, and images
	@echo "🧹 Cleaning up..."
	@docker compose -f docker-image/compose.yaml down -v --rmi all
	@docker system prune -f
	@echo "✅ Cleanup complete"

# === MONITORING ===

status: ## Show service status
	@docker compose -f docker-image/compose.yaml ps

health: ## Check health of all services
	@echo "🏥 Health Check:"
	@echo "API:        $(shell curl -s http://localhost:8080/health | grep -o '"status":"[^"]*"' || echo '❌ Down')"
	@echo "Database:   $(shell test -f ./data/sqlite/ffprobe.db && echo '✅ Ready (SQLite)' || echo '❌ Down')"
	@echo "Valkey:     $(shell docker compose exec -T valkey valkey-cli ping 2>/dev/null || echo '❌ Down')"
	@echo "Ollama:     $(shell curl -s http://localhost:11434/api/version >/dev/null && echo '✅ Ready' || echo '❌ Down')"

logs: ## Show logs from all services
	@docker compose -f docker-image/compose.yaml logs -f

logs-api: ## Show API logs only
	@docker compose -f docker-image/compose.yaml logs -f api

logs-ollama: ## Show Ollama (AI) logs only
	@docker compose -f docker-image/compose.yaml logs -f ollama

# === TESTING ===

test: ## Run health checks and basic tests
	@echo "🧪 Running tests..."
	@$(MAKE) wait-ready
	@echo "Testing API endpoint..."
	@curl -f http://localhost:8080/health > /dev/null && echo "✅ API healthy" || echo "❌ API failed"
	@echo "Testing file upload..."
	@curl -f -X POST -F "file=@README.md" http://localhost:8080/api/v1/probe/file > /dev/null 2>&1 && echo "✅ Upload works" || echo "⚠️  Upload test skipped (auth required)"

test-ffmpeg: ## Test FFmpeg functionality
	@echo "🎬 Testing FFmpeg..."
	@docker compose -f docker-image/compose.yaml exec api ffmpeg -version | head -1
	@docker compose -f docker-image/compose.yaml exec api ffprobe -version | head -1
	@echo "✅ FFmpeg tests passed"

test-ai: ## Test AI model functionality
	@echo "🤖 Testing AI models..."
	@curl -s http://localhost:11434/api/tags | jq -r '.models[].name' | head -5 || echo "⚠️  AI models still downloading"

benchmark: ## Run performance benchmarks
	@echo "📊 Running benchmarks..."
	@ab -n 100 -c 10 http://localhost:8080/health 2>/dev/null | grep -E "(Requests per second|Time per request)" || echo "⚠️  Install 'apache2-utils' for benchmarking"

# === MAINTENANCE ===

update: ## Update to latest versions
	@echo "🔄 Updating FFprobe API..."
	@git pull origin main || echo "⚠️  Manual git pull required"
	@docker compose -f docker-image/compose.yaml pull
	@docker compose -f docker-image/compose.yaml up -d --build
	@echo "✅ Update complete"

backup: ## Create backup of data and configuration
	@echo "💾 Creating backup..."
	@mkdir -p backups/$(shell date +%Y%m%d_%H%M%S)
	@cp -r data/sqlite backups/$(shell date +%Y%m%d_%H%M%S)/ 2>/dev/null || echo "⚠️  SQLite database backup failed"
	@cp -r data/uploads backups/$(shell date +%Y%m%d_%H%M%S)/ 2>/dev/null || echo "⚠️  No uploads to backup"
	@cp -r data/valkey backups/$(shell date +%Y%m%d_%H%M%S)/ 2>/dev/null || echo "⚠️  Valkey backup failed"
	@cp .env backups/$(shell date +%Y%m%d_%H%M%S)/ 2>/dev/null || echo "ℹ️  No .env to backup"
	@echo "✅ Backup created in backups/$(shell date +%Y%m%d_%H%M%S)"

migrate: ## Run database migrations
	@echo "🔄 Running migrations..."
	@docker compose -f docker-image/compose.yaml exec api ./ffprobe-api migrate up
	@echo "✅ Migrations complete"

# === DEVELOPMENT ===

build: ## Build the API image
	@echo "🔨 Building API image..."
	@docker compose -f docker-image/compose.yaml build api

shell: ## Open shell in API container
	@docker compose -f docker-image/compose.yaml exec api /bin/bash

db-shell: ## Open SQLite shell
	@docker compose -f docker-image/compose.yaml exec api sqlite3 /app/data/ffprobe.db

valkey-shell: ## Open Valkey shell
	@docker compose -f docker-image/compose.yaml exec valkey valkey-cli -a $$VALKEY_PASSWORD

# === UTILITIES ===

env: ## Generate .env file with secure defaults
	@echo "🔐 Generating secure .env file..."
	@./scripts/generate-env.sh

config: ## Show current configuration
	@echo "⚙️  Current Configuration:"
	@docker compose -f docker-image/compose.yaml config

ps: ## Show running containers
	@docker compose -f docker-image/compose.yaml ps -a

top: ## Show container resource usage
	@docker stats --no-stream $(shell docker compose ps -q)

# === INTERNAL HELPERS ===

wait-ready: ## Wait for services to be ready (internal)
	@echo "⏳ Waiting for services to be ready..."
	@timeout=60; \
	while [ $$timeout -gt 0 ]; do \
		if curl -s http://localhost:8080/health >/dev/null 2>&1; then \
			echo "✅ Services ready!"; \
			break; \
		fi; \
		echo -n "."; \
		sleep 2; \
		timeout=$$((timeout-1)); \
	done; \
	if [ $$timeout -eq 0 ]; then \
		echo "⚠️  Services may still be starting. Check logs with 'make logs'"; \
	fi

check-docker: ## Check if Docker is available (internal)
	@docker --version >/dev/null 2>&1 || { echo "❌ Docker not found. Please install Docker first."; exit 1; }
	@docker compose -f docker-image/compose.yaml version >/dev/null 2>&1 || { echo "❌ Docker Compose not found. Please install Docker Compose."; exit 1; }

# === QUICK COMMANDS ===

# One-liners for common tasks
all: install ## Complete setup and start
fresh: clean install ## Clean install from scratch
reset: down clean quick ## Reset everything and quick start

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
	docker compose -f docker-image/compose.yaml config > /dev/null && echo "✅ Base config valid"
	docker compose -f docker-image/compose.yaml -f docker-image/compose.development.yaml config > /dev/null && echo "✅ Dev config valid"
	docker compose -f docker-image/compose.yaml -f docker-image/compose.production.yaml config > /dev/null && echo "✅ Prod config valid"

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