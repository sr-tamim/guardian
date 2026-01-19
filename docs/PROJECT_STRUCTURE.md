# Project Structure

```
guardian/
├── cmd/guardian/           # CLI entry point and commands
├── internal/
│   ├── platform/           # Platform providers
│   ├── daemon/             # Daemon management
│   ├── autostart/          # Auto-start integration
│   ├── parser/             # Log parsers
│   └── core/               # Interfaces and errors
├── pkg/                    # Public packages (models, logger, utils)
├── configs/                # YAML configs
└── scripts/                # Build/test scripts
```
