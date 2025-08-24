# Guardian Server Testing Script
# Run as Administrator

param(
    [switch]$CheckPrereqs,
    [switch]$TestFirewall,
    [switch]$CleanupRules,
    [switch]$GenerateTestEvents,
    [switch]$ShowLogs
)

# Colors for output
$Red = "Red"
$Green = "Green" 
$Yellow = "Yellow"
$Cyan = "Cyan"

function Write-Status {
    param($Message, $Color = "White")
    Write-Host "🛡️  $Message" -ForegroundColor $Color
}

function Test-AdminPrivileges {
    $currentUser = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = New-Object Security.Principal.WindowsPrincipal($currentUser)
    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

function Test-Prerequisites {
    Write-Status "Checking Prerequisites..." $Cyan
    
    # Check admin privileges
    if (-not (Test-AdminPrivileges)) {
        Write-Status "❌ Administrator privileges required" $Red
        return $false
    }
    Write-Status "✅ Running as Administrator" $Green
    
    # Check Windows Firewall service
    $firewallSvc = Get-Service -Name MpsSvc -ErrorAction SilentlyContinue
    if (-not $firewallSvc) {
        Write-Status "❌ Windows Firewall service not found" $Red
        return $false
    }
    
    if ($firewallSvc.Status -ne "Running") {
        Write-Status "⚠️  Windows Firewall service not running, starting..." $Yellow
        Start-Service MpsSvc
    }
    Write-Status "✅ Windows Firewall service running" $Green
    
    # Check PowerShell version
    $psVersion = $PSVersionTable.PSVersion.Major
    if ($psVersion -lt 5) {
        Write-Status "⚠️  PowerShell 5.1+ recommended (current: $psVersion)" $Yellow
    } else {
        Write-Status "✅ PowerShell version: $psVersion" $Green
    }
    
    # Check Guardian binary
    if (Test-Path "guardian.exe") {
        Write-Status "✅ Guardian binary found" $Green
    } else {
        Write-Status "❌ Guardian.exe not found in current directory" $Red
        return $false
    }
    
    # Check configuration
    if (Test-Path "guardian.yaml") {
        Write-Status "✅ Configuration file found" $Green
    } else {
        Write-Status "⚠️  guardian.yaml not found, will use defaults" $Yellow
    }
    
    return $true
}

function Test-FirewallIntegration {
    Write-Status "Testing Windows Firewall Integration..." $Cyan
    
    # Test IP for blocking
    $testIP = "203.0.113.100"  # RFC 5737 test IP
    $testRuleName = "Guardian-Test-$(Get-Date -Format 'yyyyMMddHHmmss')"
    
    Write-Status "Creating test firewall rule for IP: $testIP" $Yellow
    
    # Create test rule
    try {
        netsh advfirewall firewall add rule name="$testRuleName" dir=in action=block remoteip=$testIP description="Guardian test rule"
        Write-Status "✅ Test firewall rule created successfully" $Green
    } catch {
        Write-Status "❌ Failed to create firewall rule: $_" $Red
        return $false
    }
    
    # Verify rule exists
    $ruleCheck = netsh advfirewall firewall show rule name="$testRuleName"
    if ($ruleCheck -match $testRuleName) {
        Write-Status "✅ Firewall rule verified" $Green
    } else {
        Write-Status "❌ Firewall rule verification failed" $Red
    }
    
    # Cleanup test rule
    Write-Status "Cleaning up test rule..." $Yellow
    netsh advfirewall firewall delete rule name="$testRuleName"
    Write-Status "✅ Test rule removed" $Green
    
    return $true
}

function Remove-GuardianRules {
    Write-Status "Removing all Guardian firewall rules..." $Yellow
    
    # Get all Guardian rules
    $allRules = netsh advfirewall firewall show rule name=all
    $guardianRules = $allRules | Select-String "Guardian"
    
    if (-not $guardianRules) {
        Write-Status "ℹ️  No Guardian rules found" $Cyan
        return
    }
    
    # Extract rule names and remove them
    $ruleCount = 0
    foreach ($ruleLine in $guardianRules) {
        if ($ruleLine -match "Rule Name:\s*(.+Guardian.+)") {
            $ruleName = $matches[1].Trim()
            try {
                netsh advfirewall firewall delete rule name="$ruleName"
                Write-Status "🗑️  Removed: $ruleName" $Green
                $ruleCount++
            } catch {
                Write-Status "❌ Failed to remove: $ruleName" $Red
            }
        }
    }
    
    Write-Status "✅ Removed $ruleCount Guardian firewall rules" $Green
}

function Show-SecurityLogs {
    Write-Status "Recent Failed RDP Attempts (Event ID 4625):" $Cyan
    
    try {
        $events = Get-WinEvent -FilterHashtable @{LogName='Security'; ID=4625} -MaxEvents 10 -ErrorAction SilentlyContinue
        
        if (-not $events) {
            Write-Status "ℹ️  No recent failed logon events found" $Cyan
            return
        }
        
        foreach ($event in $events) {
            $message = $event.Message
            if ($message -match "Source Network Address:\s*([\d\.]+)") {
                $sourceIP = $matches[1]
                $timestamp = $event.TimeCreated.ToString("yyyy-MM-dd HH:mm:ss")
                Write-Status "🚨 $timestamp - Failed logon from $sourceIP" $Yellow
            }
        }
    } catch {
        Write-Status "❌ Error reading Security Event Log: $_" $Red
        Write-Status "💡 Ensure you're running as Administrator" $Yellow
    }
}

function Generate-TestEvents {
    Write-Status "Generating Test RDP Failures..." $Yellow
    Write-Status "⚠️  This will attempt failed RDP connections to localhost" $Yellow
    Write-Status "Press Ctrl+C to cancel..." $Red
    Start-Sleep -Seconds 3
    
    # This is a simplified test - in reality you'd attempt connections from different IPs
    Write-Status "💡 To generate real test events:" $Cyan
    Write-Status "   1. From another machine, attempt RDP with wrong credentials" $Cyan
    Write-Status "   2. Use different usernames (admin, administrator, test)" $Cyan  
    Write-Status "   3. Repeat 3-5 times to trigger Guardian threshold" $Cyan
    Write-Status "   4. Watch Guardian console for blocking actions" $Cyan
}

# Main execution
Write-Status "Guardian Server Testing Script" $Cyan
Write-Status "==============================" $Cyan

if ($CheckPrereqs) {
    Test-Prerequisites
}

if ($TestFirewall) {
    if (-not (Test-AdminPrivileges)) {
        Write-Status "❌ Administrator privileges required for firewall testing" $Red
        exit 1
    }
    Test-FirewallIntegration
}

if ($CleanupRules) {
    if (-not (Test-AdminPrivileges)) {
        Write-Status "❌ Administrator privileges required for rule cleanup" $Red
        exit 1
    }
    Remove-GuardianRules
}

if ($GenerateTestEvents) {
    Generate-TestEvents
}

if ($ShowLogs) {
    Show-SecurityLogs
}

# If no parameters, show help
if (-not ($CheckPrereqs -or $TestFirewall -or $CleanupRules -or $GenerateTestEvents -or $ShowLogs)) {
    Write-Status "Usage Examples:" $Cyan
    Write-Status "  ./test-server.ps1 -CheckPrereqs      # Check prerequisites" $Yellow
    Write-Status "  ./test-server.ps1 -TestFirewall      # Test firewall integration" $Yellow
    Write-Status "  ./test-server.ps1 -ShowLogs          # Show recent failed attempts" $Yellow
    Write-Status "  ./test-server.ps1 -CleanupRules      # Remove Guardian rules" $Yellow
    Write-Status "" 
    Write-Status "💡 Start with: ./test-server.ps1 -CheckPrereqs" $Green
}
