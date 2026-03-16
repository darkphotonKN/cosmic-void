# Variables
MIGRATIONS_PATH = ./migrations

# Load the .env file if it exists
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Construct the DB_STRING dynamically
DB_STRING=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable
MIGRATE_COMMAND = migrate -path $(MIGRATIONS_PATH) -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable"

build:
	go build -o bin/stats-service cmd/server/main.go

run:
	go run cmd/server/main.go

dev:
	@air

# =============================================================================
# Testing Commands
# =============================================================================

.PHONY: test
test: ## Run all tests
	@echo "Running tests..."
	@go test -v ./... -count=1

.PHONY: test-unit
test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	@go test -v -short ./... -count=1

.PHONY: test-integration
test-integration: ## Run integration tests only
	@echo "Running integration tests..."
	@go test -v ./internal/stats -run Integration -count=1

.PHONY: test-cover
test-cover: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@go test -v ./... -coverprofile=coverage.out -covermode=atomic
	@go tool cover -func=coverage.out
	@echo ""
	@echo "✓ Coverage report saved to coverage.out"

.PHONY: test-cover-html
test-cover-html: test-cover ## Generate HTML coverage report
	@echo "Generating HTML coverage report..."
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report: coverage.html"
	@echo "Opening in browser..."
	@open coverage.html || xdg-open coverage.html || start coverage.html

.PHONY: test-race
test-race: ## Run tests with race detector
	@echo "Running tests with race detector..."
	@go test -race -v ./... -count=1

.PHONY: test-watch
test-watch: ## Watch and run tests on file changes (requires watchexec)
	@echo "Watching for changes and running tests..."
	@watchexec -e go -c -r "make test"

.PHONY: service-test
service-test: ## Run service tests only
	@echo "Running service tests..."
	@go test -v ./internal/stats -run "TestPublishMatchComplete" -count=1

# Run tests with coverage output and preview in a browser
test-preview:
	@go test ./filename/ -coverprofile=coverage.out
	@go tool cover -html=coverage.out

# Migration commands using golang-migrate
migrate-up:
	@migrate -path ./migrations -database "$(DB_STRING)" up

migrate-down:
	@migrate -path ./migrations -database "$(DB_STRING)" down

migrate-status:
	@migrate -path ./migrations -database "$(DB_STRING)" version

migrate-fix:
	@echo "Fixing dirty migration by setting version to $(version)..."
	$(MIGRATE_COMMAND) force $(version)


migrate-down-to:
	@if [ -z "$(VERSION)" ]; then \
		echo "Usage: make migrate-down-to VERSION=<version>"; \
		exit 1; \
	fi; \
	migrate -path ./migrations -database "$(DB_STRING)" down $(VERSION)

migrate-reset:
	@migrate -path ./migrations -database "$(DB_STRING)" down
	@migrate -path ./migrations -database "$(DB_STRING)" up

migrate-create:
	@if [ -z "$(NAME)" ]; then \
		echo "Usage: make migrate-create NAME=<migration_name>"; \
		exit 1; \
	fi; \
	migrate create -ext sql -dir ./migrations -seq $(NAME)

.PHONY: run test migrate-up migrate-down migrate-status migrate-down-to migrate-reset migrate-create
