# CORS Middleware Implementation

**Date:** January 28, 2025
**Status:** ✅ Implemented

## Overview

CORS (Cross-Origin Resource Sharing) middleware has been implemented to allow the Next.js frontend to communicate with the Go backend API securely.

## What is CORS?

CORS is a browser security feature that restricts cross-origin HTTP requests initiated from scripts. Our middleware allows the frontend to:

1. Make API requests to the backend
2. Include authentication headers
3. Handle preflight OPTIONS requests
4. Receive proper CORS headers

## Implementation

### File: `internal/api/middleware/cors.go`

**Key Features:**
- Configurable allowed origins
- Configurable allowed methods
- Configurable allowed headers
- Exposed headers for response
- Credentials support (for cookies/auth tokens)
- Preflight OPTIONS handling
- Max-age caching for preflight
- Environment-based configurations

### Configuration Modes

#### 1. Default Configuration
```go
config := middleware.DefaultConfig()
```

**Settings:**
- **Origins**: localhost:3000, localhost:3001, 127.0.0.1:3000, 127.0.0.1:3001
- **Methods**: GET, POST, PUT, DELETE, PATCH, OPTIONS, HEAD
- **Headers**: Origin, Content-Type, Accept, Authorization, X-Requested-With
- **Credentials**: Enabled
- **MaxAge**: 24 hours

#### 2. Development Configuration
```go
config := middleware.DevelopmentConfig()
```

**Settings:**
- **Origins**: * (wildcard - allows all origins)
- **Methods**: All methods
- **Headers**: * (wildcard - allows all headers)
- **Credentials**: Enabled
- **MaxAge**: 24 hours

⚠️ **Use only in development!**

#### 3. Production Configuration
```go
config := middleware.ProdConfig()
```

**Settings:**
- **Origins**: Configured list (must add your domains)
- **Methods**: GET, POST, PUT, DELETE, PATCH, OPTIONS
- **Headers**: Origin, Content-Type, Accept, Authorization
- **Credentials**: Enabled
- **MaxAge**: 24 hours

## Usage

### Basic Setup

In `cmd/api/main.go`:

```go
import "github.com/yourorg/querybase/internal/api/middleware"

// Add CORS middleware
corsConfig := middleware.DefaultConfig()
router.Use(middleware.CORSMiddleware(corsConfig))
```

### Environment-Based Setup

```go
import "github.com/yourorg/querybase/internal/api/middleware"

// Use environment variable
corsMiddleware := middleware.CORSMiddlewareFromEnv(cfg.Server.Mode)
router.Use(corsMiddleware)
```

## Configuration

### AllowedOrigins

Add your frontend origins to the allowed list:

```go
AllowedOrigins: []string{
	"http://localhost:3000",
	"http://localhost:3001",
	"https://querybase.example.com",  // Production domain
	"https://app.querybase.com",        // Another production domain
}
```

**Development Tip:** Use `["*"]` to allow all origins (not for production!)

### AllowedMethods

Specify which HTTP methods are allowed:

```go
AllowedMethods: []string{
	"GET",
	"POST",
	"PUT",
	"DELETE",
	"PATCH",
	"OPTIONS",
	"HEAD",
}
```

### AllowedHeaders

Headers that can be included in requests:

```go
AllowedHeaders: []string{
	"Origin",
	"Content-Type",
	"Accept",
	"Authorization",
	"X-Requested-With",
}
```

### ExposedHeaders

Headers that browsers can read from responses:

```go
ExposedHeaders: []string{
	"Content-Length",
	"Content-Type",
}
```

### AllowCredentials

Set to `true` to allow:
- Cookies
- Authorization headers
- TLS client certificates

```go
AllowCredentials: true
```

**Note:** When `AllowCredentials` is true, `AllowedOrigins` cannot be `*` (wildcard).

### MaxAge

How long (in seconds) to cache preflight OPTIONS responses:

```go
MaxAge: 86400  // 24 hours
```

## Testing

### Unit Tests

Run CORS middleware tests:

```bash
go test ./internal/api/middleware/cors_test.go -v
```

**Tests include:**
- ✅ Preflight OPTIONS requests
- ✅ Actual GET/POST requests
- ✅ Unauthorized origin rejection
- ✅ Wildcard origin allows all
- ✅ Credentials handling
- ✅ Exposed headers

### Manual Testing

#### 1. Test Preflight Request

```bash
curl -X OPTIONS http://localhost:8080/api/v1/auth/login \
  -H "Origin: http://localhost:3000" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: content-type" \
  -v
```

**Expected Response:**
- Status: 204 No Content
- Headers:
  - `Access-Control-Allow-Origin: http://localhost:3000`
  - `Access-Control-Allow-Methods: GET, POST, ...`
  - `Access-Control-Allow-Headers: Origin, Content-Type, ...`

#### 2. Test Actual Request

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Origin: http://localhost:3000" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' \
  -v
```

**Expected Response:**
- Status: 200 OK (or appropriate response)
- Headers:
  - `Access-Control-Allow-Origin: http://localhost:3000`
  - `Access-Control-Allow-Credentials: true`

#### 3. Test Unauthorized Origin

```bash
curl -X GET http://localhost:8080/health \
  -H "Origin: http://malicious-site.com" \
  -v
```

**Expected Response:**
- Status: 403 Forbidden
- Body: `{"error":"Origin not allowed"}`

## Security Considerations

### Production Deployment

1. **Update AllowedOrigins** with your production domains:
   ```go
   AllowedOrigins: []string{
       "https://querybase.example.com",
       "https://app.querybase.com",
   }
   ```

2. **Remove wildcard origins:**
   - ❌ Don't use `AllowedOrigins: []string{"*"}` in production

3. **Use HTTPS only:**
   - Only include HTTPS origins in production
   - Remove localhost origins

4. **Set specific headers:**
   - Only include headers your application needs
   - Avoid `AllowedHeaders: []string{"*"}` in production

### Common CORS Errors

#### Error 1: "No 'Access-Control-Allow-Origin' header"

**Problem:** CORS middleware not added or misconfigured

**Solution:** Ensure CORS middleware is added before routes:
```go
router.Use(middleware.CORSMiddleware(config))
routes.SetupRoutes(router, ...)  // CORS must be first
```

#### Error 2: "Credentials flag is true, but Origin is '*'"

**Problem:** Cannot use wildcard with credentials

**Solution:** Specify exact origins:
```go
AllowedOrigins: []string{"http://localhost:3000"}  // Exact origin
AllowCredentials: true
```

#### Error 3: "Preflight request fails"

**Problem:** OPTIONS request not handled

**Solution:** CORS middleware handles OPTIONS automatically

## Browser Console Testing

Test CORS from browser console:

```javascript
// Test fetch
fetch('http://localhost:8080/health', {
  method: 'GET',
  headers: {
    'Content-Type': 'application/json',
  },
  credentials: 'include'  // Include cookies
})
.then(response => response.json())
.then(data => console.log(data))
.catch(error => console.error('Error:', error));
```

## Integration with Frontend

### Frontend Configuration

The Next.js frontend should make requests like:

```typescript
const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

fetch(`${API_URL}/api/v1/auth/login`, {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  credentials: 'include',  // Important for cookies
  body: JSON.stringify({
    username: 'admin',
    password: 'admin123',
  }),
});
```

### Environment Variables

In `.env.local`:
```env
NEXT_PUBLIC_API_URL=http://localhost:8080
```

## Troubleshooting

### Issue: CORS errors in browser console

**Symptoms:**
```
Access to fetch at 'http://localhost:8080/api/v1/auth/login' from origin 'http://localhost:3000'
has been blocked by CORS policy
```

**Solutions:**

1. **Check backend is running CORS middleware:**
   ```bash
   curl -I http://localhost:8080/health \
     -H "Origin: http://localhost:3000"
   ```

2. **Verify CORS middleware is registered:**
   - Check `cmd/api/main.go` includes `router.Use(middleware.CORSMiddleware(config))`

3. **Check configuration:**
   - Verify `AllowedOrigins` includes your frontend origin
   - Verify `AllowCredentials` is true if sending cookies

4. **Restart API server:**
   ```bash
   make run-api
   ```

### Issue: Preflight request fails

**Symptoms:**
- OPTIONS request returns 404 or 403
- Actual POST request never happens

**Solution:**
- Ensure CORS middleware handles OPTIONS
- OPTIONS handler should return 204 with proper headers

## Configuration Examples

### Example 1: Development Setup

```go
// In main.go
corsConfig := middleware.DevelopmentConfig()
router.Use(middleware.CORSMiddleware(corsConfig))
```

**Result:**
- Allows all origins
- Allows all headers
- Allows all methods
- Suitable for local development

### Example 2: Production with Specific Domain

```go
// In main.go
corsConfig := &middleware.Config{
    AllowedOrigins: []string{
        "https://querybase.example.com",
    },
    AllowedMethods: []string{
        "GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS",
    },
    AllowedHeaders: []string{
        "Origin", "Content-Type", "Accept", "Authorization",
    },
    ExposedHeaders: []string{
        "Content-Length", "Content-Type",
    },
    AllowCredentials: true,
    MaxAge: 86400,
}
router.Use(middleware.CORSMiddleware(corsConfig))
```

**Result:**
- Only allows specified domain
- Restricts methods to those needed
- Production-ready configuration

### Example 3: Multiple Frontend Domains

```go
corsConfig := &middleware.Config{
    AllowedOrigins: []string{
        "https://app.querybase.com",
        "https://admin.querybase.com",
        "https://querybase.com",
    },
    AllowedMethods: []string{
        "GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS",
    },
    AllowedHeaders: []string{
        "Origin", "Content-Type", "Accept", "Authorization",
        "X-Requested-With", "X-HTTP-Method-Override",
    },
    ExposedHeaders: []string{
        "Content-Length", "Content-Type", "X-Total-Count",
    },
    AllowCredentials: true,
    MaxAge: 86400,
}
router.Use(middleware.CORSMiddleware(corsConfig))
```

**Result:**
- Multiple frontend domains supported
- Consistent security policy across all

## Best Practices

1. **Always specify exact origins in production**
   - Never use `*` in production
   - List each allowed domain explicitly

2. **Keep MaxAge reasonable**
   - 86400 seconds (24 hours) is typical
   - Longer values reduce security

3. **Limit AllowedHeaders**
   - Only include headers your app uses
   - Avoid wildcard headers in production

4. **Test CORS before deploying**
   - Use curl to test preflight requests
   - Test from browser console
   - Verify with production domains

5. **Monitor CORS logs**
   - Log unauthorized origin attempts
   - Alert on suspicious patterns

## Related Documentation

- [MDN Web Docs - CORS](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS)
- [Gin CORS Documentation](https://github.com/gin-gonic/gin)
- [Testing Guide](../../docs/development/testing.md)

---

**Last Updated:** January 28, 2025
**Status:** ✅ Implemented and Tested
