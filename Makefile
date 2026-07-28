.PHONY: run dev build test tidy clean migrate-up migrate-down migrate-create swagger

# Load environment variables from .env if present
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

DB_URL ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

# Run the application
run:
	go run cmd/api/main.go

# Run the application with auto-reload (Air)
dev:
	air

# Build the application
build:
	go build -o bin/api cmd/api/main.go

# Generate Swagger API documentation
swagger:
	swag init -g cmd/api/main.go

# Run tests
test:
	go test -v ./...

# Download dependencies
tidy:
	go mod tidy

# Clean build artifacts
clean:
	rm -rf bin/ tmp/ build-errors.log

# Run migrations up
migrate-up:
	migrate -path migrations -database "$(DB_URL)" up

# Run migrations down
migrate-down:
	migrate -path migrations -database "$(DB_URL)" down

# Create a new migration file
migrate-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir migrations -seq $$name

