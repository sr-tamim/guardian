#!/bin/bash

# Guardian TUI Demo Script
# Quick demo of the v0.0.2 TUI functionality

echo "🛡️  Guardian v0.0.2 TUI Demo"
echo "==========================="
echo
echo "✅ Guardian TUI is ready!"
echo
echo "🎯 Quick TUI Demo:"
echo "   • Interactive terminal dashboard"
echo "   • Tab navigation (Tab/Shift+Tab)"
echo "   • Service management (press 's' in Service tab)"
echo "   • Real-time refresh (press 'r')"
echo "   • Development/Production mode indicators"
echo "   • Platform provider integration"
echo
echo "🚀 To launch TUI:"
echo "   Development mode: ./bin/guardian.exe --dev tui"
echo "   Production mode:  ./bin/guardian.exe tui"
echo
echo "⌨️  TUI Controls:"
echo "   Tab/Shift+Tab - Navigate between tabs"
echo "   r - Refresh data"
echo "   s - Toggle service (in Service tab)"
echo "   q - Quit"
echo
echo "📋 Available tabs:"
echo "   1. Dashboard - Overview and statistics"
echo "   2. Blocked IPs - Currently blocked addresses"
echo "   3. Logs - Recent activity logs"
echo "   4. Service - Service management controls"
echo "   5. Settings - Configuration options"
echo

# Display current version info
./bin/guardian.exe version
