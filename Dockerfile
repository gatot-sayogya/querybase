# Stage 1: Build Frontend
FROM node:18-alpine AS frontend-builder
WORKDIR /app/web

# Copy package.json and package-lock.json
COPY web/package.json web/package-lock.json ./

# Install dependencies
RUN npm ci

# Copy frontend source
COPY web/ ./

# Build frontend (config is already set to output: export)
RUN npm run build

# Stage 2: Build Backend
FROM golang:1.23-alpine AS backend-builder
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git make

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy backend source
COPY . .

# Build backend
# CGO_ENABLED=0 for static binary
RUN CGO_ENABLED=0 GOOS=linux go build -o api cmd/api/main.go

# Stage 3: Final Image
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies if needed
RUN apk add --no-cache ca-certificates tzdata

# Copy backend binary
COPY --from=backend-builder /app/api .

# Copy config directory (assuming config.yaml is needed)
# You might want to use env vars in production instead of file
COPY --from=backend-builder /app/config ./config

# Copy frontend static files to web/out
COPY --from=frontend-builder /app/web/out ./web/out

# Expose port
EXPOSE 8080

# Environment variables
ENV GIN_MODE=release
ENV PORT=8080

# Run the binary
CMD ["./api"]
