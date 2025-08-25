#!/bin/bash

# Guardian v0.0.2 TUI Testing Script
# Tests the interactive Terminal User Interface

set -e

echo "🧪 Guardian v0.0.2 TUI Testing Suite"
echo "===================================="
echo

# Build the latest version
echo "📦 Building Guardian with TUI..."
go build -o bin/guardian.exe ./cmd/guardian
echo "✅ Build completed"
echo

# Test TUI command availability
echo "🔍 Testing TUI command availability..."
./bin/guardian.exe --help | grep -q "tui.*Launch interactive TUI dashboard"
if [ $? -eq 0 ]; then
    echo "✅ TUI command found in help"
else
    echo "❌ TUI command not found in help"
    exit 1
fi

# Test TUI help
echo "📚 Testing TUI help..."
./bin/guardian.exe tui --help | grep -q "Launch the Guardian interactive terminal user interface"
if [ $? -eq 0 ]; then
    echo "✅ TUI help text correct"
else
    echo "❌ TUI help text incorrect"
    exit 1
fi

# Test development mode availability
echo "🔧 Testing development mode compatibility..."
./bin/guardian.exe --dev tui --help > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✅ Development mode TUI compatible"
else
    echo "❌ Development mode TUI not working"
    exit 1
fi

echo
echo "🎉 All TUI tests passed!"
echo
echo "💡 Manual Test Instructions:"
echo "   1. Run: ./bin/guardian.exe --dev tui"
echo "   2. Test navigation with Tab/Shift+Tab"
echo "   3. Test service toggle with 's' key"
echo "   4. Test refresh with 'r' key"
echo "   5. Test quit with 'q' key"
echo
echo "🏁 TUI v0.0.2 is ready for interactive testing!"
