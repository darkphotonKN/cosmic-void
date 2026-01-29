.PHONY: dev build test migrate-up migrate-down migrate-create

dev:
	@air

build:
	go build -o bin/payment-service cmd/server/main.go

test:
	@go test -v ./... -cover

migrate-up:
	@migrate -path ./migrations -database "postgres://user:password@localhost:5111/cosmic_void_payment_service_db?sslmode=disable" up

migrate-down:
	@migrate -path ./migrations -database "postgres://user:password@localhost:5111/cosmic_void_payment_service_db?sslmode=disable" down

migrate-create:
	migrate create -ext sql -dir ./migrations -seq $(NAME)
