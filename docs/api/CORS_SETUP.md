# CORS Configuration Guide

QueryBase uses environment-based CORS configuration to easily control which frontend domains can access the backend API.

---

## Quick Setup

### 1. Backend Configuration (Go API)

**Option A: Using Environment Variables (.env)**

Create a `.env` file in the project root:

```bash
# Backend URL (for reference)
BACKEND_URL=http://localhost:8080

# Frontend URL(s) that can access the backend
CORS_ALLOWED_ORIGINS=http://localhost:3000,https://yourdomain.com

# Allow credentials (cookies, auth headers)
CORS_ALLOW_CREDENTIALS=true

# Preflight cache duration (seconds)
CORS_MAX_AGE=86400
```

**Option B: Using config.yaml**

Edit `config/config.yaml`:

```yaml
cors:
  allowed_origins: "http://localhost:3000,http://localhost:3001"
  allow_credentials: true
  max_age: 86400
```

### 2. Frontend Configuration (Next.js)

Create `web/.env.local`:

```bash
# Backend API URL
NEXT_PUBLIC_API_URL=http://localhost:8080
```

---

## Configuration Priority

1. **Environment Variables** (highest priority)
2. **config.yaml** (fallback)
3. **Code defaults** (last resort)

---

## Common Scenarios

### Development Setup

**Backend .env:**
```bash
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001,http://127.0.0.1:3000
```

**Frontend .env.local:**
```bash
NEXT_PUBLIC_API_URL=http://localhost:8080
```

### Production Setup

**Backend .env:**
```bash
BACKEND_URL=https://api.querybase.com
FRONTEND_URL=https://app.querybase.com
CORS_ALLOWED_ORIGINS=https://app.querybase.com,https://www.querybase.com
```

**Frontend .env.local:**
```bash
NEXT_PUBLIC_API_URL=https://api.querybase.com
```

### Multiple Frontend Domains

```bash
CORS_ALLOWED_ORIGINS=https://app1.example.com,https://app2.example.com,https://app3.example.com
```

### Development + Production Domains

```bash
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001,https://app.querybase.com
```

---

## How It Works

### Backend (Go)

**File:** `internal/config/config.go`

1. **Config Structure:**
   ```go
   type CORSConfig struct {
       AllowedOrigins   string  // Comma-separated origins
       AllowCredentials bool    // Allow cookies/auth
       MaxAge           int     // Preflight cache duration
   }
   ```

2. **Environment Variable Binding:**
   ```go
   viper.SetDefault("cors.allowed_origins", "http://localhost:3000")
   viper.SetDefault("cors.allow_credentials", true)
   viper.SetDefault("cors.max_age", 86400)
   viper.AutomaticEnv()  // Automatically read env vars
   ```

3. **Middleware Registration** (`cmd/api/main.go`):
   ```go
   router.Use(middleware.CORSMiddlewareFromConfig(cfg.CORS))
   ```

### Frontend (Next.js)

**File:** `web/src/lib/api-client.ts`

```typescript
const baseURL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

this.client = axios.create({
    baseURL,
    headers: {
        'Content-Type': 'application/json',
    },
});
```

---

## Testing CORS

### 1. Test Preflight Request

```bash
curl -X OPTIONS http://localhost:8080/api/v1/auth/me \
  -H "Origin: http://localhost:3000" \
  -H "Access-Control-Request-Method: GET" \
  -H "Access-Control-Request-Headers: authorization" \
  -v
```

**Expected Response:**
```
HTTP/1.1 204 No Content
Access-Control-Allow-Origin: http://localhost:3000
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, PATCH, OPTIONS, HEAD
Access-Control-Allow-Headers: Origin, Content-Type, Accept, Authorization
Access-Control-Max-Age: 86400
Access-Control-Allow-Credentials: true
```

### 2. Test Actual Request

```bash
curl http://localhost:8080/health \
  -H "Origin: http://localhost:3000" \
  -v
```

**Expected Response:**
```
HTTP/1.1 200 OK
Access-Control-Allow-Origin: http://localhost:3000
Access-Control-Allow-Credentials: true
```

---

## Troubleshooting

### Issue: "Origin not allowed" (403)

**Cause:** Origin not in `CORS_ALLOWED_ORIGINS`

**Solution:**
```bash
# Add your origin to the list
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://your-frontend.com
```

### Issue: "Credentials mode is 'include', but Access-Control-Allow-Origin is '*'"

**Cause:** Cannot use wildcard with credentials

**Solution:** Use specific origins instead of `*`
```bash
# ❌ Wrong
CORS_ALLOWED_ORIGINS=*
CORS_ALLOW_CREDENTIALS=true

# ✅ Correct
CORS_ALLOWED_ORIGINS=http://localhost:3000
CORS_ALLOW_CREDENTIALS=true
```

### Issue: Frontend can't reach backend

**Check:**
1. Backend is running: `./bin/api`
2. `NEXT_PUBLIC_API_URL` matches backend URL
3. `CORS_ALLOWED_ORIGINS` includes frontend URL
4. Check browser console for CORS errors

### Issue: Preflight requests failing

**Check:**
1. CORS headers are being set (use curl test above)
2. `CORS_ALLOW_CREDENTIALS` matches your needs
3. `CORS_MAX_AGE` is set (86400 recommended)

---

## File Structure

```
querybase/
├── .env                          # Backend environment variables
├── .env.example                  # Backend env template
├── config/
│   └── config.yaml              # Backend configuration (fallback)
├── internal/
│   ├── config/
│   │   └── config.go            # Config structure & parsing
│   └── api/middleware/
│       └── cors.go              # CORS middleware implementation
├── cmd/api/
│   └── main.go                  # API server (uses CORS config)
└── web/
    ├── .env.local               # Frontend environment variables
    ├── .env.local.example       # Frontend env template
    └── src/lib/
        └── api-client.ts        # API client (uses NEXT_PUBLIC_API_URL)
```

---

## Security Best Practices

### Development
```bash
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001
```

### Production
```bash
# ❌ Don't use wildcard
CORS_ALLOWED_ORIGINS=*

# ✅ Use specific domains
CORS_ALLOWED_ORIGINS=https://app.querybase.com

# ✅ Include www variant if needed
CORS_ALLOWED_ORIGINS=https://app.querybase.com,https://www.querybase.com
```

### Staging
```bash
CORS_ALLOWED_ORIGINS=https://staging.querybase.com,https://app.querybase.com
```

---

## Migration Notes

### Old Approach (Hardcoded)

```go
// ❌ Old way - hardcoded in main.go
corsConfig := middleware.DefaultConfig()
router.Use(middleware.CORSMiddleware(corsConfig))
```

### New Approach (Environment-based)

```go
// ✅ New way - from config
router.Use(middleware.CORSMiddlewareFromConfig(cfg.CORS))
```

---

## Quick Reference

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `CORS_ALLOWED_ORIGINS` | `http://localhost:3000,http://localhost:3001` | Comma-separated allowed origins |
| `CORS_ALLOW_CREDENTIALS` | `true` | Allow cookies/auth headers |
| `CORS_MAX_AGE` | `86400` | Preflight cache (seconds) |
| `NEXT_PUBLIC_API_URL` | `http://localhost:8080` | Backend URL for frontend |
