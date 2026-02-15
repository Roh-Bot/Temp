# Bonus Features - Quick Reference

## What Was Added

### 1. ✅ Pagination with Scroll ID
- **Cursor-based pagination** instead of offset-based
- **Stored procedure** `list_tasks_paginated()` in PostgreSQL
- **API changes**: Use `scroll_id` parameter instead of `page`
- **Response includes** `next_scroll` field for next page

### 2. ✅ Rate Limiting
- **Global rate limit**: 100 req/s (burst: 200)
- **Per-IP rate limit**: 10 req/s (burst: 20)
- **HTTP 429** response when limit exceeded
- **Configurable** via `config.yaml`

## Files Changed

```
migrations/001_init.sql          # Added stored procedure
internal/store/tasks.go          # Updated List() method
internal/application/tasks.go    # Updated List() signature
internal/config/config.go        # Added RateLimit config
internal/config/config.yaml      # Added rate limit values
cmd/api/tasks.go                 # Updated listTasks handler
cmd/api/middleware.go            # Added rateLimiter middleware
cmd/api/server.go                # Registered rate limiter
```

## New Files

```
BONUS_FEATURES.md               # Detailed documentation
IMPLEMENTATION_SUMMARY.md       # Implementation details
test_bonus_features.sh          # Test script
```

## Quick Start

### 1. Apply Database Changes
```bash
# If using docker-compose
docker-compose down
docker-compose up -d

# Or manually apply migration
docker-compose exec db psql -U postgres -d taskmanager -f /docker-entrypoint-initdb.d/001_init.sql
```

### 2. Test Pagination
```bash
# Get first page
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/tasks?limit=5"

# Use next_scroll from response for next page
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8000/api/tasks?limit=5&scroll_id=<next_scroll_value>"
```

### 3. Test Rate Limiting
```bash
# Run test script
./test_bonus_features.sh

# Or manually trigger rate limit
for i in {1..30}; do
  curl -H "Authorization: Bearer $TOKEN" \
    "http://localhost:8000/api/tasks"
done
```

## API Examples

### Pagination Request
```http
GET /api/tasks?limit=10&scroll_id=550e8400-e29b-41d4-a716-446655440010&status=pending
Authorization: Bearer <token>
```

### Pagination Response
```json
{
  "tasks": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440010",
      "title": "Task 1",
      "description": "Description",
      "status": "pending",
      "user_id": "550e8400-e29b-41d4-a716-446655440001",
      "created_at": "2026-02-15T06:00:00Z",
      "updated_at": "2026-02-15T06:00:00Z"
    }
  ],
  "total": 50,
  "limit": 10,
  "next_scroll": "550e8400-e29b-41d4-a716-446655440020"
}
```

### Rate Limit Response
```json
{
  "error": "rate limit exceeded for your IP"
}
```

## Configuration

Edit `internal/config/config.yaml`:

```yaml
RateLimit:
  global_rate: 100    # Adjust for your traffic
  global_burst: 200   # Allow temporary spikes
  ip_rate: 10         # Per-IP limit
  ip_burst: 20        # Per-IP burst
```

## Key Benefits

### Pagination
- ✅ O(1) performance vs O(n) for offset
- ✅ No duplicate/missing records
- ✅ Database-optimized with stored procedure
- ✅ Scales to millions of records

### Rate Limiting
- ✅ DDoS protection
- ✅ Fair resource allocation
- ✅ Zero external dependencies
- ✅ Configurable without code changes

## Troubleshooting

### Stored Procedure Not Found
```bash
# Recreate database
docker-compose down -v
docker-compose up -d
```

### Rate Limiting Not Working
- Check config.yaml values are loaded
- Verify middleware is registered in server.go
- Increase request count in tests

### Pagination Returns Empty
- Ensure tasks exist in database
- Check user permissions (admin vs regular user)
- Verify scroll_id is valid

## Next Steps

1. ✅ Review `BONUS_FEATURES.md` for detailed documentation
2. ✅ Run `./test_bonus_features.sh` to verify functionality
3. ✅ Adjust rate limits in `config.yaml` for your use case
4. ✅ Update Swagger docs: `make swagger`

## Notes

- Rate limiters are in-memory (per server instance)
- For distributed systems, consider Redis-based rate limiting
- Scroll IDs are task IDs, ensuring stable pagination
- All inline SQL queries moved to stored procedure as requested
