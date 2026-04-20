# Rate-Limited API Service

A production-considerate rate-limiting API service built with **Go** and **Gin**, implementing the **Sliding Window Log** algorithm for accurate per-user rate limiting under concurrent load.

## 🏗️ Architecture

```
cmd/main.go                          ← Entry point & dependency wiring
internal/
├── handler/handler.go               ← HTTP layer (thin — parse, respond)
├── ratelimiter/limiter.go           ← Business logic (Sliding Window Log)
├── ratelimiter/limiter_test.go      ← Comprehensive test suite
└── store/
    ├── store.go                     ← Store interface (abstraction)
    └── memory.go                    ← In-memory implementation (sync.RWMutex)
```

### Layer Responsibilities

| Layer | Responsibility | Depends On |
|-------|---------------|------------|
| **Handler** | Parse HTTP requests, send responses | RateLimiter |
| **RateLimiter** | Rate limit decisions (allow/block) | Store (interface) |
| **Store** | Data persistence (timestamps, counts) | Nothing |

The `Store` is defined as an **interface**, allowing the storage backend to be swapped (e.g., Memory → Redis) without modifying any business logic. This follows the **Dependency Inversion Principle**.

## 🚀 How to Run

### Prerequisites
- Go 1.22+

### Run Locally
```bash
git clone https://github.com/Mohammad-Niyas/rate-limiter-api.git
cd rate-limiter-api
go mod tidy
go run cmd/main.go
```

Server starts at `http://localhost:8080`

### Run with Docker
```bash
docker-compose up --build
```

### Run Tests
```bash
go test ./... -v
```

### Run with Race Detector (validates concurrency safety)
```bash
go test ./... -race -v
```

## 📌 API Endpoints

### POST /request
Submit a rate-limited request.

**Request:**
```json
{
  "user_id": "user123",
  "payload": "any data here"
}
```

**Success Response (200):**
```json
{
  "status": "accepted",
  "message": "Request processed successfully",
  "remaining": 4
}
```

**Rate Limited Response (429):**
```json
{
  "status": "rejected",
  "message": "Rate limit exceeded. Maximum 5 requests per minute allowed.",
  "retry_after_seconds": 45.2
}
```
Also sets the standard `Retry-After` HTTP header.

**Bad Request (400):**
```json
{
  "error": "user_id is required and must be a non-empty string"
}
```

### GET /stats
Returns per-user request statistics.

**Response (200):**
```json
{
  "stats": [
    {
      "user_id": "user123",
      "total_requests": 12,
      "requests_in_current_window": 3,
      "remaining_requests": 2
    }
  ]
}
```

### GET /health
Health check endpoint.

**Response (200):**
```json
{
  "status": "healthy",
  "time": "2026-04-20T18:00:00+05:30"
}
```

## 🧪 Testing

The test suite covers 6 critical scenarios:

| Test | What It Validates |
|------|------------------|
| `TestAllowUnderLimit` | Requests under the limit are allowed with correct remaining count |
| `TestBlockOverLimit` | 6th request is correctly blocked |
| `TestDifferentUsersIndependent` | User A's rate limit doesn't affect User B |
| `TestWindowExpiry` | Requests are allowed again after the window expires |
| `TestConcurrentRequests` | **20 goroutines, 1 user — exactly 5 allowed** (validates concurrency safety) |
| `TestGetStats` | Stats endpoint returns correct per-user data |

## 🔧 Design Decisions

### Why Sliding Window Log?

I evaluated three algorithms:

| Algorithm | Accuracy | Memory | Chosen? |
|-----------|----------|--------|---------|
| Fixed Window Counter | ❌ Allows 2x burst at window boundary | Low | No |
| **Sliding Window Log** | **✅ Exact accuracy** | **Moderate (5 timestamps/user = 40 bytes)** | **✅ Yes** |
| Sliding Window Counter | ⚠️ Approximate | Low | No |

**Fixed Window** has a known "boundary attack" vulnerability: a user can send 5 requests at 10:00:59 and 5 more at 10:01:00, achieving 10 requests in 2 seconds despite a 5/minute limit. Sliding Window Log eliminates this by tracking individual timestamps.

The memory overhead is negligible — each user stores at most 5 `time.Time` values (40 bytes). Even with 1 million active users, this is only ~40 MB.

### Why Two Levels of Mutex?

- **Store level (`sync.RWMutex`):** Protects the underlying map from concurrent read/write crashes.
- **RateLimiter level (`sync.Mutex`):** Ensures the check-then-add sequence is **atomic**. Without this, two goroutines could both check "count=4 < 5" and both add a request, exceeding the limit.

### Why Interface for Store?

The `Store` interface allows swapping the storage backend without modifying business logic:
- **Current:** `MemoryStore` (in-memory, suitable for single-instance deployment)
- **Future:** `RedisStore` (for distributed rate limiting across multiple instances)

This follows the **Dependency Inversion Principle** (SOLID).

## ⚠️ Limitations & Improvements

### Current Limitations
1. **In-memory storage:** Data is lost on server restart. Not suitable for distributed deployments with multiple instances.
2. **Global mutex in RateLimiter:** All users share one lock. Under very high load (100K+ concurrent users), this could become a bottleneck.
3. **No persistent logging:** Uses standard `log` package. Production would need structured logging (e.g., Zap).
4. **No authentication:** Any client can submit requests. Production would need API key validation.

### What I Would Improve With More Time
1. **Redis backend:** Implement `RedisStore` for distributed rate limiting using `MULTI/EXEC` transactions or Lua scripts for atomic operations.
2. **Per-user locking:** Replace global mutex with `sync.Map` of per-user mutexes to reduce lock contention.
3. **Structured logging:** Add Zap for JSON-formatted, leveled logging.
4. **Metrics endpoint:** Add `/metrics` with Prometheus counters for total requests, blocked requests, and latency histograms.
5. **Configuration:** Load `MaxRequests` and `WindowSize` from environment variables or config file.
6. **Graceful shutdown:** Handle `SIGTERM` signals and drain in-flight requests before stopping.
7. **API authentication:** Add API key middleware to prevent unauthorized access.
8. **Retry queue:** For rejected requests, optionally queue them and process when the window clears.

## 📝 License

MIT
