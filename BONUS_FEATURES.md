# Bonus Features Implementation

This document describes the implementation of the bonus features: **Pagination with Scroll ID** and **Rate Limiting**.

## 1. Pagination with Scroll ID

### Overview
Implemented cursor-based pagination using scroll IDs instead of traditional offset-based pagination. This approach is more efficient for large datasets and prevents issues with data consistency during pagination.

### Implementation Details

#### Database Layer (Stored Procedure)
- Created `list_tasks_paginated()` stored procedure in `migrations/001_init.sql`
- The procedure accepts:
  - `p_user_id`: User ID for filtering (if not admin)
  - `p_is_admin`: Boolean flag for admin access
  - `p_status`: Optional status filter
  - `p_limit`: Number of records to return
  - `p_scroll_id`: Cursor position for pagination

- Returns:
  - Task records
  - Total count
  - `next_scroll_id`: ID to use for the next page

#### Store Layer
- Updated `TaskStore.List()` method signature:
  ```go
  List(ctx context.Context, userID string, isAdmin bool, limit int, scrollID, status string) ([]entity.Task, int, string, error)
  ```
- Removed inline SQL queries and now calls the stored procedure

#### Application Layer
- Updated `TaskUseCase.List()` to pass through scroll_id parameter

#### API Layer
- Modified `/tasks` endpoint to accept `scroll_id` query parameter instead of `page`
- Response includes `next_scroll` field for pagination
- Example request:
  ```
  GET /api/tasks?limit=10&scroll_id=<previous_scroll_id>&status=pending
  ```

### Usage Example

**First Request:**
```bash
curl -H "Authorization: Bearer <token>" \
  "http://localhost:8000/api/tasks?limit=10"
```

**Response:**
```json
{
  "tasks": [...],
  "total": 50,
  "limit": 10,
  "next_scroll": "550e8400-e29b-41d4-a716-446655440010"
}
```

**Next Page Request:**
```bash
curl -H "Authorization: Bearer <token>" \
  "http://localhost:8000/api/tasks?limit=10&scroll_id=550e8400-e29b-41d4-a716-446655440010"
```

## 2. Rate Limiting

### Overview
Implemented two-tier rate limiting:
1. **Global Rate Limiting**: Limits total requests across all users
2. **Per-IP Rate Limiting**: Limits requests per individual IP address

### Implementation Details

#### Configuration
Added rate limit configuration in `config.yaml`:
```yaml
RateLimit:
  global_rate: 100    # 100 requests per second globally
  global_burst: 200   # Allow burst of 200 requests
  ip_rate: 10         # 10 requests per second per IP
  ip_burst: 20        # Allow burst of 20 requests per IP
```

#### Middleware
- Created `rateLimiter` middleware in `cmd/api/middleware.go`
- Uses `golang.org/x/time/rate` package for token bucket algorithm
- Maintains separate limiters for:
  - Global requests (shared across all IPs)
  - Per-IP requests (isolated per client)

#### Rate Limiter Logic
1. First checks global rate limit
2. Then checks per-IP rate limit
3. Returns HTTP 429 (Too Many Requests) if limit exceeded

#### Response on Rate Limit
```json
{
  "error": "rate limit exceeded for your IP"
}
```
or
```json
{
  "error": "global rate limit exceeded"
}
```

### Testing Rate Limits

**Test Global Rate Limit:**
```bash
# Send rapid requests to trigger global limit
for i in {1..250}; do
  curl -H "Authorization: Bearer <token>" \
    "http://localhost:8000/api/tasks" &
done
```

**Test Per-IP Rate Limit:**
```bash
# Send rapid requests from single IP
for i in {1..30}; do
  curl -H "Authorization: Bearer <token>" \
    "http://localhost:8000/api/tasks"
done
```

### Configuration Tuning

Adjust rate limits in `config.yaml` based on your needs:
- **Higher traffic**: Increase `global_rate` and `global_burst`
- **Stricter per-user limits**: Decrease `ip_rate` and `ip_burst`
- **Development**: Set higher values to avoid hitting limits during testing

## Architecture Benefits

### Pagination
- **Performance**: Cursor-based pagination is O(1) vs O(n) for offset-based
- **Consistency**: No duplicate/missing records when data changes during pagination
- **Scalability**: Efficient for large datasets
- **Database-level**: Logic in stored procedure reduces application complexity

### Rate Limiting
- **DDoS Protection**: Global limit prevents server overload
- **Fair Usage**: Per-IP limit ensures equitable resource distribution
- **Configurable**: Easy to adjust limits without code changes
- **Memory Efficient**: Token bucket algorithm with minimal overhead

## Migration Notes

If you have an existing database, run the migration to add the stored procedure:
```bash
psql -U postgres -d taskmanager -f migrations/001_init.sql
```

Or if using Docker:
```bash
docker-compose exec db psql -U postgres -d taskmanager -f /docker-entrypoint-initdb.d/001_init.sql
```
