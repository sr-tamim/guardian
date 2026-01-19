# Guardian Development Roadmap

## 🎯 **Current Status: v0.0.3**
✅ **Architecture Refactoring & Daemon Mode Complete**

## 📊 **Version Comparison Matrix**

| Feature | v0.0.1-3 | v0.0.4 | v0.0.5 | v0.0.6 | v0.1.0 |
|---------|----------|--------|--------|--------|--------|
| **Windows Foundation** | ✅ Complete | ✅ Complete | ✅ Complete | ✅ Complete | ✅ Complete |
| **Windows Core Logic** | ❌ Missing | ✅ Complete | ✅ Complete | ✅ Complete | ✅ Complete |
| **Linux Support** | ❌ | ❌ | ✅ Service | ✅ Full | ✅ Full |
| **Interactive UI** | ✅ TUI | ✅ TUI | ✅ TUI | ✅ TUI | ✅ Web + TUI |
| **Background Service** | ✅ Windows | ✅ Windows | ✅ All | ✅ All | ✅ All |
| **Attack Detection** | ❌ Missing | ✅ Complete | ✅ Complete | ✅ Complete | ✅ Complete |
| **Auto Blocking** | ❌ Missing | ✅ Complete | ✅ Complete | ✅ Complete | ✅ Complete |
| **ML Detection** | ❌ | ❌ | ❌ | ❌ | ✅ |
| **Multi-server** | ❌ | ❌ | ❌ | ❌ | ✅ |

✅ **Modular Command Structure Complete (v0.0.3)**
- Improved command structure with separated files (63% main.go reduction)
- True background daemon mode with Windows support
- PID management and daemon process control
- Cross-platform daemon implementation
- System tray integration (Windows)
- Professional code organization and maintainability

✅ **Windows Service Reliability (v0.0.3)**
- Windows Service command support (install/start/stop/status)
- Service-run mode for headless operation

✅ **Interactive TUI Dashboard Complete (v0.0.2)**
- Interactive TUI dashboard with Bubble Tea framework
- Real-time service management and monitoring
- Beautiful terminal styling and keyboard navigation
- Cross-platform GUI-friendly operation

🔄 **Windows Platform Foundation Complete, Core Logic In Development (v0.0.1)**
- ✅ Windows Event Log monitoring (Event ID 4625 detection)
- ✅ Windows Firewall integration (`netsh advfirewall`)
- ✅ Event parsing and IP extraction
- ✅ Firewall rule creation/deletion
- ✅ Cross-platform architecture foundation
- ✅ Proper versioning system
- ✅ Server testing and deployment tools
- ✅ **Attack detection logic** (threshold counting per IP, Windows)
- ✅ **Threat assessment system** (whitelist checks, Windows)
- ✅ **Event processing pipeline** (monitoring → detection → blocking, Windows)
- ❌ **Missing: Persistent storage** (SQLite integration)

---

## 🗺️ **Upcoming Releases**

### **🛡️ Priority: v0.0.4 - Windows Core Logic Completion**
**Target: IMMEDIATE - Complete Windows Platform**

#### **Critical Missing Components**
- **Persistent Storage** - SQLite integration for attack/block records
- **Integration Layer** - Wire all components together in main monitoring loop

#### **Immediate Next: 4625 Config-Driven Detection**
- **Config-driven Event ID 4625 filtering** - Use services config to control Windows Security log monitoring
- **Flexible service selection** - Enable/disable Windows log sources via YAML
- **Parity with PowerShell** - Ensure 4625 counts include all logon types when configured

#### **Implementation Priority**
1. **Storage persistence** - Save attack attempts and block records
2. **Full end-to-end testing** - Real attack detection and blocking

---

### **🔧 Step 5: v0.0.5 - System Service Integration**
**Target: Next Release**

#### **Service Management**
- **Windows Service** - Install/uninstall as Windows Service
- **Systemd integration** - Linux service management
- **Auto-start capability** - Launch at system boot  
- **Service status integration** - Show service state in TUI/CLI
- **Service logs** - Integration with system logging

#### **Linux Reliability Plan**
- **systemd unit + installer** - Provide unit file and installation helper
- **Journal logging** - Ensure logs are available via `journalctl`
- **Restart policy** - Configure `Restart=always` with backoff

#### **Enhanced Process Management**
- **Service lifecycle** - Install, start, stop, uninstall commands
- **Process monitoring** - Health checks and automatic restart
- **Resource management** - Memory and CPU monitoring
- **Signal handling** - Graceful shutdown and restart

#### **Desktop Integration**
- **Emergency controls** - Quick stop/unblock features via tray
- **Desktop notifications** - Attack alerts and status updates

#### **Future: SQL Server Detection**
- **SQL Server failed login detection** - Add parser for SQL error log or Windows Application events
- **Config-driven source** - Allow selecting SQL log source/path in YAML

---

### **🐧 Step 6: v0.0.6 - Linux Platform Support**
**Target: After Windows Completion**

#### **Linux-Specific Features**
- **iptables integration** - Native Linux firewall support
- **systemd service** - Proper Linux service management
- **SSH log parsing** - `/var/log/auth.log` and `/var/log/secure`
- **inotify monitoring** - Real-time log file watching

#### **Cross-Platform Validation**
- **Build matrix** - Windows + Linux builds
- **Platform feature parity** - Consistent functionality
- **Configuration compatibility** - Shared config format

---

### **🧠 Step 7: v0.1.0 - Advanced Intelligence**
**Target: Major Feature Release**

#### **AI/ML Integration**
- **Pattern recognition** - ML-based attack detection
- **Behavioral analysis** - User behavior anomaly detection
- **Geographic filtering** - IP geolocation-based blocking
- **Threat intelligence** - External threat feed integration

#### **Advanced Features**
- **Web dashboard** - Browser-based management interface
- **REST API** - Programmatic access and integration
- **Multi-server coordination** - Centralized management
- **Advanced reporting** - Attack analytics and trends

---

##  **Version Comparison Matrix**

| Feature | v0.0.1-3 | v0.0.4 | v0.0.5 | v0.0.6 | v0.1.0 |
|---------|----------|--------|--------|--------|--------|
| **Windows Foundation** | ✅ Complete | ✅ Complete | ✅ Complete | ✅ Complete | ✅ Complete |
| **Windows Core Logic** | ❌ Missing | ✅ Complete | ✅ Complete | ✅ Complete | ✅ Complete |
| **Linux Support** | ❌ | ❌ | ✅ Service | ✅ Full | ✅ Full |
| **Interactive UI** | ✅ TUI | ✅ TUI | ✅ TUI | ✅ TUI | ✅ Web + TUI |
| **Background Service** | ✅ Windows | ✅ Windows | ✅ All | ✅ All | ✅ All |
| **Attack Detection** | ❌ Missing | ✅ Complete | ✅ Complete | ✅ Complete | ✅ Complete |
| **Auto Blocking** | ❌ Missing | ✅ Complete | ✅ Complete | ✅ Complete | ✅ Complete |
| **ML Detection** | ❌ | ❌ | ❌ | ❌ | ✅ |
| **Multi-server** | ❌ | ❌ | ❌ | ❌ | ✅ |

---

## 🎯 **Success Metrics**

### **v0.0.3 Current Status**
- ✅ Architecture refactoring complete (modular commands)
- ✅ TUI dashboard functional and beautiful
- ✅ Daemon mode working with PID management
- ✅ Windows Event Log monitoring implemented
- ✅ Windows Firewall integration implemented
- ✅ Attack detection logic (threshold-based, Windows)
- ✅ Event processing pipeline (monitoring → detection → blocking, Windows)
- ❌ **Missing: Persistent storage** (all data lost on restart)

### **v0.0.4 Success Criteria (Windows Completion)**
- ✅ Attack counting per IP address
- ✅ Threshold-based automatic blocking
- ❌ Persistent storage of attacks and blocks
- ✅ IP whitelist functionality
- ✅ End-to-end attack detection and blocking
- ✅ Can detect real RDP attacks and block automatically
- ❌ Survives restarts with persistent data

### **Development Timeline for v0.0.4**
- **Attack Detection Engine**: 3-4 days
- **Event Processing Pipeline**: 2-3 days  
- **SQLite Storage Integration**: 2-3 days
- **Whitelist & Configuration**: 1-2 days
- **End-to-End Testing**: 2-3 days
- **Total Estimate**: 10-15 days

**Next Priority: Complete Windows Core Logic**
