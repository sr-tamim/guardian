# Architecture

## Platform Abstraction

Guardian uses a platform factory to select providers based on runtime OS.

```
Application (CLI/TUI)
  → Platform Factory (GOOS)
    → Windows Provider
    → Mock Provider
```

## Windows Implementation

```
Guardian TUI
  → Daemon Manager
    → Windows Provider
      → Event Log Monitor (4625)
      → Firewall Manager (netsh)
```

Service mode (Windows):

```
Windows Service
  → Service Manager
    → Daemon Manager
      → Windows Provider
```

## Design Principles
- Interface segregation (`PlatformProvider`)
- Dependency injection for testability
- Build tags for OS-specific code
- Mock provider for safe development
- Service mode shares the same monitoring pipeline as daemon mode
