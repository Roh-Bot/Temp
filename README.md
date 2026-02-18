# Task Management REST API

A scalable RESTful Task Management Service built with Go, demonstrating clean architecture, JWT authentication, PostgreSQL persistence, and background workers.

## Features

### Core Requirements ✅
- **REST APIs**: Complete CRUD operations for tasks
- **Task Model**: ID, Title, Description, Status (pending/in_progress/completed), Timestamps
- **PostgreSQL Persistence**: Clean repository/data access layer
- **JWT Authentication**: Secure token-based authentication with expiry
- **Authorization**: Role-based access control (user/admin)
  - Users can only access their own tasks
  - Admins can access all tasks
- **Background Worker**: Auto-completes tasks after configurable minutes using goroutines and channels

### Bonus Features ✅
- **Pagination & Filtering**: List tasks with page, limit, and status filters
- **Dockerfile**: Containerized application
- **Docker Compose**: Full stack with PostgreSQL
- **Swagger/OpenAPI**: Interactive API documentation
- **Graceful Shutdown**: Context-based shutdown handling
- **Structured Logging**: Zap logger with request tracking
- **Middleware**: CORS, Recovery, Request logging
- **Input Validation**: Comprehensive request validation
- **Clean Architecture**: Separation of concerns (handlers, services, repositories)

## Architecture

```
cmd/
  api/          - HTTP handlers and routes
  task-manager/     - Main application entry point
internal/
  application/  - Business logic layer
  auth/         - JWT and encryption
  config/       - Configuration management
  database/     - Database connections
  entity/       - Domain models
  store/        - Data access layer
  validator/    - Input validation
  worker/       - Background task worker
pkg/
  logger/       - Logging utilities
  global/       - Global context management
```

## API Endpoints

### Authentication
- `POST /api/auth/register` - Register new user
- `POST /api/auth/login` - Login and get JWT token

### Tasks (Protected)
- `POST /api/tasks` - Create a task
- `GET /api/tasks` - List tasks (with pagination & filtering)
- `GET /api/tasks/{id}` - Get task by ID
- `DELETE /api/tasks/{id}` - Delete task

### Health
- `GET /api/health` - Health check

## Quick Start

### Prerequisites
- Docker & Docker Compose
- Go 1.21+ (for local development)

### Using Docker Compose

```bash
# Start the application
docker-compose up -d

# View logs
docker-compose logs -f api-go

# Stop the application
docker-compose down
```

The API will be available at `http://localhost:8000`

### Local Development

```bash
# Install dependencies
go mod download

# Run PostgreSQL
docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=admin -e POSTGRES_DB=taskmanager postgres:latest

# Run migrations
psql -h localhost -U postgres -d taskmanager -f migrations/001_init.sql

# Update config.yaml database host to localhost

# Run the application
go run cmd/task-manager/main.go
```

## Configuration

Configuration is managed via `internal/config/config.yaml`:

```yaml
Server:
  address: "0.0.0.0:8000"

Auth:
  token_ttl: 15  # minutes

Database:
  host: db
  port: 5432
  user: postgres
  password: admin
  database: taskmanager

AutoCompleteMin: 5  # Auto-complete tasks after 5 minutes
```

Environment variables can override config values.

## Usage Examples

### 1. Register a User

```bash
curl -X POST http://localhost:8000/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john",
    "email": "john@example.com",
    "password": "password123",
    "role": "user"
  }'
```

### 2. Login

```bash
curl -X POST http://localhost:8000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john",
    "password": "password123"
  }'
```

Response:
```json
{
  "Status": 1,
  "Error": "",
  "Data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

### 3. Create a Task

```bash
curl -X POST http://localhost:8000/api/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "title": "Complete assignment",
    "description": "Finish the Go task manager project"
  }'
```

### 4. List Tasks (with pagination)

```bash
curl -X GET "http://localhost:8000/api/tasks?page=1&limit=10&status=pending" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 5. Get Task by ID

```bash
curl -X GET http://localhost:8000/api/tasks/{task-id} \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 6. Delete Task

```bash
curl -X DELETE http://localhost:8000/api/tasks/{task-id} \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## Background Worker

The background worker automatically marks tasks as "completed" after a configurable time period (default: 5 minutes) if they remain in "pending" or "in_progress" status.

**Implementation Details:**
- Uses goroutines for concurrent processing
- Channel-based task queue for thread-safe communication
- Periodic scanning (every 30 seconds) for eligible tasks
- Non-blocking queue operations to prevent deadlocks
- Graceful shutdown with context cancellation

## Authentication & Authorization

### JWT Token Structure
```json
{
  "user_id": "uuid",
  "username": "john",
  "role": "user",
  "exp": 1234567890,
  "iss": "TaskManager",
  "aud": "TaskApp"
}
```

### Authorization Rules
- **Users**: Can only create, view, and delete their own tasks
- **Admins**: Can view and delete all tasks across all users

## API Documentation

Swagger documentation is available at:
```
http://localhost:8000/swagger/index.html
```

To regenerate Swagger docs:
```bash
swag init -g cmd/task-manager/main.go
```

## Database Schema

### Users Table
```sql
CREATE TABLE users (
    id VARCHAR(36) PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL CHECK (role IN ('user', 'admin'))
);
```

### Tasks Table
```sql
CREATE TABLE tasks (
    id VARCHAR(36) PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    status VARCHAR(50) NOT NULL CHECK (status IN ('pending', 'in_progress', 'completed')),
    user_id VARCHAR(36) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
```

## Error Handling

The API returns consistent JSON error responses:

```json
{
  "Status": 0,
  "Error": "error message",
  "Data": null
}
```

HTTP Status Codes:
- `200` - Success
- `201` - Created
- `204` - No Content
- `400` - Bad Request
- `401` - Unauthorized
- `404` - Not Found
- `500` - Internal Server Error

## Testing

Run tests:
```bash
go test ./...
```

## Project Structure Highlights

- **Clean Architecture**: Clear separation between handlers, business logic, and data access
- **Dependency Injection**: Services are injected through constructors
- **Interface-based Design**: Easy to mock and test
- **Context Propagation**: Request context flows through all layers
- **Graceful Shutdown**: Proper cleanup of resources on termination

## Technologies Used

- **Framework**: Echo (high-performance HTTP framework)
- **Database**: PostgreSQL with pgx driver
- **Authentication**: JWT with golang-jwt
- **Logging**: Zap (structured, high-performance logging)
- **Validation**: go-playground/validator
- **Documentation**: Swagger/OpenAPI
- **Containerization**: Docker & Docker Compose

## License

Apache 2.0

## Author

Built as part of the CashInvoice Golang Assignment
