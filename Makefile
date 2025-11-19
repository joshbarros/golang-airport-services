.PHONY: help deps build build-all run test test-unit test-integration test-e2e test-coverage lint fmt clean docker-build docker-up docker-down migrate-create migrate-up migrate-down migrate-status proto openapi

# Colors for output
BLUE := \033[0;34m
GREEN := \033[0;32m
YELLOW := \033[0;33m
NC := \033[0m # No Color

# Variables
GO := go
GOFLAGS := -v
BINARY_DIR := bin
COVERAGE_DIR := coverage

# Services
SERVICES := flight-service booking-service passenger-service payment-service \
            checkin-service baggage-service notification-service auth-service \
            aircraft-service crew-service gate-service boarding-service \
            loyalty-service ancillary-service security-service immigration-service \
            ground-transport-service terminal-service lost-found-service \
            analytics-service integration-service api-gateway

help: ## Show this help message
	@echo "$(BLUE)Airport Services - Makefile Help$(NC)"
	@echo ""
	@echo "$(GREEN)Available targets:$(NC)"
	@awk 'BEGIN {FS = ":.*##"; printf "\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  $(YELLOW)%-20s$(NC) %s\n", $$1, $$2 } /^##@/ { printf "\n$(BLUE)%s$(NC)\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
	@echo ""

##@ Dependencies

deps: ## Install Go dependencies
	@echo "$(BLUE)Installing dependencies...$(NC)"
	$(GO) mod download
	$(GO) mod tidy
	@echo "$(GREEN)✓ Dependencies installed$(NC)"

deps-update: ## Update Go dependencies
	@echo "$(BLUE)Updating dependencies...$(NC)"
	$(GO) get -u ./...
	$(GO) mod tidy
	@echo "$(GREEN)✓ Dependencies updated$(NC)"

tools: ## Install development tools
	@echo "$(BLUE)Installing development tools...$(NC)"
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GO) install github.com/swaggo/swag/cmd/swag@latest
	$(GO) install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	$(GO) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	$(GO) install github.com/pressly/goose/v3/cmd/goose@latest
	@echo "$(GREEN)✓ Development tools installed$(NC)"

##@ Build

build: ## Build all services
	@echo "$(BLUE)Building all services...$(NC)"
	@mkdir -p $(BINARY_DIR)
	@for service in $(SERVICES); do \
		echo "Building $$service..."; \
		$(GO) build $(GOFLAGS) -o $(BINARY_DIR)/$$service ./cmd/$$service || exit 1; \
	done
	@echo "$(GREEN)✓ All services built successfully$(NC)"

build-%: ## Build specific service (e.g., make build-flight-service)
	@echo "$(BLUE)Building $*...$(NC)"
	@mkdir -p $(BINARY_DIR)
	$(GO) build $(GOFLAGS) -o $(BINARY_DIR)/$* ./cmd/$*
	@echo "$(GREEN)✓ $* built successfully$(NC)"

build-linux: ## Build all services for Linux
	@echo "$(BLUE)Building all services for Linux...$(NC)"
	@mkdir -p $(BINARY_DIR)/linux
	@for service in $(SERVICES); do \
		echo "Building $$service for Linux..."; \
		GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BINARY_DIR)/linux/$$service ./cmd/$$service || exit 1; \
	done
	@echo "$(GREEN)✓ All services built for Linux$(NC)"

##@ Run

run-%: ## Run specific service (e.g., make run-flight-service)
	@echo "$(BLUE)Running $*...$(NC)"
	$(GO) run ./cmd/$*

run-all: ## Run all services (not recommended for local dev)
	@echo "$(YELLOW)⚠ This will run all services. Use docker-compose instead for local development$(NC)"
	@for service in $(SERVICES); do \
		echo "Starting $$service..."; \
		$(GO) run ./cmd/$$service & \
	done

##@ Testing

test: test-unit ## Run all unit tests

test-unit: ## Run unit tests
	@echo "$(BLUE)Running unit tests...$(NC)"
	$(GO) test -v -race -timeout 30s ./...
	@echo "$(GREEN)✓ Unit tests passed$(NC)"

test-integration: ## Run integration tests
	@echo "$(BLUE)Running integration tests...$(NC)"
	$(GO) test -v -race -tags=integration -timeout 5m ./test/integration/...
	@echo "$(GREEN)✓ Integration tests passed$(NC)"

test-e2e: ## Run end-to-end tests
	@echo "$(BLUE)Running E2E tests...$(NC)"
	$(GO) test -v -tags=e2e -timeout 10m ./test/e2e/...
	@echo "$(GREEN)✓ E2E tests passed$(NC)"

test-coverage: ## Run tests with coverage report
	@echo "$(BLUE)Running tests with coverage...$(NC)"
	@mkdir -p $(COVERAGE_DIR)
	$(GO) test -v -race -coverprofile=$(COVERAGE_DIR)/coverage.out -covermode=atomic ./...
	$(GO) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	$(GO) tool cover -func=$(COVERAGE_DIR)/coverage.out
	@echo "$(GREEN)✓ Coverage report generated: $(COVERAGE_DIR)/coverage.html$(NC)"

test-bench: ## Run benchmark tests
	@echo "$(BLUE)Running benchmark tests...$(NC)"
	$(GO) test -bench=. -benchmem -run=^$$ ./...

##@ Code Quality

lint: ## Run linter
	@echo "$(BLUE)Running linter...$(NC)"
	golangci-lint run --timeout 5m ./...
	@echo "$(GREEN)✓ Linting passed$(NC)"

lint-fix: ## Run linter with auto-fix
	@echo "$(BLUE)Running linter with auto-fix...$(NC)"
	golangci-lint run --fix --timeout 5m ./...
	@echo "$(GREEN)✓ Linting completed$(NC)"

fmt: ## Format code
	@echo "$(BLUE)Formatting code...$(NC)"
	$(GO) fmt ./...
	goimports -w .
	@echo "$(GREEN)✓ Code formatted$(NC)"

vet: ## Run go vet
	@echo "$(BLUE)Running go vet...$(NC)"
	$(GO) vet ./...
	@echo "$(GREEN)✓ Go vet passed$(NC)"

check: fmt vet lint test ## Run all checks (fmt, vet, lint, test)

##@ Docker

docker-build: ## Build all Docker images
	@echo "$(BLUE)Building Docker images...$(NC)"
	@for service in $(SERVICES); do \
		echo "Building Docker image for $$service..."; \
		docker build -t airport-services/$$service:latest -f deployments/docker/$$service.Dockerfile . || exit 1; \
	done
	@echo "$(GREEN)✓ All Docker images built$(NC)"

docker-build-%: ## Build specific Docker image (e.g., make docker-build-flight-service)
	@echo "$(BLUE)Building Docker image for $*...$(NC)"
	docker build -t airport-services/$*:latest -f deployments/docker/$*.Dockerfile .
	@echo "$(GREEN)✓ Docker image built for $*$(NC)"

docker-up: ## Start all infrastructure services with Docker Compose
	@echo "$(BLUE)Starting infrastructure services...$(NC)"
	docker-compose up -d
	@echo "$(GREEN)✓ Infrastructure services started$(NC)"

docker-down: ## Stop all Docker Compose services
	@echo "$(BLUE)Stopping infrastructure services...$(NC)"
	docker-compose down
	@echo "$(GREEN)✓ Infrastructure services stopped$(NC)"

docker-logs: ## Show Docker Compose logs
	docker-compose logs -f

docker-ps: ## Show running containers
	docker-compose ps

##@ Database Migrations

migrate-create: ## Create new migration (usage: make migrate-create name=add_flights_table service=flight-service)
	@echo "$(BLUE)Creating migration: $(name) for $(service)...$(NC)"
	@mkdir -p migrations/$(service)
	goose -dir migrations/$(service) create $(name) sql
	@echo "$(GREEN)✓ Migration created$(NC)"

migrate-up: ## Run migrations (usage: make migrate-up service=flight-service)
	@echo "$(BLUE)Running migrations for $(service)...$(NC)"
	goose -dir migrations/$(service) postgres "$(DB_CONNECTION)" up
	@echo "$(GREEN)✓ Migrations applied$(NC)"

migrate-down: ## Rollback last migration (usage: make migrate-down service=flight-service)
	@echo "$(BLUE)Rolling back migration for $(service)...$(NC)"
	goose -dir migrations/$(service) postgres "$(DB_CONNECTION)" down
	@echo "$(GREEN)✓ Migration rolled back$(NC)"

migrate-status: ## Show migration status (usage: make migrate-status service=flight-service)
	@echo "$(BLUE)Migration status for $(service):$(NC)"
	goose -dir migrations/$(service) postgres "$(DB_CONNECTION)" status

migrate-reset: ## Reset all migrations (usage: make migrate-reset service=flight-service)
	@echo "$(YELLOW)⚠ Resetting all migrations for $(service)...$(NC)"
	goose -dir migrations/$(service) postgres "$(DB_CONNECTION)" reset
	@echo "$(GREEN)✓ Migrations reset$(NC)"

##@ Code Generation

proto: ## Generate Go code from protobuf files
	@echo "$(BLUE)Generating code from protobuf files...$(NC)"
	@find api/proto -name "*.proto" -exec protoc \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		{} \;
	@echo "$(GREEN)✓ Protobuf code generated$(NC)"

openapi: ## Generate OpenAPI documentation
	@echo "$(BLUE)Generating OpenAPI documentation...$(NC)"
	swag init -g cmd/api-gateway/main.go -o api/openapi
	@echo "$(GREEN)✓ OpenAPI documentation generated$(NC)"

mocks: ## Generate mocks for testing
	@echo "$(BLUE)Generating mocks...$(NC)"
	$(GO) generate ./...
	@echo "$(GREEN)✓ Mocks generated$(NC)"

generate: proto openapi mocks ## Run all code generation

##@ Kubernetes

k8s-deploy-dev: ## Deploy to development Kubernetes cluster
	@echo "$(BLUE)Deploying to development cluster...$(NC)"
	kubectl apply -k deployments/kubernetes/overlays/dev
	@echo "$(GREEN)✓ Deployed to development$(NC)"

k8s-deploy-staging: ## Deploy to staging Kubernetes cluster
	@echo "$(BLUE)Deploying to staging cluster...$(NC)"
	kubectl apply -k deployments/kubernetes/overlays/staging
	@echo "$(GREEN)✓ Deployed to staging$(NC)"

k8s-deploy-prod: ## Deploy to production Kubernetes cluster
	@echo "$(YELLOW)⚠ Deploying to production cluster...$(NC)"
	kubectl apply -k deployments/kubernetes/overlays/prod
	@echo "$(GREEN)✓ Deployed to production$(NC)"

k8s-status: ## Show Kubernetes deployment status
	@echo "$(BLUE)Kubernetes Status:$(NC)"
	kubectl get pods,svc,deploy

k8s-logs-%: ## Show logs for specific service (e.g., make k8s-logs-flight-service)
	kubectl logs -l app=$* -f

##@ Infrastructure

infra-plan: ## Plan infrastructure changes with Terraform
	@echo "$(BLUE)Planning infrastructure changes...$(NC)"
	cd deployments/terraform && terraform plan

infra-apply: ## Apply infrastructure changes with Terraform
	@echo "$(BLUE)Applying infrastructure changes...$(NC)"
	cd deployments/terraform && terraform apply

infra-destroy: ## Destroy infrastructure (use with caution!)
	@echo "$(YELLOW)⚠ Destroying infrastructure...$(NC)"
	cd deployments/terraform && terraform destroy

##@ Cleanup

clean: ## Clean build artifacts
	@echo "$(BLUE)Cleaning build artifacts...$(NC)"
	rm -rf $(BINARY_DIR)
	rm -rf $(COVERAGE_DIR)
	rm -rf vendor
	$(GO) clean -cache -testcache -modcache
	@echo "$(GREEN)✓ Cleaned$(NC)"

clean-docker: ## Remove all Docker images and volumes
	@echo "$(BLUE)Cleaning Docker resources...$(NC)"
	docker-compose down -v
	docker rmi $$(docker images -q airport-services/*) 2>/dev/null || true
	@echo "$(GREEN)✓ Docker resources cleaned$(NC)"

##@ Development

dev-setup: deps tools docker-up ## Complete development setup
	@echo "$(GREEN)✓ Development environment ready!$(NC)"
	@echo ""
	@echo "$(BLUE)Next steps:$(NC)"
	@echo "  1. Run migrations: make migrate-up service=flight-service"
	@echo "  2. Start a service: make run-flight-service"
	@echo "  3. View logs: make docker-logs"

dev-reset: clean docker-down ## Reset development environment
	@echo "$(GREEN)✓ Development environment reset$(NC)"

##@ Documentation

docs: ## Generate documentation
	@echo "$(BLUE)Generating documentation...$(NC)"
	godoc -http=:6060 &
	@echo "$(GREEN)✓ Documentation available at http://localhost:6060$(NC)"

##@ Utilities

version: ## Show project version
	@echo "Airport Services v0.1.0-alpha"

services: ## List all services
	@echo "$(BLUE)Available Services:$(NC)"
	@for service in $(SERVICES); do \
		echo "  - $$service"; \
	done

env: ## Show environment variables
	@echo "$(BLUE)Environment Variables:$(NC)"
	@env | grep -i "GO\|DOCKER\|KUBE" || true

# Default target
.DEFAULT_GOAL := help
