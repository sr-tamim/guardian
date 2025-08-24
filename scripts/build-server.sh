#!/bin/bash
# Build script for production testing

echo "🏗️  Building Guardian for Windows Server Testing..."

# Clean previous builds
rm -rf bin/
mkdir -p bin/

# Get version information
VERSION="0.0.1"
if git rev-parse --git-dir > /dev/null 2>&1; then
    GIT_COMMIT=$(git rev-parse --short HEAD)
    if [ -n "$(git status --porcelain)" ]; then
        GIT_COMMIT="$GIT_COMMIT-dirty"
    fi
else
    GIT_COMMIT="unknown"
fi
BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)

# Build with version info
LDFLAGS="-X github.com/sr-tamim/guardian/pkg/version.Version=${VERSION}"
LDFLAGS="$LDFLAGS -X github.com/sr-tamim/guardian/pkg/version.GitCommit=${GIT_COMMIT}"
LDFLAGS="$LDFLAGS -X github.com/sr-tamim/guardian/pkg/version.BuildTime=${BUILD_TIME}"

echo "📦 Version: $VERSION"
echo "🔧 Git Commit: $GIT_COMMIT"
echo "🕐 Build Time: $BUILD_TIME"

# Build for Windows
echo "🖥️  Building Windows binary..."
GOOS=windows GOARCH=amd64 go build -ldflags "$LDFLAGS" -o bin/guardian.exe ./cmd/guardian

if [ $? -eq 0 ]; then
    echo "✅ Build successful: bin/guardian.exe"
    ls -la bin/guardian.exe
    echo ""
    echo "🚀 Ready for server deployment!"
    echo ""
    echo "📋 Next steps:"
    echo "   1. Copy bin/guardian.exe to your Windows server"
    echo "   2. Copy configs/guardian.yaml to the server" 
    echo "   3. Run as Administrator: guardian.exe monitor"
else
    echo "❌ Build failed"
    exit 1
fi
