# Quick Start Guide

## Prerequisites
- Docker and Docker Compose installed
- (Optional) Go 1.21+ for local development
- (Optional) curl and jq for testing

## Step 1: Start the Application

```bash
cd /mnt/d/Rohbot/task-manager
docker-compose up -d
```

Wait for the containers to start (~30 seconds).

## Step 2: Verify the Application is Running

```bash
curl http://localhost:8000/api/health
```

Expected response:
```json
{
  "status": "ok"
}
```

## Step 3: Register a User

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

## Step 4: Login and Get Token

```bash
curl -X POST http://localhost:8000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john",
    "password": "password123"
  }'
```

Copy the token from the response.

## Step 5: Create a Task

```bash
export TOKEN="your-token-here"

curl -X POST http://localhost:8000/api/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "My first task",
    "description": "This is a test task"
  }'
```

## Step 6: List Your Tasks

```bash
curl -X GET http://localhost:8000/api/tasks \
  -H "Authorization: Bearer $TOKEN"
```

## Step 7: Access Swagger Documentation

Open your browser and navigate to:
```
http://localhost:8000/swagger/index.html
```

## Step 8: Run Automated Tests

```bash
./test_api.sh
```

## Step 9: View Logs

```bash
docker-compose logs -f api-go
```

## Step 10: Stop the Application

```bash
docker-compose down
```

## Testing Background Worker

The background worker auto-completes tasks after 5 minutes (configurable).

1. Create a task
2. Wait 5 minutes
3. Check the task status - it should be "completed"

To test faster, modify `AutoCompleteMin` in `internal/config/config.yaml` to 1 minute.

## Troubleshooting

### Port Already in Use
If port 8000 or 5432 is already in use:
```bash
# Change ports in docker-compose.yaml
ports:
  - "8001:8000"  # API
  - "5433:5432"  # PostgreSQL
```

### Database Connection Issues
```bash
# Check if PostgreSQL is running
docker-compose ps

# View database logs
docker-compose logs db
```

### Application Errors
```bash
# View application logs
docker-compose logs api-go

# Restart the application
docker-compose restart api-go
```

## Next Steps

- Import `postman_collection.json` into Postman for easier testing
- Read `README.md` for detailed documentation
- Check `IMPLEMENTATION.md` for architecture details
- Explore the Swagger UI for interactive API testing

## Admin User Testing

To test admin functionality:

1. Register an admin user:
```bash
curl -X POST http://localhost:8000/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "email": "admin@example.com",
    "password": "admin123",
    "role": "admin"
  }'
```

2. Login as admin and get token
3. List all tasks (you'll see tasks from all users)

## Configuration

Edit `internal/config/config.yaml` to customize:
- Server port
- JWT token TTL
- Auto-complete delay
- Database connection
- Log levels

After changes, restart:
```bash
docker-compose restart api-go
```
