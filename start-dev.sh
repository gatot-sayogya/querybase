#!/bin/bash
# QueryBase Dev Startup Script
# Starts the Go API locally, connecting to Docker-hosted databases.
# Automatically re-encrypts datasource passwords after startup.

set -e

echo "==================================="
echo "  QueryBase Dev Startup"
echo "==================================="

# Kill existing API process on 8080 if any
echo "Clearing port 8080..."
lsof -ti :8080 | xargs kill -9 2>/dev/null || true

# Load .env then override hosts for local (non-Docker) run
set -a
source "$(dirname "$0")/.env"
set +a

export DATABASE_HOST=localhost
export REDIS_HOST=localhost

echo "Starting Go API on :8080..."

# Start API in background
go run ./cmd/api/main.go &
API_PID=$!

# Wait for API to become healthy (max 30s)
echo "Waiting for API to be ready..."
for i in $(seq 1 30); do
    if curl -s http://localhost:8080/health > /dev/null 2>&1; then
        echo "✅ API is ready"
        break
    fi
    sleep 1
    if [ $i -eq 30 ]; then
        echo "❌ API did not start in time"
        kill $API_PID 2>/dev/null
        exit 1
    fi
done

# Auto-fix datasource passwords after every startup
echo ""
echo "Re-encrypting datasource passwords..."
bash "$(dirname "$0")/scripts/fix-datasource-passwords.sh" || true

echo ""
echo "==================================="
echo "  Dev environment ready!"
echo "  Frontend: http://localhost:3000"
echo "  Backend:  http://localhost:8080"
echo "  Press Ctrl+C to stop both"
echo "==================================="

# Start Next.js frontend in background
echo "Starting Next.js frontend on :3000..."
cd "$(dirname "$0")/web" && npm run dev &
FRONTEND_PID=$!

# Cleanup both on exit
trap "kill $API_PID $FRONTEND_PID 2>/dev/null; exit 0" INT TERM

# Wait for either process to exit
wait $API_PID $FRONTEND_PID
