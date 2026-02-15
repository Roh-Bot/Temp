# Architecture Diagram - Bonus Features

## Request Flow with Rate Limiting

```
┌─────────────────────────────────────────────────────────────────┐
│                         Client Request                          │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Rate Limiter Middleware                      │
│  ┌──────────────────────┐    ┌──────────────────────┐          │
│  │  Global Rate Limit   │    │   Per-IP Rate Limit  │          │
│  │  100 req/s (200 burst)│   │  10 req/s (20 burst) │          │
│  └──────────────────────┘    └──────────────────────┘          │
│                                                                  │
│  ✅ Pass → Continue    ❌ Fail → HTTP 429                       │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      HTTP Logger Middleware                     │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Auth Validation Middleware                   │
│                    (JWT Token Verification)                     │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                         API Handler                             │
│                      (e.g., listTasks)                          │
└─────────────────────────────────────────────────────────────────┘
```

## Pagination Flow

```
┌─────────────────────────────────────────────────────────────────┐
│  Client: GET /api/tasks?limit=10                                │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  API Handler (cmd/api/tasks.go)                                 │
│  - Extract: limit, scroll_id, status                            │
│  - Validate: limit (1-100)                                      │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  Application Layer (internal/application/tasks.go)              │
│  - Pass parameters to store                                     │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  Store Layer (internal/store/tasks.go)                          │
│  - Call stored procedure: list_tasks_paginated()                │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  PostgreSQL Stored Procedure                                    │
│  ┌───────────────────────────────────────────────────┐          │
│  │ 1. Get scroll position (if scroll_id provided)   │          │
│  │ 2. Calculate total count                         │          │
│  │ 3. Fetch tasks with LEAD() for next_scroll_id    │          │
│  │ 4. Return: tasks + total + next_scroll_id        │          │
│  └───────────────────────────────────────────────────┘          │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  Response to Client                                             │
│  {                                                              │
│    "tasks": [...],                                              │
│    "total": 50,                                                 │
│    "limit": 10,                                                 │
│    "next_scroll": "550e8400-..."  ← Use for next request       │
│  }                                                              │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  Client: GET /api/tasks?limit=10&scroll_id=550e8400-...         │
│  (Next page request)                                            │
└─────────────────────────────────────────────────────────────────┘
```

## Data Flow Comparison

### Before (Offset-based)
```
API → Application → Store → Inline SQL Query
                            ↓
                    SELECT * FROM tasks
                    WHERE user_id = $1
                    ORDER BY created_at DESC
                    LIMIT $2 OFFSET $3
                            ↓
                    Performance: O(n) - scans offset rows
```

### After (Cursor-based with Stored Procedure)
```
API → Application → Store → Stored Procedure
                            ↓
                    list_tasks_paginated(
                      user_id, is_admin, status,
                      limit, scroll_id
                    )
                            ↓
                    Performance: O(1) - direct seek
```

## Rate Limiter State Management

```
┌─────────────────────────────────────────────────────────────────┐
│                    Rate Limiter State                           │
│                                                                  │
│  ┌────────────────────────────────────────────────┐             │
│  │  Global Limiter (Shared)                       │             │
│  │  - Rate: 100 req/s                             │             │
│  │  - Burst: 200                                  │             │
│  │  - Token Bucket Algorithm                      │             │
│  └────────────────────────────────────────────────┘             │
│                                                                  │
│  ┌────────────────────────────────────────────────┐             │
│  │  Per-IP Limiters (Map)                         │             │
│  │  ┌──────────────────────────────────────────┐  │             │
│  │  │ IP: 192.168.1.1 → Limiter (10/s, 20)    │  │             │
│  │  │ IP: 192.168.1.2 → Limiter (10/s, 20)    │  │             │
│  │  │ IP: 192.168.1.3 → Limiter (10/s, 20)    │  │             │
│  │  └──────────────────────────────────────────┘  │             │
│  │  - Created on-demand per IP                    │             │
│  │  - Mutex-protected map                         │             │
│  └────────────────────────────────────────────────┘             │
└─────────────────────────────────────────────────────────────────┘
```

## Configuration Flow

```
┌─────────────────────────────────────────────────────────────────┐
│  config.yaml                                                    │
│  ┌────────────────────────────────────────────────┐             │
│  │ RateLimit:                                     │             │
│  │   global_rate: 100                             │             │
│  │   global_burst: 200                            │             │
│  │   ip_rate: 10                                  │             │
│  │   ip_burst: 20                                 │             │
│  └────────────────────────────────────────────────┘             │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  Config Loader (internal/config/config.go)                      │
│  - Parse YAML                                                   │
│  - Load into Config struct                                      │
│  - Store in AtomicConfig                                        │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  Server Initialization (cmd/api/server.go)                      │
│  - Access config via s.Config.Get()                             │
│  - Initialize rate limiters with config values                  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  Rate Limiter Middleware (cmd/api/middleware.go)                │
│  - Use configured limits                                        │
│  - Apply to all API requests                                    │
└─────────────────────────────────────────────────────────────────┘
```

## File Structure

```
task-manager/
├── migrations/
│   └── 001_init.sql                    ← Stored procedure added here
├── internal/
│   ├── config/
│   │   ├── config.go                   ← RateLimit struct added
│   │   └── config.yaml                 ← Rate limit values added
│   ├── store/
│   │   └── tasks.go                    ← List() updated to use stored proc
│   └── application/
│       └── tasks.go                    ← List() signature updated
├── cmd/
│   └── api/
│       ├── middleware.go               ← rateLimiter() added
│       ├── server.go                   ← Middleware registered
│       └── tasks.go                    ← listTasks() updated for scroll_id
├── BONUS_FEATURES.md                   ← Detailed documentation
├── IMPLEMENTATION_SUMMARY.md           ← Technical details
├── QUICK_REFERENCE.md                  ← Quick start guide
├── CHANGES.txt                         ← Summary of changes
└── test_bonus_features.sh              ← Test script
```
