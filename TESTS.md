# Unit Tests Implementation

## Test Coverage

### 1. API Layer Tests (`cmd/api/tasks_test.go`)
**Mocks:** Application layer (TaskUseCase, AuthUseCase)

**Tests:**
- `TestCreateTask` - Create task endpoint
  - Success case
  - Validation error
- `TestListTasks` - List tasks with pagination
  - Success case
- `TestGetTask` - Get task by ID
  - Success case
  - Not found case
- `TestDeleteTask` - Delete task
  - Success case
- `TestLogin` - User login
  - Success case
  - Invalid credentials
- `TestRegister` - User registration
  - Success case

### 2. Application Layer Tests (`internal/application/tasks_test.go`)
**Mocks:** Store layer (TaskStore)

**Tests:**
- `TestTaskUseCase_Create` - Create task use case
  - Success case
  - Store error
- `TestTaskUseCase_GetByID` - Get task by ID
  - Success case
  - Not found case
- `TestTaskUseCase_List` - List tasks
  - Success case
  - Empty result
- `TestTaskUseCase_Delete` - Delete task
  - Success case
  - Store error

### 3. Worker Tests (`internal/worker/task_worker_test.go`)
**Mocks:** Store layer (TaskStore)

**Tests:**
- `TestTaskWorker_ProcessTaskQueue` - Process task queue
  - Success case
- `TestTaskWorker_ScanPendingTasks` - Scan pending tasks
  - Fetches and queues tasks
  - Handles empty result
- `TestTaskWorker_UpdateStatus` - Update task status
  - Success case
  - Error case

## Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./cmd/api
go test ./internal/application
go test ./internal/worker

# Verbose output
go test -v ./...

# Coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Test Structure

### API Layer
```
Request → Mock Application Layer → Assertions
```

### Application Layer
```
Use Case → Mock Store Layer → Assertions
```

### Worker
```
Worker Logic → Mock Store Layer → Assertions
```

## Mock Objects

All tests use `testify/mock` for creating mocks:
- `MockTaskUseCase` - Mocks application.ITaskUseCase
- `MockAuthUseCase` - Mocks application.IAuthUseCase
- `MockTaskStore` - Mocks store.ITaskStore

## Key Testing Patterns

1. **Setup function** - Creates test server/use case with mocks
2. **Table-driven tests** - Multiple scenarios per function
3. **Mock expectations** - Verify correct method calls
4. **Assertions** - Validate responses and errors

## Example Test Run

```bash
$ go test ./cmd/api -v
=== RUN   TestCreateTask
=== RUN   TestCreateTask/success
=== RUN   TestCreateTask/validation_error
--- PASS: TestCreateTask (0.00s)
    --- PASS: TestCreateTask/success (0.00s)
    --- PASS: TestCreateTask/validation_error (0.00s)
=== RUN   TestListTasks
--- PASS: TestListTasks (0.00s)
PASS
ok      github.com/Roh-Bot/blog-api/cmd/api     0.123s
```
