# Guardian Development Roadmap

## 🎯 **Current Status: v0.0.1**
✅ **Windows-First Implementation Complete**
- Windows RDP monitoring with PowerShell script parity
- Windows Firewall integration
- Cross-platform architecture foundation
- Proper versioning system
- Server testing and deployment tools

---

## 🗺️ **Upcoming Releases**

### **📋 Step 5: v0.0.2 - Interactive TUI Mode**
**Target: Next Release**

#### **Core TUI Features**
- **Double-click launcher** - GUI-friendly executable behavior
- **Interactive dashboard** - Real-time attack visualization  
- **Service management** - Start/stop background service via UI
- **Background operation** - Service runs independently of TUI window
- **Status monitoring** - Live blocked IPs, attack counts, service status

#### **Desktop Integration**
- **Windows Service mode** - Install/uninstall Windows Service
- **System tray integration** - Minimize to tray, show notifications
- **Auto-start capability** - Launch at system boot
- **Emergency controls** - Quick stop/unblock features

#### **Technical Implementation**
- **Bubble Tea TUI framework** - Modern terminal UI
- **Service communication** - IPC between TUI and background service
- **State persistence** - Remember service state across restarts
- **Process management** - Proper service lifecycle handling

---

### **🐧 Step 6: v0.0.3 - Linux Platform Support**
**Target: Follow-up Release**

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

### **🍎 Step 7: v0.0.4 - macOS Platform Support**
**Target: Full Cross-Platform**

#### **macOS-Specific Features**
- **pfctl integration** - macOS firewall management
- **launchd service** - Native macOS service system
- **FSEvents monitoring** - macOS file system events
- **Homebrew packaging** - Easy installation via brew

---

### **🧠 Step 8: v0.1.0 - Advanced Intelligence**
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

## 🔥 **Step 5 Focus: TUI Implementation Plan**

### **Priority 1: Double-Click Launcher**
```go
// Check if launched without arguments (double-click)
if len(os.Args) == 1 {
    // Launch interactive TUI mode
    startTUI()
} else {
    // Normal CLI mode
    rootCmd.Execute()
}
```

### **Priority 2: Service Management**
```
Main Menu:
┌─────────────────────────────────────┐
│  🛡️  Guardian v0.0.2                 │
│  ═══════════════════════════════════ │
│                                     │
│  Service Status: ●  Running         │
│  Active Blocks:      3 IPs          │
│  Last Activity:      2 min ago      │
│                                     │
│  [S] Start Service                  │
│  [T] Stop Service                   │
│  [V] View Blocked IPs               │
│  [L] View Live Attacks              │
│  [C] Configuration                  │
│  [Q] Quit                           │
└─────────────────────────────────────┘
```

### **Priority 3: Background Service**
- **Windows Service** - Install as actual Windows Service
- **Service communication** - Named pipes or TCP for IPC
- **Persistent operation** - Continue running after TUI close
- **Auto-recovery** - Restart on crash or system reboot

---

## 📊 **Version Comparison Matrix**

| Feature | v0.0.1 | v0.0.2 | v0.0.3 | v0.1.0 |
|---------|--------|--------|--------|--------|
| **Windows Support** | ✅ CLI | ✅ TUI | ✅ TUI | ✅ TUI |
| **Linux Support** | ❌ | ❌ | ✅ Full | ✅ Full |
| **macOS Support** | ❌ | ❌ | ❌ | ✅ Full |
| **Interactive UI** | ❌ CLI Only | ✅ TUI | ✅ TUI | ✅ Web + TUI |
| **Background Service** | ❌ | ✅ Windows | ✅ All | ✅ All |
| **Double-click Launch** | ❌ | ✅ | ✅ | ✅ |
| **ML Detection** | ❌ | ❌ | ❌ | ✅ |
| **Multi-server** | ❌ | ❌ | ❌ | ✅ |

---

## 🎯 **Success Metrics**

### **v0.0.2 Success Criteria**
- ✅ Double-click launches TUI (no command line needed)
- ✅ Can start/stop service from TUI
- ✅ TUI window can close while service runs
- ✅ Re-opening shows current service status
- ✅ Windows Service integration works
- ✅ All v0.0.1 functionality preserved

### **Development Timeline**
- **TUI Framework Setup**: 2-3 days
- **Service Architecture**: 3-4 days  
- **Windows Service Integration**: 2-3 days
- **Testing & Polish**: 2-3 days
- **Total Estimate**: 10-12 days

**Next: Step 5 - Interactive TUI Implementation** 🚀
