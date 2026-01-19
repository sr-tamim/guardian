# Usage Guide

## Quick start

```bash
# Show version
./guardian.exe --version

# Launch TUI
./guardian.exe tui

# Development mode (no real blocking)
./guardian.exe --dev monitor
```

## Daemon mode

```bash
# Start daemon
./guardian.exe monitor -d --config "C:\path\to\guardian.yaml"

# Check daemon status
./guardian.exe status

# Stop daemon
./guardian.exe stop
```

## Service mode (Windows)

```bash
./guardian.exe --config "C:\path\to\guardian.yaml" service install
./guardian.exe service start
./guardian.exe service status
```

## Autostart (user login)

```bash
./guardian.exe autostart enable
./guardian.exe autostart status
```
