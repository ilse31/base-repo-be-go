# Go Clean Architecture Template

A reusable Go project template following clean architecture principles with Echo framework and Bun ORM for easy maintenance and scalability.

## 🏗️ Architecture

This project follows **Clean Architecture** principles with clear separation of concerns:

```
.
├── cmd/api/              # Application entry point
├── internal/
│   ├── domain/          # Domain entities and repository interfaces
│   ├── application/     # Business logic and use cases
│   ├── infrastructure/  # External dependencies (database, etc.)
│   └── presentation/    # HTTP handlers and routes
├── pkg/                 # Reusable packages (config, logger)
├── migrations/          # Database migrations
└── Makefile            # Build and run commands
```

### Layers

1. **Domain Layer** (`internal/domain/`)
   - Entities: Core business objects
   - Repository Interfaces: Contracts for data access

2. **Application Layer** (`internal/application/`)
   - Services: Business logic implementation
   - Use cases: Application-specific operations

3. **Infrastructure Layer** (`internal/infrastructure/`)
   - Database: Bun ORM implementation
   - External services integrations

4. **Presentation Layer** (`internal/presentation/`)
   - Handlers: HTTP request handlers
   - Router: Route definitions and middleware

## 🚀 Getting Started

### Prerequisites

- Go 1.21 or higher
- PostgreSQL database
- Redis server

### Installation

1. Clone the repository:

```bash
git clone https://github.com/ilse31/base-repo-be-go.git
cd base-repo-be-go
```

2. Install dependencies:

```bash
go mod download
```

3. Copy environment configuration:

```bash
cp .env.example .env
```

4. Update `.env` with your database and Redis credentials:

```env
SERVER_PORT=8080
ENV=development

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=mydb
DB_SSLMODE=disable

REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

JWT_SECRET=your-secret-key-change-in-production
JWT_EXPIRATION=24
```

5. Run database migrations:

```bash
# Use your preferred migration tool (e.g., migrate, golang-migrate)
# Example with migrate:
migrate -path migrations -database "postgres://user:password@localhost:5432/mydb?sslmode=disable" up
```

### Running the Application

#### Using Makefile (Development with Auto-Reload)

```bash
make dev
```

#### Using Makefile (Standard Run)

```bash
make run
```

#### Using Go directly

```bash
go run cmd/api/main.go
```

#### Build binary

```bash
make build
./bin/api
```


## 📡 API Endpoints

### Health Check

```
GET /health
```

### Authentication

```
POST   /api/v1/auth/register          Register a new user
POST   /api/v1/auth/login             Login user (sets HTTP-only cookies)
POST   /api/v1/auth/logout            Logout user (clears cookies)
POST   /api/v1/auth/refresh           Refresh access token using refresh token
POST   /api/v1/auth/forgot-password   Request password reset
POST   /api/v1/auth/reset-password    Reset password with token
GET    /api/v1/auth/me                Get current authenticated user
```

**Token Flow:**

- **Access Token**: Short-lived (1 hour by default), used for API authentication
- **Refresh Token**: Long-lived (7 days by default), used to obtain new access tokens
- Both tokens are stored in HTTP-only cookies for security
- Refresh endpoint implements token rotation for enhanced security

### Users

```
POST   /api/v1/users          Create a new user
GET    /api/v1/users/:id      Get user by ID
PUT    /api/v1/users/:id      Update user
DELETE /api/v1/users/:id      Delete user
GET    /api/v1/users          List users (with pagination)
```

### Example Requests

#### Register

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "securepassword123"
  }'
```

#### Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "securepassword123"
  }' \
  -c cookies.txt
```

#### Forgot Password

```bash
curl -X POST http://localhost:8080/api/v1/auth/forgot-password \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com"
  }'
```

#### Reset Password

```bash
curl -X POST http://localhost:8080/api/v1/auth/reset-password \
  -H "Content-Type: application/json" \
  -d '{
    "token": "reset-token-from-forgot-password",
    "new_password": "newsecurepassword123"
  }'
```

#### Get Current User

```bash
curl http://localhost:8080/api/v1/auth/me \
  -b cookies.txt
```

#### Refresh Token

```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -b cookies.txt \
  -c cookies.txt
```

This will generate new access and refresh tokens and update the cookies.

#### Create User

```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "securepassword123"
  }'
```

#### Get User

```bash
curl http://localhost:8080/api/v1/users/{user-id}
```

#### List Users

```bash
curl "http://localhost:8080/api/v1/users?limit=10&offset=0"
```

#### Update User

```bash
curl -X PUT http://localhost:8080/api/v1/users/{user-id} \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Updated"
  }'
```

#### Delete User

```bash
curl -X DELETE http://localhost:8080/api/v1/users/{user-id}
```

## 🛠️ Development

### Available Commands

```bash
make dev          # Run application with auto-reload (Air)
make run          # Run the application
make build        # Build the application
make swagger      # Generate Swagger API documentation
make test         # Run tests
make tidy         # Tidy go.mod
make clean        # Clean build artifacts
```

### 📖 API Documentation (Swagger)

Interactive Swagger UI documentation is available at:
`http://localhost:8080/swagger/index.html`

To regenerate Swagger documentation after updating API annotations:
```bash
make swagger
```

### Adding a New Entity

1. **Create Domain Entity** (`internal/domain/entity.go`)

```go
type Product struct {
    BaseEntity
    Name  string `bun:"name,notnull" json:"name"`
    Price int    `bun:"price,notnull" json:"price"`
}
```

2. **Create Repository Interface** (`internal/domain/repository.go`)

```go
type ProductRepository interface {
    Create(ctx context.Context, product *Product) error
    GetByID(ctx context.Context, id string) (*Product, error)
    // ... other methods
}
```

3. **Implement Repository** (`internal/infrastructure/database/product_repository.go`)

```go
type productRepository struct {
    db *bun.DB
}

func NewProductRepository(db *bun.DB) domain.ProductRepository {
    return &productRepository{db: db}
}

// Implement interface methods...
```

4. **Create Service** (`internal/application/product_service.go`)

```go
type ProductService struct {
    productRepo domain.ProductRepository
}

func NewProductService(productRepo domain.ProductRepository) *ProductService {
    return &ProductService{productRepo: productRepo}
}

// Implement business logic...
```

5. **Create Handler** (`internal/presentation/handler/product_handler.go`)

```go
type ProductHandler struct {
    productService *application.ProductService
}

func NewProductHandler(productService *application.ProductService) *ProductHandler {
    return &ProductHandler{productService: productService}
}

// Implement HTTP handlers...
```

6. **Register Routes** (`internal/presentation/router/router.go`)

```go
func Setup(e *echo.Echo, userHandler *handler.UserHandler, productHandler *handler.ProductHandler) {
    // ... existing routes

    products := v1.Group("/products")
    products.POST("", productHandler.CreateProduct)
    products.GET("/:id", productHandler.GetProduct)
    // ... other product routes
}
```

7. **Wire Dependencies** (`cmd/api/main.go`)

```go
productRepo := database.NewProductRepository(db.DB)
productService := application.NewProductService(productRepo)
productHandler := handler.NewProductHandler(productService)

router.Setup(e, userHandler, productHandler)
```

## 📦 Dependencies

- **Echo v4** - High performance, minimalist Go web framework
- **Bun** - SQL-first Golang ORM for PostgreSQL, MySQL, SQLite
- **Zap** - Fast, structured, leveled logging
- **Godotenv** - Load environment variables from .env file
- **Validator** - Struct validation for Go
- **UUID** - UUID generation

## 🧪 Testing

Run tests with:

```bash
make test
```

Or directly:

```bash
go test -v ./...
```

## 📝 Configuration

Configuration is managed through environment variables:

| Variable    | Description                          | Default     |
| ----------- | ------------------------------------ | ----------- |
| SERVER_PORT | Server port                          | 8080        |
| ENV         | Environment (development/production) | development |
| DB_HOST     | Database host                        | localhost   |
| DB_PORT     | Database port                        | 5432        |
| DB_USER     | Database user                        | postgres    |
| DB_PASSWORD | Database password                    | -           |
| DB_NAME     | Database name                        | mydb        |
| DB_SSLMODE  | SSL mode                             | disable     |

## 🔄 Migrations

Database migrations are stored in the `migrations/` directory. Use a migration tool like `golang-migrate` to apply them:

```bash
# Install migrate
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run migrations
migrate -path migrations -database "postgres://user:password@localhost:5432/mydb?sslmode=disable" up

# Rollback migrations
migrate -path migrations -database "postgres://user:password@localhost:5432/mydb?sslmode=disable" down
```

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License.

## 🙏 Acknowledgments

- Clean Architecture principles by Robert C. Martin
- Echo framework community
- Bun ORM contributors
