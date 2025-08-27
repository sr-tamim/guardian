# Guardian Version History

## Version 0.0.3 (2025-01-XX) - Daemon Mode & Code Refactoring

### 🚀 **Major Architecture & Daemon Implementation**

#### ✨ **New Features - Daemon Mode**
- **True Background Daemon** - `guardian monitor -d` runs in background and returns immediately
- **Windows-Compatible Daemonization** - Proper process detachment using Windows APIs  
- **Daemon Process Management** - PID file management with cross-platform support
- **Daemon Detection** - Running `monitor` detects existing daemon and shows status instead of duplicate
- **Stop Command** - `guardian stop` cleanly terminates daemon processes
- **Status Integration** - Daemon status displayed in `guardian status` command
- **Log File Management** - Daemon logs to `%LOCALAPPDATA%\Guardian\logs\` on Windows
- **Internal Flag System** - `--daemon-internal` flag for spawned daemon processes

#### 🏗️ **Code Architecture Refactoring** 
- **Modular Command Structure** - Split 389-line main.go into organized command files
- **Command Package** - New `cmd/guardian/commands/` package with separate files:
  - `monitor.go` (144 lines) - Monitor & daemon functionality
  - `stop.go` (39 lines) - Daemon stop command  
  - `status.go` (31 lines) - Status information
  - `version.go` (33 lines) - Version details
  - `tui.go` (44 lines) - Terminal UI launcher
  - `root.go` (25 lines) - Root command definition
- **Sleek Main Entry Point** - Reduced main.go from 389 to 145 lines (63% reduction)
- **Lazy Configuration Loading** - Configuration loaded only when needed with caching
- **Clean Function Separation** - Single responsibility principle throughout

#### 🔧 **Technical Enhancements**
- **Cross-Platform Process Management** - Windows uses `tasklist`/`taskkill`, Unix uses signals
- **Process Group Creation** - Windows `CREATE_NEW_PROCESS_GROUP` for proper detachment  
- **File Handle Management** - Proper stdout/stderr redirection to log files
- **PID File System** - Robust PID management with stale file cleanup
- **Background Process Verification** - Daemon startup verification and health checks
- **Graceful Shutdown** - Signal handling for clean daemon termination

#### 🛠️ **Platform Compatibility**
- **Windows Daemon Support** - Full Windows backgrounding with `CREATE_NO_WINDOW`
- **Windows Process Detection** - `tasklist` integration for reliable process checking
- **Cross-Platform Log Paths** - Platform-specific log directory resolution
- **Windows-Compatible Termination** - `taskkill` for reliable process cleanup

## Version 0.0.2 (2025-01-XX) - Interactive TUI Dashboard

### 🎉 **Major Feature Release - Interactive Terminal Interface**

#### ✨ **New Features**
- **Interactive TUI Dashboard** - Full terminal user interface with Bubble Tea framework
- **Tab Navigation System** - Dashboard, Blocked IPs, Logs, Service, and Settings tabs
- **Real-time Service Management** - Start/stop controls via TUI interface
- **Live Statistics Display** - Blocked IPs, attack counts, and update timestamps
- **Platform Provider Integration** - Development/production mode indicators
- **Beautiful Terminal Styling** - Professional UI using Lip Gloss framework
- **Comprehensive Keyboard Shortcuts** - Tab navigation, refresh, service toggle, quit

#### 🔧 **Technical Enhancements**
- **TUI Command Integration** - New `guardian tui` command for interactive mode
- **Enhanced CLI Architecture** - Provider factory integration with TUI
- **Real-time Data Updates** - Live refresh functionality with provider communication
- **Cross-Platform TUI** - Works seamlessly on Windows with provider integration

#### 📚 **Documentation & Testing**
- **TUI Usage Guide** - Comprehensive README section with keyboard shortcuts
- **Interactive Demo Scripts** - Testing and demonstration utilities
- **Enhanced Help System** - Detailed TUI command documentation

#### 🛠️ **Dependencies Added**
- Bubble Tea framework (github.com/charmbracelet/bubbletea)
- Lip Gloss styling (github.com/charmbracelet/lipgloss)
- Bubbles components (github.com/charmbracelet/bubbles)

---

## Version 0.0.1 (2025-08-24)

### 🎉 **Initial Release - Windows-First Implementation**

#### ✨ **Core Features**
- **Windows RDP Monitoring** - Real-time Windows Security Event Log monitoring
- **Windows Firewall Integration** - Automatic IP blocking using `netsh advfirewall`
- **Cross-Platform Architecture** - Platform abstraction ready for Linux/macOS
- **Configurable Rules** - Template-based firewall rule naming
- **Automatic Cleanup** - Time-based rule expiration and removal

#### 🔧 **Windows-Specific Implementation**
- **Event ID 4625 Processing** - Failed RDP logon attempt detection
- **PowerShell Script Parity** - Matches production PowerShell script functionality
- **Regex Pattern Matching** - Exact IP extraction patterns from PowerShell
- **Admin Privilege Detection** - Graceful handling of permission requirements
- **Build Tag Architecture** - Platform-specific code compilation

#### 🛠️ **Technical Features**
- **Modern CLI Interface** - Cobra-based command structure
- **YAML Configuration** - Human-readable configuration files
- **Platform-Aware Paths** - Intelligent default path detection
- **Mock Provider** - Safe development and testing mode
- **Comprehensive Logging** - Detailed operation logging

#### 📦 **Build System**
- **Semantic Versioning** - Proper version information system
- **Cross-Platform Builds** - Windows/Linux/macOS binary generation
- **Build Metadata** - Git commit, build time, Go version tracking
- **Deployment Package** - Ready-to-deploy server packages

#### 🧪 **Testing & Deployment**
- **Server Testing Scripts** - PowerShell utilities for Windows Server
- **Prerequisites Checking** - Automated environment validation
- **Emergency Cleanup** - Firewall rule removal utilities
- **Documentation** - Complete testing and deployment guides

### 🎯 **Production Readiness**
- ✅ **Real Windows Event Log Integration**
- ✅ **Real Windows Firewall Management**
- ✅ **Administrator Privilege Support**
- ✅ **Configurable Thresholds and Durations**
- ✅ **Automatic Rule Cleanup**
- ✅ **Error Handling and Recovery**

### 🔮 **Planned Features (Next Releases)**
- **0.0.2**: Interactive TUI Mode with service management
- **0.0.3**: Linux platform implementation
- **0.0.4**: macOS platform support
- **0.1.0**: Advanced threat detection and ML integration

### 📊 **Specifications**
- **Binary Size**: ~8.4MB (single executable)
- **Memory Usage**: <50MB runtime
- **Platforms**: Windows (x64)
- **Dependencies**: None (statically linked)
- **Go Version**: 1.25.0

### 🚀 **Migration from PowerShell**
This release provides 1:1 feature parity with existing PowerShell-based intrusion prevention scripts while offering:
- **Better Performance** - Native Go vs PowerShell execution
- **Modern Architecture** - Extensible cross-platform design
- **Enhanced Monitoring** - Real-time event processing
- **Improved Reliability** - Robust error handling and recovery
- **Future-Proof** - Ready for advanced features and platforms
