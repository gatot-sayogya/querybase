# QueryBase Build Guide

This document explains how to build QueryBase for different architectures.

## Quick Start

### Build for Native Platform
```bash
# Using Make
make build

# Using build script
./build.sh native
```

### Build for All Platforms
```bash
# Using Make (builds for linux, darwin, windows - both arm64 and amd64)
make build-all

# Using build script
./build.sh all
```

## Build for Specific Architecture

### Using Make
```bash
# Build API server for all architectures
make build-api-multi

# Build worker for all architectures
make build-worker-multi

# Or build for native only
make build-api
make build-worker
```

### Using Build Script
```bash
# Native platform (automatic detection)
./build.sh native

# Specific platforms
./build.sh linux-arm64
./build.sh linux-amd64
./build.sh darwin-arm64
./build.sh darwin-amd64
./build.sh windows-amd64
```

## Available Binaries

After building, you'll find the following binaries in the `bin/` directory:

### API Server
- `api` - Native architecture binary
- `api-linux-arm64` - Linux ARM64 (aarch64)
- `api-linux-amd64` - Linux AMD64 (x86_64)
- `api-darwin-arm64` - macOS ARM64 (Apple Silicon M1/M2/M3)
- `api-darwin-amd64` - macOS AMD64 (Intel)
- `api-windows-amd64.exe` - Windows AMD64

### Worker
- `worker` - Native architecture binary
- `worker-linux-arm64` - Linux ARM64
- `worker-linux-amd64` - Linux AMD64
- `worker-darwin-arm64` - macOS ARM64
- `worker-darwin-amd64` - macOS AMD64
- `worker-windows-amd64.exe` - Windows AMD64

## Listing Binaries

```bash
make list
```

This will show all built binaries with their file sizes.

## Cleaning Build Artifacts

```bash
make clean
```

This removes all binaries from the `bin/` directory.

## Architecture Detection

The build script automatically detects your current platform:

- **OS**: Linux, Darwin (macOS), or Windows
- **Architecture**: ARM64 or AMD64

When using `./build.sh native`, it builds for your detected platform.

## Cross-Compilation Requirements

Go supports cross-compilation out of the box. However, for some platforms you may need:

### Linux (no requirements)
Go can compile for Linux from any platform without additional dependencies.

### macOS (no requirements)
Go can compile for macOS from any platform without additional dependencies.

### Windows (no requirements)
Go can compile for Windows from any platform without additional dependencies.

## Running the Binaries

### Linux / macOS
```bash
# Make executable (if needed)
chmod +x bin/api-linux-amd64

# Run
./bin/api-linux-amd64
```

### Windows
```cmd
bin\api-windows-amd64.exe
```

## Build Script Options

```bash
./build.sh [target]

Targets:
  all           - Build for all platforms (linux, darwin, windows)
  native        - Build for current platform (auto-detected)
  linux-arm64   - Build for Linux ARM64
  linux-amd64   - Build for Linux AMD64
  darwin-arm64  - Build for macOS ARM64 (Apple Silicon)
  darwin-amd64  - Build for macOS AMD64 (Intel)
  windows-amd64 - Build for Windows AMD64
```

## Docker Deployment

When deploying to Docker, use the Linux binaries:

```dockerfile
FROM alpine:latest
COPY bin/api-linux-amd64 /usr/local/bin/querybase-api
COPY bin/worker-linux-amd64 /usr/local/bin/querybase-worker
```

For ARM64 Docker hosts:
```dockerfile
FROM alpine:latest
COPY bin/api-linux-arm64 /usr/local/bin/querybase-api
COPY bin/worker-linux-arm64 /usr/local/bin/querybase-worker
```

## Troubleshooting

### Build fails with "command not found"
Make sure you have Go installed:
```bash
go version
```

### Binaries have wrong architecture
Check what you built:
```bash
file bin/api
```

Expected output for different platforms:
- Linux ARM64: `ELF 64-bit LSB executable, ARM aarch64`
- Linux AMD64: `ELF 64-bit LSB executable, x86-64`
- macOS ARM64: `Mach-O 64-bit executable arm64`
- macOS AMD64: `Mach-O 64-bit executable x86_64`
- Windows AMD64: `PE32+ executable (console) x86-64, for MS Windows`

### Permission denied when running binary
```bash
chmod +x bin/api-*
```

## Make Commands Reference

| Command | Description |
|---------|-------------|
| `make help` | Show all available commands |
| `make build` | Build for native platform |
| `make build-all` | Build for all platforms |
| `make build-api` | Build API server (native) |
| `make build-worker` | Build worker (native) |
| `make build-api-multi` | Build API server (all platforms) |
| `make build-worker-multi` | Build worker (all platforms) |
| `make list` | List built binaries |
| `make clean` | Remove build artifacts |
| `make run-api` | Run API server directly |
| `make run-worker` | Run worker directly |
