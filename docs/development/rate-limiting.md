# Rate Limiting Middleware Implementation

**Date:** January 28, 2026
**Status:** ✅ Implemented

## Overview

Rate limiting middleware has been implemented to protect the API from abuse and ensure fair resource allocation among all clients.

## What is Rate Limiting?

Rate limiting restricts the number of requests a client can make to the API within a specific time window. This helps:
- Prevent DDoS attacks and abuse
- Ensure fair resource allocation
- Protect server resources
- Maintain API performance for all users

## Implementation

### File: `internal/api/middleware/ratelimit.go`

**Key Features:**
- Token bucket algorithm for efficient rate limiting
- In-memory storage with automatic cleanup
- Per-IP address rate limiting
- Configurable limits and burst sizes
- Skip paths for public endpoints
- Per-path rate limiting support
- Automatic token replenishment over time

### Algorithm: Token Bucket

The middleware uses the **token bucket algorithm**:

1. Each client has a bucket of tokens
2. Each request consumes 1 token
3. Tokens replenish at a fixed rate over time
4. Requests are allowed when tokens are available
5. Requests are blocked (429) when bucket is empty

**Advantages:**
- Allows bursts of traffic (up to burst size)
- Smooth rate limiting over time
- Efficient memory usage
- Fair to all clients

## Configuration

### Default Configuration

```go
rateLimitConfig := middleware.DefaultRateLimitConfig()
```

**Default Settings:**
- **RequestsPerMinute**: 60 (1 request per second)
- **BurstSize**: 10 (allow bursts of up to 10 requests)
- **SkipSuccessfulRequests**: false
- **SkipPaths**: /health, /favicon.ico, /static, /api/v1/auth/login

### Custom Configuration

```go
config := &middleware.RateLimiterConfig{
    RequestsPerMinute:      120, // 2 requests per second
    BurstSize:              20,   // Allow bursts of 20
    SkipSuccessfulRequests: false,
    SkipPaths: []string{
        "/health",
        "/api/v1/public",
    },
}
```

### Rate Limiter Parameters

#### RequestsPerMinute
- **Type**: int
- **Description**: Maximum number of requests allowed per minute
- **Default**: 60
- **Formula**: `RequestsPerMinute / 60 = requests per second`

#### BurstSize
- **Type**: int
- **Description**: Maximum number of requests allowed in a short burst
- **Default**: 10
- **Purpose**: Handles temporary traffic spikes
- **Recommendation**: 10-20% of RequestsPerMinute

#### SkipSuccessfulRequests
- **Type**: bool
- **Description**: Don't count successful requests (status < 400)
- **Default**: false
- **Use Case**: Only rate limit failed requests (not recommended)

#### SkipPaths
- **Type**: []string
- **Description**: Paths to exclude from rate limiting
- **Default**: /health, /favicon.ico, /static, /api/v1/auth/login
- **Recommendation**: Skip public endpoints and health checks

## Usage

### Basic Setup

In `cmd/api/main.go`:

```go
import "github.com/yourorg/querybase/internal/api/middleware"

// Add rate limiting middleware
rateLimitConfig := middleware.DefaultRateLimitConfig()
router.Use(middleware.RateLimiterMiddleware(rateLimitConfig))
```

### Per-Path Rate Limiting

Configure different rate limits for different endpoints:

```go
// Create rate limiter with default config
limiter := middleware.NewRateLimiterByPath(middleware.DefaultRateLimitConfig())

// Add stricter limits for expensive operations
strictConfig := &middleware.RateLimiterConfig{
    RequestsPerMinute: 10, // Only 10 requests per minute
    BurstSize:         2,
}
limiter.AddPath("/api/v1/queries", strictConfig)

// Add lenient limits for public endpoints
publicConfig := &middleware.RateLimiterConfig{
    RequestsPerMinute: 120, // 2 requests per second
    BurstSize:         30,
}
limiter.AddPath("/api/v1/public", publicConfig)

// Use the middleware
router.Use(limiter.Middleware())
```

## Response

When rate limit is exceeded, clients receive:

**Status Code:** `429 Too Many Requests`

**Response Body:**
```json
{
  "error": "Rate limit exceeded. Please try again later.",
  "code": 429
}
```

**Headers (Future Enhancement):**
```
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1640745600
```

## Testing

### Unit Tests

Run rate limiter tests:

```bash
go test ./internal/api/middleware/ratelimit_test.go -v
```

**Tests include:**
- ✅ Allows requests within rate limit
- ✅ Blocks requests exceeding rate limit
- ✅ Skips specified paths
- ✅ Different IP addresses have separate limits
- ✅ Tokens replenish over time
- ✅ Per-path rate limiting

### Manual Testing

#### 1. Test Normal Requests (Should Pass)

```bash
# First request
curl -i http://localhost:8080/api/v1/queries

# Second request (burst allowed)
curl -i http://localhost:8080/api/v1/queries
```

**Expected:** Both requests return their normal response codes

#### 2. Test Rate Limit Exceeded

```bash
# Send many requests quickly
for i in {1..15}; do
  echo "Request $i:"
  curl -s -o /dev/null -w "Status: %{http_code}\n" \
    http://localhost:8080/api/v1/queries
done
```

**Expected:** First 10-12 requests succeed, subsequent requests return `429 Too Many Requests`

#### 3. Test Skipped Paths

```bash
# Health endpoint should not be rate limited
for i in {1..100}; do
  curl -s http://localhost:8080/health > /dev/null
done
echo "All 100 requests completed successfully"
```

**Expected:** All requests succeed with `200 OK`

#### 4. Test Token Replenishment

```bash
# Exhaust rate limit
for i in {1..20}; do
  curl -s http://localhost:8080/api/v1/test > /dev/null
done

# Wait for tokens to replenish
sleep 10

# Request should succeed again
curl -i http://localhost:8080/api/v1/test
```

**Expected:** Request after waiting returns `200 OK`

## Architecture

### In-Memory Storage

The rate limiter uses in-memory storage with these characteristics:

**Pros:**
- Fast lookups (O(1))
- No external dependencies
- Simple implementation

**Cons:**
- Not distributed (each server has its own limiter)
- Lost on restart
- Not suitable for horizontal scaling

**Data Structure:**
```go
type tokenBucket struct {
    tokens     int           // Current token count
    lastUpdate time.Time     // Last update timestamp
}

type inMemoryRateLimiter struct {
    buckets map[string]*tokenBucket  // IP -> bucket mapping
    mu      sync.RWMutex              // Read-write lock
    config  *RateLimiterConfig        // Configuration
}
```

### Cleanup Process

To prevent memory leaks, a background goroutine runs every 5 minutes to remove old entries:

```go
// Removes entries not used in 10 minutes
if now.Sub(bucket.lastUpdate) > 10*time.Minute {
    delete(rl.buckets, key)
}
```

### Token Replenishment

Tokens are added based on elapsed time:

```go
elapsed := now.Sub(bucket.lastUpdate)
tokensToAdd := int(elapsed.Minutes() * float64(rl.config.RequestsPerMinute))
bucket.tokens += tokensToAdd
```

**Example:**
- RequestsPerMinute: 60
- Elapsed time: 30 seconds
- Tokens added: 30 (0.5 minutes * 60)

## Production Recommendations

### For Small Applications (< 1000 users)

Use the current in-memory rate limiter:
- Simple and fast
- Single server deployment
- Default limits (60/min)

### For Large Applications (> 1000 users)

Consider distributed rate limiting:

**Option 1: Redis-based Rate Limiter**
```go
// Use Redis for distributed rate limiting
// Share state across multiple servers
// Persistent across restarts
```

**Option 2: External API Gateway**
```
Client → API Gateway (rate limit) → Backend Servers
         - AWS API Gateway
         - Cloudflare
         - Nginx rate limiting
```

### Rate Limit Guidelines

| Endpoint Type | Requests/Minute | Burst Size | Reason |
|--------------|----------------|------------|---------|
| Public API | 60 | 10 | Standard usage |
| Authentication | 120 | 20 | Login attempts |
| Query Execution | 30 | 5 | Expensive operations |
| Data writes | 10 | 2 | Very expensive |
| Webhooks | 600 | 100 | High frequency |

### Monitoring

Monitor these metrics:

1. **Rate 429 responses**
   - High 429 rates = limits too strict
   - Track per-endpoint

2. **Memory usage**
   - Monitor rate limiter map size
   - Check cleanup is working

3. **Response times**
   - Rate limiter should add < 1ms overhead

4. **User impact**
   - Track legitimate users being blocked
   - Adjust limits based on feedback

## Configuration Examples

### Example 1: Development Environment

```go
config := &middleware.RateLimiterConfig{
    RequestsPerMinute: 120, // Lenient for development
    BurstSize:         30,
    SkipPaths: []string{
        "/health",
        "/api/v1/auth/login",
    },
}
```

### Example 2: Production - Strict

```go
config := &middleware.RateLimiterConfig{
    RequestsPerMinute: 30, // Strict limit
    BurstSize:         5,  // Small burst
    SkipPaths: []string{
        "/health",
    },
}
```

### Example 3: Production - Per-Endpoint

```go
limiter := middleware.NewRateLimiterByPath(nil)

// Query execution - strict
limiter.AddPath("/api/v1/queries", &RateLimiterConfig{
    RequestsPerMinute: 10,
    BurstSize:         2,
})

// Data sources - moderate
limiter.AddPath("/api/v1/datasources", &RateLimiterConfig{
    RequestsPerMinute: 30,
    BurstSize:         5,
})

// Approvals - lenient
limiter.AddPath("/api/v1/approvals", &RateLimiterConfig{
    RequestsPerMinute: 60,
    BurstSize:         10,
})
```

### Example 4: API Key Based (Future Enhancement)

```go
// Rate limit by API key instead of IP
// Allows different limits per API tier
limiter := NewAPIKeyRateLimiter()
limiter.SetLimit("free", 60)      // 60/min for free tier
limiter.SetLimit("pro", 600)      // 600/min for pro tier
limiter.SetLimit("enterprise", 6000) // 6000/min for enterprise
```

## Security Considerations

### Rate Limit Bypass Prevention

**Issue:** Clients can use multiple IPs to bypass rate limits

**Solutions:**
1. **User-based rate limiting** (if authenticated)
   ```go
   userID := c.GetString("user_id")
   key := fmt.Sprintf("user:%s", userID)
   ```

2. **API key rate limiting**
   ```go
   apiKey := c.GetHeader("X-API-Key")
   key := fmt.Sprintf("apikey:%s", apiKey)
   ```

3. **CAPTCHA after excessive failures**
   - Show CAPTCHA after N failed requests
   - Require CAPTCHA to continue

### DDoS Protection

Rate limiting is one layer of DDoS protection. Combine with:

1. **Connection limiting** (web server level)
2. **IP whitelisting/blacklisting**
3. **Geographic blocking**
4. **CDN protection** (Cloudflare, AWS Shield)
5. **Traffic analysis** (detect patterns)

## Troubleshooting

### Issue: All requests return 429

**Symptoms:**
- Legitimate users getting rate limited immediately
- No requests succeed

**Solutions:**
1. Check rate limit configuration:
   ```go
   config.RequestsPerMinute  // Is this too low?
   config.BurstSize          // Is this too low?
   ```
2. Check if IP detection is working:
   ```go
   log.Printf("Client IP: %s", c.ClientIP())
   ```
3. Check if cleanup is running (memory exhaustion)

### Issue: Rate limit not working

**Symptoms:**
- Unlimited requests allowed
- No 429 responses

**Solutions:**
1. Verify middleware is registered:
   ```go
   router.Use(middleware.RateLimiterMiddleware(config))
   ```
2. Check if path is in SkipPaths
3. Check middleware order (must be before routes)

### Issue: Memory usage increasing

**Symptoms:**
- Memory grows over time
- OOM kills

**Solutions:**
1. Verify cleanup goroutine is running:
   ```go
   go limiter.cleanup()  // Should be started
   ```
2. Reduce cleanup interval
3. Implement Redis-based rate limiter for scale

## Best Practices

1. **Set reasonable defaults**
   - Start with 60 requests/min
   - Adjust based on usage patterns

2. **Skip public endpoints**
   - Health checks
   - Static assets
   - Documentation pages

3. **Monitor and adjust**
   - Track 429 rates
   - Get user feedback
   - Adjust limits quarterly

4. **Document rate limits**
   - Include in API documentation
   - Show headers in response
   - Provide retry guidance

5. **Consider authentication**
   - Use user-based limits for authenticated users
   - Use IP-based limits for anonymous users

6. **Test before production**
   - Load test with expected traffic
   - Verify limits trigger correctly
   - Check memory usage

## Future Enhancements

### Planned Features

1. **Sliding Window Rate Limiting**
   - More accurate than token bucket
   - Prevents bursts at window boundaries

2. **Response Headers**
   ```
   X-RateLimit-Limit: 60
   X-RateLimit-Remaining: 45
   X-RateLimit-Reset: 1640745600
   Retry-After: 30
   ```

3. **Redis-based Storage**
   - Distributed rate limiting
   - Persistent across restarts
   - Share state across servers

4. **User-based Rate Limiting**
   - Different limits per user role
   - Track usage per user
   - Premium tier support

5. **Rate Limit Analytics**
   - Track rate limit hits
   - Identify abusers
   - Export metrics

6. **Dynamic Configuration**
   - Adjust limits without restart
   - API endpoint to change limits
   - Per-client overrides

## Related Documentation

- [CORS Implementation](cors-implementation.md)
- [Authentication](../architecture/authentication.md)
- [API Documentation](../api/README.md)
- [Security Best Practices](../security/README.md)

---

**Last Updated:** January 28, 2026
**Status:** ✅ Implemented and Tested
