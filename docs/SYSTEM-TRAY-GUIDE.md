# Guardian System Tray & Auto-Startup Implementation

## 🎯 **New Features in v0.0.2 Enhanced**

### **System Tray Integration**
Guardian now runs with full system tray support, allowing background operation while keeping the user interface accessible.

#### **Key Features:**
- ✅ **Minimize to Tray** - Closing TUI minimizes to system tray instead of exiting
- ✅ **Background Monitoring** - Service continues running even when TUI is closed
- ✅ **Tray Menu** - Right-click access to all Guardian functions
- ✅ **Auto-startup** - Integrated Windows registry management
- ✅ **Native Integration** - Uses Windows system tray APIs

### **Auto-Startup with Windows**
Automatic startup configuration through Windows registry modification.

#### **Features:**
- ✅ **Registry Integration** - Adds/removes from `HKEY_CURRENT_USER\...\Run`
- ✅ **User-Controlled** - Enable/disable through tray menu
- ✅ **Safe Installation** - No admin privileges required for user startup
- ✅ **Status Detection** - Shows current auto-start state in tray menu

## 🚀 **Usage Guide**

### **Starting Guardian with TUI & Tray**
```bash
# Launch Guardian with TUI and system tray support
./guardian.exe --dev tui

# Production mode with tray support
./guardian.exe tui
```

### **TUI Behavior Changes**
- **Q key**: Minimizes to tray (preserves background service)
- **Ctrl+C**: Minimizes to tray (preserves background service)
- **Service Toggle**: Controls actual background monitoring

### **System Tray Menu Options**
1. **Show Dashboard** - Restore TUI interface
2. **Status** - View current protection status
3. **Start/Stop Monitoring** - Toggle background protection
4. **Auto-start with Windows** - Enable/disable startup
5. **Exit Guardian** - Complete shutdown

### **Auto-Startup Management**
- **Enable**: Right-click tray → "Auto-start with Windows"
- **Disable**: Right-click tray → Uncheck "Auto-start with Windows"
- **Registry Location**: `HKEY_CURRENT_USER\Software\Microsoft\Windows\CurrentVersion\Run`

## 🛡️ **Compatibility with Previous Implementation**

### **Windows Service Integration**
- ✅ **Full Compatibility** - Works with existing Windows provider
- ✅ **Event Log Monitoring** - Maintains RDP monitoring functionality
- ✅ **Firewall Integration** - Preserves Windows Firewall blocking
- ✅ **Configuration Support** - Uses same YAML configuration files

### **PowerShell Script Parity**
The enhanced TUI maintains 100% feature parity with the original PowerShell implementation:

- ✅ **Event ID 4625** - Failed RDP logon detection
- ✅ **IP Extraction** - Same regex patterns for IP parsing
- ✅ **Firewall Rules** - Identical `netsh advfirewall` commands
- ✅ **Block Duration** - Same 20-hour default blocking
- ✅ **Threshold Logic** - 5 failed attempts trigger blocking

### **Background Service Architecture**
```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   TUI Interface │────│ ServiceManager   │────│ Windows Provider│
│   (Minimizable) │    │  (Always Active) │    │ (Event Monitor) │
└─────────────────┘    └──────────────────┘    └─────────────────┘
          │                       │                       │
          │                       │                       │
    ┌─────────────┐         ┌─────────────┐         ┌─────────────┐
    │ System Tray │         │ Background  │         │ Windows     │
    │ Integration │         │ Monitoring  │         │ Firewall    │
    └─────────────┘         └─────────────┘         └─────────────┘
```

## 🔧 **Technical Implementation**

### **Dependencies Added**
- `fyne.io/systray` - Cross-platform system tray support
- `golang.org/x/sys/windows/registry` - Windows registry access

### **New Components**
- `internal/tui/service.go` - ServiceManager with tray integration
- `internal/tui/startup_windows.go` - Windows startup management
- `internal/tui/icon.go` - Embedded Guardian icon

### **Enhanced Files**
- `cmd/guardian/main.go` - Updated TUI command with tray support
- `internal/tui/dashboard.go` - Modified quit behavior for tray minimization

## 📋 **User Experience Flow**

### **First Launch**
1. User runs `guardian.exe tui`
2. TUI opens with system tray icon
3. User can optionally enable auto-startup
4. Background monitoring starts automatically

### **Daily Usage**
1. Guardian starts with Windows (if auto-startup enabled)
2. Runs silently in background with tray icon
3. User can restore TUI by clicking tray icon
4. Closing TUI returns to background mode (doesn't stop service)

### **Service Management**
1. Toggle monitoring via TUI 'S' key or tray menu
2. View status via tray menu or TUI dashboard
3. Complete exit only through tray "Exit Guardian" option

## 🔍 **Advantages Over Traditional Windows Service**

### **User-Friendly Installation**
- ❌ **No Admin Required** - Installs in user context only
- ❌ **No Service Installation** - Avoids Windows service complexity
- ❌ **No System Modification** - Clean user-space operation

### **Enhanced Accessibility**
- ✅ **Interactive Interface** - TUI available on-demand
- ✅ **Visual Feedback** - System tray status indication
- ✅ **Easy Management** - Right-click menu access
- ✅ **User Control** - Start/stop without admin privileges

### **Flexible Deployment**
- ✅ **Portable Operation** - Single executable with all features
- ✅ **User-Specific Config** - No system-wide configuration conflicts
- ✅ **Easy Updates** - Replace single executable file
- ✅ **Clean Uninstall** - Remove auto-start + delete executable

This implementation provides the best of both worlds: the reliability of a background service with the accessibility of a user application! 🛡️✨
