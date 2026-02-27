#!/bin/bash
# QueryBase Dev Startup Script
# Starts the Go API + Next.js frontend, connecting to Docker-hosted databases.
# Clears Next.js cache on every start to prevent stale CSS/asset 404s.
# Both services are health-checked before printing the "ready" banner.

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

echo "==================================="
echo "  QueryBase Dev Startup"
echo "==================================="

# Kill any existing processes on 8080 and 3000
echo "Clearing ports 8080 and 3000..."
lsof -ti :8080 | xargs kill -9 2>/dev/null || true
lsof -ti :3000 | xargs kill -9 2>/dev/null || true
sleep 1  # Give OS time to release the ports

# Load .env then override hosts for local (non-Docker) run
set -a
source "$SCRIPT_DIR/.env"
set +a

export DATABASE_HOST=localhost
export REDIS_HOST=localhost

# ── 1. Start Go API ──────────────────────────────────────────────────────────
echo ""
echo "Starting Go API on :8080..."
go run "$SCRIPT_DIR/cmd/api/main.go" &
API_PID=$!

echo "Waiting for API to be ready..."
for i in $(seq 1 30); do
    if curl -s http://localhost:8080/health > /dev/null 2>&1; then
        echo "✅ API is ready"
        break
    fi
    sleep 1
    if [ "$i" -eq 30 ]; then
        echo "❌ API did not start in time"
        kill "$API_PID" 2>/dev/null
        exit 1
    fi
done

# Auto-fix datasource passwords after every startup
echo ""
echo "Re-encrypting datasource passwords..."
bash "$SCRIPT_DIR/scripts/fix-datasource-passwords.sh" || true

# ── 2. Start Next.js frontend ─────────────────────────────────────────────────
echo ""
echo "Starting Next.js frontend on :3000..."
cd "$SCRIPT_DIR/web"

# Always clear the Next.js cache to prevent stale CSS / 404 asset issues.
# This happens when prod builds (npm run build) overwrite the dev cache,
# or when the dev server is restarted without cleaning up.
echo "Clearing Next.js cache (.next)..."
rm -rf .next

npm run dev &
FRONTEND_PID=$!

# Wait for Next.js to finish its initial compilation and begin serving pages.
# We poll port 3000 the same way we poll the API.
echo "Waiting for frontend to be ready (initial compilation may take ~15s)..."
for i in $(seq 1 60); do
    if curl -s -o /dev/null -w "%{http_code}" http://localhost:3000 2>/dev/null | grep -qE '^[23]'; then
        echo "✅ Frontend is ready"
        break
    fi
    sleep 1
    if [ "$i" -eq 60 ]; then
        echo "⚠️  Frontend is taking longer than expected — check for errors above."
        # Don't exit; the server may still be compiling large pages.
    fi
done

# ── 3. Ready banner ───────────────────────────────────────────────────────────
echo ""
echo "==================================="
echo "  ✅ Dev environment ready!"
echo "  Frontend: http://localhost:3000"
echo "  Backend:  http://localhost:8080"
echo "  Press Ctrl+C to stop both"
echo "==================================="

# Cleanup both on exit
trap "echo ''; echo 'Stopping...'; kill $API_PID $FRONTEND_PID 2>/dev/null; exit 0" INT TERM

# Keep script alive until one of the processes exits
wait $API_PID $FRONTEND_PID
