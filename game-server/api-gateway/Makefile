# Load the .env file if it exists
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Construct the DB_STRING dynamically
DB_STRING=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

build:
	@go build -o bin/community-builds ./cmd/

run: build
	@./bin/community-builds

# Run tests with verbose output and coverage
test:
	@go test -v ./... -cover

# Run tests with coverage output and preview in a browser
test-preview:
	@go test ./filename/ -coverprofile=coverage.out 
	@go tool cover -html=coverage.out

# Migration commands using golang-migrate
migrate-up:
	@migrate -path migrations -database "$(DB_STRING)" up

migrate-down:
	@migrate -path migrations -database "$(DB_STRING)" down

migrate-status:
	@migrate -path migrations -database "$(DB_STRING)" version

migrate-down-to:
	@if [ -z "$(VERSION)" ]; then \
		echo "Usage: make migrate-down-to VERSION=<version>"; \
		exit 1; \
	fi; \
	migrate -path migrations -database "$(DB_STRING)" down $(VERSION)

# reset everything to a specific version
migrate-fix:
	@echo "Fixing dirty migration by setting version to $(version)..."
	$(MIGRATE_COMMAND) force $(version)


# drop ALL table
migrate-reset:
	@migrate -path migrations -database "$(DB_STRING)" drop

# Create new migration files
migrate-create:
	@if [ -z "$(NAME)" ]; then \
		echo "Usage: make migrate-create NAME=<migration_name>"; \
		exit 1; \
	fi; \
	migrate create -ext sql -dir migrations -seq $(NAME)

.PHONY: run test migrate-up migrate-down migrate-status migrate-down-to migrate-reset migrate-create








