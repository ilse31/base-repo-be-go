.PHONY: run dev build test tidy clean migrate-up migrate-down swagger

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
	@echo "Run migrations manually or use a migration tool like migrate"

# Run migrations down
migrate-down:
	@echo "Run migrations manually or use a migration tool like migrate"
