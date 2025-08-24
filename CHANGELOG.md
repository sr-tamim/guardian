# Guardian Version History

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
