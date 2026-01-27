#!/bin/bash

# QueryBase Build Script
# Builds binaries for multiple architectures

set -e

echo "🔨 QueryBase Multi-Architecture Build Script"
echo "============================================"
echo ""

# Detect current platform
OS="$(uname -s)"
ARCH="$(uname -m)"

# Map architecture names
case "$ARCH" in
    x86_64)
        NATIVE_ARCH="amd64"
        ;;
    aarch64|arm64)
        NATIVE_ARCH="arm64"
        ;;
    *)
        echo "⚠️  Unknown architecture: $ARCH"
        NATIVE_ARCH="amd64"
        ;;
esac

# Map OS names
case "$OS" in
    Linux)
        NATIVE_OS="linux"
        ;;
    Darwin)
        NATIVE_OS="darwin"
        ;;
    MINGW*|MSYS*|CYGWIN*)
        NATIVE_OS="windows"
        ;;
    *)
        echo "⚠️  Unknown OS: $OS"
        NATIVE_OS="linux"
        ;;
esac

echo "Detected platform: $NATIVE_OS/$NATIVE_ARCH"
echo ""

# Function to build a binary
build_binary() {
    local name=$1
    local cmd_path=$2
    local os=$3
    local arch=$4
    local ext=""

    if [ "$os" = "windows" ]; then
        ext=".exe"
    fi

    local output="bin/${name}-${os}-${arch}${ext}"
    local full_name="${name} (${os}/${arch})"

    echo "  → Building $full_name..."
    GOOS=$os GOARCH=$arch go build -o "$output" "$cmd_path"

    if [ $? -eq 0 ]; then
        size=$(ls -lh "$output" | awk '{print $5}')
        echo "    ✅ Built: $output ($size)"
    else
        echo "    ❌ Failed to build $full_name"
        return 1
    fi
}

# Parse arguments
BUILD_TARGET="${1:-all}"

# Create bin directory
mkdir -p bin

case "$BUILD_TARGET" in
    all)
        echo "Building for ALL platforms..."
        echo ""

        # Build API server
        echo "📦 Building API Server..."
        build_binary "api" "./cmd/api" "linux" "arm64"
        build_binary "api" "./cmd/api" "linux" "amd64"
        build_binary "api" "./cmd/api" "darwin" "arm64"
        build_binary "api" "./cmd/api" "darwin" "amd64"
        build_binary "api" "./cmd/api" "windows" "amd64"
        echo ""

        # Build Worker
        echo "📦 Building Worker..."
        build_binary "worker" "./cmd/worker" "linux" "arm64"
        build_binary "worker" "./cmd/worker" "linux" "amd64"
        build_binary "worker" "./cmd/worker" "darwin" "arm64"
        build_binary "worker" "./cmd/worker" "darwin" "amd64"
        build_binary "worker" "./cmd/worker" "windows" "amd64"
        echo ""

        echo "✅ Build complete! All binaries created in bin/"
        ;;

    native)
        echo "Building for NATIVE platform ($NATIVE_OS/$NATIVE_ARCH)..."
        echo ""

        build_binary "api" "./cmd/api" "$NATIVE_OS" "$NATIVE_ARCH"
        build_binary "worker" "./cmd/worker" "$NATIVE_OS" "$NATIVE_ARCH"
        echo ""

        echo "✅ Build complete! Native binaries created in bin/"
        ;;

    linux-arm64|linux-amd64|darwin-arm64|darwin-amd64|windows-amd64)
        OS_PART=$(echo $BUILD_TARGET | cut -d'-' -f1)
        ARCH_PART=$(echo $BUILD_TARGET | cut -d'-' -f2)

        echo "Building for $BUILD_TARGET..."
        echo ""

        build_binary "api" "./cmd/api" "$OS_PART" "$ARCH_PART"
        build_binary "worker" "./cmd/worker" "$OS_PART" "$ARCH_PART"
        echo ""

        echo "✅ Build complete! Binaries for $BUILD_TARGET created in bin/"
        ;;

    *)
        echo "❌ Invalid target: $BUILD_TARGET"
        echo ""
        echo "Usage: $0 [target]"
        echo ""
        echo "Targets:"
        echo "  all         - Build for all platforms (default)"
        echo "  native      - Build for current platform"
        echo "  linux-arm64 - Build for Linux ARM64"
        echo "  linux-amd64 - Build for Linux AMD64"
        echo "  darwin-arm64- Build for macOS ARM64 (Apple Silicon)"
        echo "  darwin-amd64- Build for macOS AMD64 (Intel)"
        echo "  windows-amd64- Build for Windows AMD64"
        echo ""
        exit 1
        ;;
esac

echo ""
echo "📋 Built binaries:"
ls -lh bin/ 2>/dev/null | tail -n +2 | while read -r line; do
    filename=$(echo "$line" | awk '{print $9}')
    size=$(echo "$line" | awk '{print $5}')
    if [ -n "$filename" ]; then
        echo "  $filename ($size)"
    fi
done
