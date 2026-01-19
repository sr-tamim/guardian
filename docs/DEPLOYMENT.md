# Deployment Guide (Windows)

## Build

```bash
# Build directly
go build -o bin/guardian.exe ./cmd/guardian

# Or use the build script (adds version metadata)
./scripts/build-server.sh
```

## Configuration

Default config lookup order (no --config flag):
- ./configs/guardian.yaml
- ./guardian.yaml
- %APPDATA%\Guardian\guardian.yaml

Recommended: copy and edit configs/guardian.yaml on the server.

## Run as daemon

```bash
# Start daemon (background)
./guardian.exe monitor -d --config "C:\path\to\guardian.yaml"

# Start daemon with system tray (interactive desktop session)
./guardian.exe monitor -d --tray --config "C:\path\to\guardian.yaml"
```

## Run as Windows Service (recommended)

```bash
# Install and start the service (config path captured at install time)
./guardian.exe --config "C:\path\to\guardian.yaml" service install
./guardian.exe service start

# Check status
./guardian.exe service status

# Stop and uninstall
./guardian.exe service stop
./guardian.exe service uninstall
```

If you need to change the config path, uninstall and re-install the service with the new --config.

## Logs

- Daemon logs (stdout/stderr redirection): %LOCALAPPDATA%\Guardian\logs
- Application logs (when enable_file=true): configured file_path in YAML
- Windows Service errors: Windows Event Viewer → Windows Logs → Application (Source: Guardian)
