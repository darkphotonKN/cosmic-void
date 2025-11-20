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

.PHONY: build run test clean migrate-up migrate-down docker-up docker-down

# Build the application
build:
	go build -o bin/auth-service cmd/auth-service/main.go

# Run the application
run:
	go run cmd/auth-service/main.go

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/

# Database migrations
migrate-up:
	migrate -path migrations -database "postgres://user:password@localhost:5003/community_builds_auth_service_db?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgres://user:password@localhost:5003/community_builds_auth_service_db?sslmode=disable" down

# Docker commands
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

# Development with hot reload
dev:
	air

# Run tests with verbose output and coverage
test-preview:
	@go test ./filename/ -coverprofile=coverage.out 
	@go tool cover -html=coverage.out

# Migration commands using golang-migrate
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









