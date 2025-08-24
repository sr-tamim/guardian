# Guardian Server Testing Package

## 🚀 Windows Server Testing Guide

### **Package Contents**
- `guardian.exe` - Main Guardian binary (8.4MB)
- `guardian.yaml` - Production configuration
- `test-server.ps1` - PowerShell testing script
- `install-service.ps1` - Service installation script

### **Prerequisites**
- Windows Server 2016+ or Windows 10+
- Administrator privileges
- PowerShell 5.1 or higher
- Windows Firewall service enabled

---

## 🧪 Testing Steps

### **1. Initial Setup**
```powershell
# Run PowerShell as Administrator
# Navigate to Guardian directory
cd C:\Guardian

# Verify prerequisites
./test-server.ps1 -CheckPrereqs
```

### **2. Configuration Testing**
```powershell
# Test configuration loading
./guardian.exe status

# Test platform detection
./guardian.exe monitor --config guardian.yaml --dry-run
```

### **3. Real-World Testing**

#### **Option A: Safe Testing (Recommended First)**
```powershell
# Test with mock data generation
./guardian.exe monitor --dev
# This will simulate attacks safely without real blocking
```

#### **Option B: Real Event Log Monitoring**
```powershell
# Monitor actual Windows Security Event Log
# Requires Administrator privileges
./guardian.exe monitor

# In another terminal, generate test failures:
# Try RDP with wrong credentials to your server
# Watch Guardian detect and block the attempts
```

### **4. Firewall Integration Testing**
```powershell
# Test firewall rule creation (requires admin)
./test-server.ps1 -TestFirewall

# Manual verification
netsh advfirewall firewall show rule name=all | findstr "Guardian"
```

### **5. Production Simulation**
```powershell
# Run with production settings
./guardian.exe monitor --config guardian.yaml

# Monitor logs in real-time
Get-EventLog -LogName Security -InstanceId 4625 -Newest 10
```

---

## 🔧 Configuration Options

### **Production Settings** (`guardian.yaml`)
- **Failure Threshold**: 5 attempts (vs 3 in dev)
- **Block Duration**: 20 hours (vs 2 minutes in dev)
- **Event Source**: Windows Security Event Log
- **Service**: RDP monitoring enabled

### **For Your Production Environment**
1. **Adjust IP Whitelist**: Add your trusted networks
2. **Set Block Duration**: Match your security policy
3. **Configure Thresholds**: Based on your attack patterns
4. **Enable Logging**: Set persistent log files

---

## 🛡️ Real Attack Detection

### **What Guardian Will Detect**
- Failed RDP logon attempts (Event ID 4625)
- Multiple failures from same IP
- Administrator account targeting
- Dictionary attacks on common usernames

### **What Guardian Will Do**
1. **Parse** Windows Security Event Log in real-time
2. **Count** failed attempts per IP address
3. **Block** IPs exceeding threshold using Windows Firewall
4. **Log** all activity with timestamps
5. **Auto-cleanup** expired blocks

---

## 📊 Monitoring Commands

```powershell
# Check Guardian status
./guardian.exe status

# List currently blocked IPs
netsh advfirewall firewall show rule name=all | findstr "Guardian"

# View recent security events
Get-WinEvent -FilterHashtable @{LogName='Security'; ID=4625} -MaxEvents 20

# Monitor Guardian in real-time (separate terminal)
./guardian.exe monitor
```

---

## 🚨 Emergency Commands

```powershell
# Stop Guardian immediately
Ctrl+C  # In Guardian terminal

# Remove all Guardian firewall rules
./test-server.ps1 -CleanupRules

# Emergency unblock specific IP
netsh advfirewall firewall delete rule name="Guardian - [timestamp] - [IP]"
```

---

## 🔍 Troubleshooting

### **Common Issues**

1. **"Access Denied" Errors**
   - Solution: Run PowerShell as Administrator
   - Verify: `([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")`

2. **"Windows Firewall Service Not Available"**
   - Solution: `Start-Service MpsSvc`
   - Check: `Get-Service MpsSvc`

3. **"Event Log Access Denied"**
   - Solution: Requires Administrator privileges for Security Event Log
   - Alternative: Test with mock provider first

4. **No Events Detected**
   - Check: Are RDP connections actually failing?
   - Verify: `Get-WinEvent -FilterHashtable @{LogName='Security'; ID=4625} -MaxEvents 5`

### **Log Files**
- Guardian logs: Check console output
- Windows Event Log: Event Viewer → Windows Logs → Security
- Firewall rules: `netsh advfirewall firewall show rule name=all`

---

## 🎯 Testing Success Criteria

✅ **Platform Detection**: Shows "Windows Provider" (not Mock Provider)
✅ **Event Parsing**: Successfully extracts IP addresses from Event ID 4625
✅ **Firewall Rules**: Creates rules with proper naming convention
✅ **Blocking Logic**: Blocks IPs after threshold reached
✅ **Cleanup**: Automatically removes expired blocks
✅ **Configuration**: Loads production settings correctly

---

**🔐 Security Note**: Test in a controlled environment first. Guardian will create real firewall rules that block network access.
