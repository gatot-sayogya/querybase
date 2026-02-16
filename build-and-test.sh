#!/bin/bash
set -e

echo "Running tests..."
# Run Go tests
go test ./...

# Run frontend check (optional, basic check)
cd web && npm ci && npm run lint && cd ..

echo "Building Docker image..."
docker build -t querybase:latest .

echo "Build complete. You can run the container with:"
echo "docker run -p 8080:8080 querybase:latest"
