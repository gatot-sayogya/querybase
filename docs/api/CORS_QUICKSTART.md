# CORS Configuration - Quick Start

## What Changed?

**Before (Hardcoded):**
```go
// cmd/api/main.go
corsConfig := middleware.DefaultConfig()  // Hardcoded origins
router.Use(middleware.CORSMiddleware(corsConfig))
```

**After (Environment-based):**
```go
// cmd/api/main.go
router.Use(middleware.CORSMiddlewareFromConfig(cfg.CORS))  // From config
```

---

## Setup in 3 Steps

### Step 1: Configure Backend (.env)

Create `.env` in project root:

```bash
# Frontend URL(s) - comma separated
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001

# Optional settings with defaults
CORS_ALLOW_CREDENTIALS=true
CORS_MAX_AGE=86400
```

### Step 2: Configure Frontend (web/.env.local)

Create `web/.env.local`:

```bash
# Backend API URL
NEXT_PUBLIC_API_URL=http://localhost:8080
```

### Step 3: Restart Services

```bash
# Stop existing services
pkill -f "./bin/api"
pkill -f "./bin/worker"

# Start API
./bin/api

# Start worker (in another terminal)
./bin/worker
```

---

## Configuration Options

| Option | Environment Variable | Default | Description |
|--------|---------------------|---------|-------------|
| Allowed Origins | `CORS_ALLOWED_ORIGINS` | `http://localhost:3000,http://localhost:3001` | Comma-separated list of allowed frontend URLs |
| Allow Credentials | `CORS_ALLOW_CREDENTIALS` | `true` | Allow cookies and auth headers |
| Max Age | `CORS_MAX_AGE` | `86400` | Preflight cache duration (24 hours) |

---

## Priority Order

1. **Environment Variables** (`.env`) - Highest priority
2. **Config File** (`config/config.yaml`) - Fallback
3. **Code Defaults** - Last resort

---

## Quick Examples

### Local Development
```bash
# .env
CORS_ALLOWED_ORIGINS=http://localhost:3000

# web/.env.local
NEXT_PUBLIC_API_URL=http://localhost:8080
```

### Production
```bash
# .env
BACKEND_URL=https://api.querybase.com
FRONTEND_URL=https://app.querybase.com
CORS_ALLOWED_ORIGINS=https://app.querybase.com

# web/.env.local
NEXT_PUBLIC_API_URL=https://api.querybase.com
```

### Multiple Frontend Domains
```bash
# .env
CORS_ALLOWED_ORIGINS=https://app1.example.com,https://app2.example.com
```

---

## Files Modified

### Backend (Go)
- ✅ `internal/config/config.go` - Added `CORSConfig` struct
- ✅ `internal/api/middleware/cors.go` - Added `CORSMiddlewareFromConfig()`
- ✅ `cmd/api/main.go` - Updated to use config-based CORS
- ✅ `config/config.yaml` - Added CORS configuration section
- ✅ `.env.example` - Updated with CORS environment variables

### Frontend (Next.js)
- ✅ `web/.env.local.example` - Enhanced with better comments

### Documentation
- ✅ `docs/CORS_SETUP.md` - Comprehensive CORS guide
- ✅ `docs/CORS_QUICKSTART.md` - This quick start guide

---

## Testing

### Test CORS is Working

```bash
curl -X OPTIONS http://localhost:8080/api/v1/queries \
  -H "Origin: http://localhost:3000" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: authorization,content-type" \
  -v 2>&1 | grep "Access-Control"
```

**Expected Output:**
```
< Access-Control-Allow-Origin: http://localhost:3000
< Access-Control-Allow-Methods: GET, POST, PUT, DELETE, PATCH, OPTIONS, HEAD
< Access-Control-Allow-Headers: Origin, Content-Type, Accept, Authorization
< Access-Control-Max-Age: 86400
< Access-Control-Allow-Credentials: true
```

---

## Troubleshooting

**Issue:** "Origin not allowed" error
```bash
# Solution: Add your origin to CORS_ALLOWED_ORIGINS
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://your-domain.com
```

**Issue:** Frontend can't connect to backend
```bash
# Check 1: Verify NEXT_PUBLIC_API_URL is correct
echo $NEXT_PUBLIC_API_URL

# Check 2: Verify backend is running
curl http://localhost:8080/health

# Check 3: Check CORS_ALLOWED_ORIGINS includes frontend URL
# (see .env or config.yaml)
```

---

## Need More Details?

See the comprehensive guide: [CORS_SETUP.md](./CORS_SETUP.md)
