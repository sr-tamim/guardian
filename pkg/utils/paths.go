package utils

import (
	"os"
	"path/filepath"
	"runtime"
)

// PlatformPaths provides platform-specific default paths
type PlatformPaths struct{}

// NewPlatformPaths creates a new platform paths utility
func NewPlatformPaths() *PlatformPaths {
	return &PlatformPaths{}
}

// GetDefaultLogDir returns the default log directory for the platform
func (p *PlatformPaths) GetDefaultLogDir() string {
	switch runtime.GOOS {
	case "windows":
		// Windows: Use ProgramData or temp
		if programData := os.Getenv("PROGRAMDATA"); programData != "" {
			return filepath.Join(programData, "Guardian", "logs")
		}
		return filepath.Join(os.TempDir(), "guardian", "logs")

	case "linux":
		// Linux: Standard system log directory
		if os.Geteuid() == 0 { // Running as root
			return "/var/log/guardian"
		}
		// Non-root user
		if home := os.Getenv("HOME"); home != "" {
			return filepath.Join(home, ".local", "share", "guardian", "logs")
		}
		return filepath.Join("/tmp", "guardian", "logs")

	case "darwin":
		// macOS: Standard log directory
		if os.Geteuid() == 0 { // Running as root
			return "/var/log/guardian"
		}
		// Non-root user
		if home := os.Getenv("HOME"); home != "" {
			return filepath.Join(home, "Library", "Logs", "Guardian")
		}
		return filepath.Join("/tmp", "guardian", "logs")

	default:
		// Fallback to temp directory
		return filepath.Join(os.TempDir(), "guardian", "logs")
	}
}

// GetDefaultDataDir returns the default data directory for the platform
func (p *PlatformPaths) GetDefaultDataDir() string {
	switch runtime.GOOS {
	case "windows":
		// Windows: Use ProgramData or AppData
		if programData := os.Getenv("PROGRAMDATA"); programData != "" {
			return filepath.Join(programData, "Guardian", "data")
		}
		if appData := os.Getenv("APPDATA"); appData != "" {
			return filepath.Join(appData, "Guardian")
		}
		return filepath.Join(os.TempDir(), "guardian", "data")

	case "linux":
		// Linux: Standard data directory
		if os.Geteuid() == 0 { // Running as root
			return "/var/lib/guardian"
		}
		// Non-root user
		if home := os.Getenv("HOME"); home != "" {
			return filepath.Join(home, ".local", "share", "guardian")
		}
		return filepath.Join("/tmp", "guardian", "data")

	case "darwin":
		// macOS: Standard data directory
		if os.Geteuid() == 0 { // Running as root
			return "/var/lib/guardian"
		}
		// Non-root user
		if home := os.Getenv("HOME"); home != "" {
			return filepath.Join(home, "Library", "Application Support", "Guardian")
		}
		return filepath.Join("/tmp", "guardian", "data")

	default:
		// Fallback to temp directory
		return filepath.Join(os.TempDir(), "guardian", "data")
	}
}

// GetDefaultServiceLogPaths returns platform-specific log paths for services
func (p *PlatformPaths) GetDefaultServiceLogPaths(service string) []string {
	switch runtime.GOOS {
	case "windows":
		return p.getWindowsServiceLogPaths(service)
	case "linux":
		return p.getLinuxServiceLogPaths(service)
	case "darwin":
		return p.getDarwinServiceLogPaths(service)
	default:
		return p.getFallbackServiceLogPaths(service)
	}
}

// getWindowsServiceLogPaths returns Windows-specific service log paths
func (p *PlatformPaths) getWindowsServiceLogPaths(service string) []string {
	switch service {
	case "RDP", "rdp":
		return []string{"Security"} // Windows Event Log name
	case "SSH", "ssh":
		// Common Windows SSH server log locations
		return []string{
			`C:\ProgramData\ssh\logs\sshd.log`,
			`C:\Windows\System32\OpenSSH\logs\sshd.log`,
			filepath.Join(p.GetDefaultLogDir(), "ssh_test.log"), // Fallback for testing
		}
	case "IIS", "iis":
		return []string{`C:\inetpub\logs\LogFiles\W3SVC1\*.log`}
	case "Apache", "apache":
		return []string{
			`C:\Apache24\logs\access.log`,
			`C:\xampp\apache\logs\access.log`,
		}
	default:
		return []string{filepath.Join(p.GetDefaultLogDir(), service+"_test.log")}
	}
}

// getLinuxServiceLogPaths returns Linux-specific service log paths
func (p *PlatformPaths) getLinuxServiceLogPaths(service string) []string {
	switch service {
	case "SSH", "ssh":
		return []string{
			"/var/log/auth.log", // Debian/Ubuntu
			"/var/log/secure",   // RHEL/CentOS
			"/var/log/messages", // Generic
		}
	case "Apache", "apache":
		return []string{
			"/var/log/apache2/access.log", // Debian/Ubuntu
			"/var/log/httpd/access_log",   // RHEL/CentOS
		}
	case "Nginx", "nginx":
		return []string{
			"/var/log/nginx/access.log",
			"/var/log/nginx/error.log",
		}
	default:
		return []string{filepath.Join("/tmp", service+"_test.log")}
	}
}

// getDarwinServiceLogPaths returns macOS-specific service log paths
func (p *PlatformPaths) getDarwinServiceLogPaths(service string) []string {
	switch service {
	case "SSH", "ssh":
		return []string{
			"/var/log/auth.log",
			"/var/log/system.log",
		}
	case "Apache", "apache":
		return []string{
			"/var/log/apache2/access_log",
			"/usr/local/var/log/apache2/access_log", // Homebrew
		}
	case "Nginx", "nginx":
		return []string{
			"/var/log/nginx/access.log",
			"/usr/local/var/log/nginx/access.log", // Homebrew
		}
	default:
		return []string{filepath.Join("/tmp", service+"_test.log")}
	}
}

// getFallbackServiceLogPaths returns fallback service log paths
func (p *PlatformPaths) getFallbackServiceLogPaths(service string) []string {
	return []string{filepath.Join(os.TempDir(), "guardian", service+"_test.log")}
}

// GetDefaultGuardianLogPath returns the default Guardian application log path
func (p *PlatformPaths) GetDefaultGuardianLogPath() string {
	return filepath.Join(p.GetDefaultLogDir(), "guardian.log")
}

// GetDefaultGuardianDatabasePath returns the default Guardian database path
func (p *PlatformPaths) GetDefaultGuardianDatabasePath() string {
	return filepath.Join(p.GetDefaultDataDir(), "guardian.db")
}

// EnsureDir creates a directory if it doesn't exist
func (p *PlatformPaths) EnsureDir(path string) error {
	return os.MkdirAll(filepath.Dir(path), 0755)
}
