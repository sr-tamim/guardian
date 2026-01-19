package daemon

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/sr-tamim/guardian/internal/core"
	"github.com/sr-tamim/guardian/pkg/logger"
	"github.com/sr-tamim/guardian/pkg/models"
)

// Manager handles daemon mode operations
type Manager struct {
	pidManager *PIDManager
	config     *models.Config
	provider   core.PlatformProvider
	devMode    bool
	configPath string
}

// NewManager creates a new daemon manager
func NewManager(config *models.Config, provider core.PlatformProvider, devMode bool, configPath string) *Manager {
	return &Manager{
		pidManager: NewPIDManager(),
		config:     config,
		provider:   provider,
		devMode:    devMode,
		configPath: configPath,
	}
}

// StartDaemon starts Guardian in daemon mode
func (dm *Manager) StartDaemon() error {
	return dm.StartDaemonWithOptions(false)
}

// StartDaemonWithTray starts Guardian in daemon mode with system tray support
func (dm *Manager) StartDaemonWithTray() error {
	return dm.StartDaemonWithOptions(true)
}

// StartDaemonWithOptions starts Guardian daemon with configurable tray support
func (dm *Manager) StartDaemonWithOptions(withTray bool) error {
	// Check if already running
	if pid, running := dm.pidManager.GetRunningPID(); running {
		return fmt.Errorf("Guardian daemon is already running (PID: %d)", pid)
	}

	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Prepare daemon command args (use internal flag to avoid recursion)
	args := []string{"monitor", "--daemon-internal"}
	if dm.devMode {
		args = append(args, "--dev")
	}
	if dm.configPath != "" {
		args = append(args, "--config", dm.configPath)
	}
	if withTray {
		args = append(args, "--tray")
	}

	// Create log directory
	logDir := dm.getLogDir()
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Prepare log files
	stdoutPath := filepath.Join(logDir, "guardian-daemon.log")
	stderrPath := filepath.Join(logDir, "guardian-daemon.err")

	// Start the daemon process
	cmd := exec.Command(execPath, args...)
	cmd.Env = os.Environ()

	// On Windows, properly detach the process
	if runtime.GOOS == "windows" {
		// For system tray support, we need to allow window creation
		// For headless daemon, we can hide the window
		creationFlags := uint32(syscall.CREATE_NEW_PROCESS_GROUP)
		if !withTray {
			creationFlags |= 0x08000000 // CREATE_NO_WINDOW - only if no tray support
		}
		cmd.SysProcAttr = &syscall.SysProcAttr{
			CreationFlags: creationFlags,
		}
	}

	// Redirect stdout/stderr to log files
	stdout, err := os.OpenFile(stdoutPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open stdout log: %w", err)
	}

	stderr, err := os.OpenFile(stderrPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		stdout.Close() // Clean up stdout if stderr fails
		return fmt.Errorf("failed to open stderr log: %w", err)
	}

	cmd.Stdout = stdout
	cmd.Stderr = stderr

	// Start the process
	if err := cmd.Start(); err != nil {
		stdout.Close()
		stderr.Close()
		return fmt.Errorf("failed to start daemon: %w", err)
	}

	// Get PID before releasing
	pid := cmd.Process.Pid

	// Release the process so it can run independently
	if err := cmd.Process.Release(); err != nil {
		// Log the error but don't fail, as the process might still work
		fmt.Printf("Warning: failed to release process: %v\n", err)
	}

	// Close file handles in parent process (daemon will keep them open)
	stdout.Close()
	stderr.Close()

	// Wait a moment to ensure process starts
	time.Sleep(time.Millisecond * 500)

	if !dm.pidManager.isProcessRunning(pid) {
		return fmt.Errorf("daemon process failed to start or exited immediately")
	}

	fmt.Printf("🛡️  Guardian daemon started successfully (PID: %d)\n", pid)
	fmt.Printf("📝 Logs: %s\n", stdoutPath)
	fmt.Printf("❌ Errors: %s\n", stderrPath)

	return nil
}

// StopDaemon stops the running Guardian daemon
func (dm *Manager) StopDaemon() error {
	pid, running := dm.pidManager.GetRunningPID()
	if !running {
		// Clean up any stale PID file
		dm.pidManager.RemovePID()
		return fmt.Errorf("no Guardian daemon is currently running")
	}

	// Find and terminate the process
	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process %d: %w", pid, err)
	}

	fmt.Printf("🛑 Stopping Guardian daemon (PID: %d)...\n", pid)

	// On Windows, use taskkill for reliable termination
	if runtime.GOOS == "windows" {
		// Use taskkill to terminate the process
		cmd := exec.Command("taskkill", "/PID", fmt.Sprintf("%d", pid), "/F")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to terminate process with taskkill: %w", err)
		}
	} else {
		// Send termination signal (Unix-like systems)
		if err := proc.Signal(os.Interrupt); err != nil {
			return fmt.Errorf("failed to send termination signal: %w", err)
		}

		// Wait for graceful shutdown
		timeout := time.After(10 * time.Second)
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-timeout:
				// Force kill if graceful shutdown failed
				fmt.Printf("⚠️  Graceful shutdown timeout, force killing process...\n")
				if err := proc.Kill(); err != nil {
					return fmt.Errorf("failed to force kill process: %w", err)
				}
				break
			case <-ticker.C:
				if !dm.pidManager.isProcessRunning(pid) {
					// Process stopped successfully
					break
				}
				continue
			}
			break
		}
	}

	// Clean up PID file
	dm.pidManager.RemovePID()
	fmt.Printf("✅ Guardian daemon stopped successfully\n")
	return nil
}

// GetStatus returns the current daemon status
func (dm *Manager) GetStatus() *DaemonStatus {
	pidStatus := dm.pidManager.GetStatus()

	status := &DaemonStatus{
		Running:   pidStatus.Running,
		PID:       pidStatus.PID,
		PIDFile:   pidStatus.PIDFile,
		Startable: pidStatus.Startable,
		Error:     pidStatus.Error,
		LogDir:    dm.getLogDir(),
	}

	// Add additional daemon-specific info
	if status.Running {
		status.LogFiles = []string{
			filepath.Join(status.LogDir, "guardian-daemon.log"),
			filepath.Join(status.LogDir, "guardian-daemon.err"),
		}
	}

	return status
}

// IsRunning checks if daemon is currently running
func (dm *Manager) IsRunning() bool {
	return dm.pidManager.IsRunning()
}

// RunMonitorInCurrentProcess runs the monitoring in the current process
// This is used when monitor is called in daemon mode
func (dm *Manager) RunMonitorInCurrentProcess(ctx context.Context) error {
	return dm.RunMonitorWithOptions(ctx, false)
}

// RunMonitorWithTray runs the monitoring with system tray support
func (dm *Manager) RunMonitorWithTray(ctx context.Context) error {
	return dm.RunMonitorWithOptions(ctx, true)
}

// RunMonitorWithOptions runs monitoring with configurable system tray support
func (dm *Manager) RunMonitorWithOptions(ctx context.Context, withTray bool) error {
	// Write PID file for current process
	if err := dm.pidManager.WritePID(); err != nil {
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	// Ensure PID file is cleaned up on exit
	defer dm.pidManager.RemovePID()

	fmt.Printf("🛡️  Guardian daemon monitoring started (PID: %d)\n", os.Getpid())
	fmt.Printf("📄 PID file: %s\n", dm.pidManager.GetPIDFilePath())

	// Create context for monitoring that can be cancelled
	monitorCtx, monitorCancel := context.WithCancel(ctx)
	defer monitorCancel()

	// Heartbeat logging for service reliability monitoring
	startTime := time.Now()
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-monitorCtx.Done():
				return
			case <-ticker.C:
				logger.Info("Guardian heartbeat",
					"uptime", time.Since(startTime).Truncate(time.Second),
					"platform", dm.provider.Name(),
					"services", dm.enabledServiceCount(),
				)
			}
		}
	}()

	// Start monitoring for enabled services
	for _, service := range dm.config.Services {
		if service.Enabled {
			logPaths, err := dm.provider.GetLogPaths(service.Name)
			if err != nil {
				fmt.Printf("❌ Failed to get log paths for %s: %v\n", service.Name, err)
				continue
			}

			for _, logPath := range logPaths {
				go func(path, serviceName string) {
					// Start log monitoring
					if err := dm.provider.StartLogMonitoring(monitorCtx, path, nil); err != nil {
						fmt.Printf("❌ Failed to start monitoring %s: %v\n", path, err)
					}
				}(logPath, service.Name)
			}
		}
	}

	// If tray support is enabled, start the system tray
	if withTray {
		fmt.Println("🖼️  Starting system tray interface...")

		// Create shutdown callback
		shutdownCallback := func() {
			fmt.Println("🛑 Shutdown requested from system tray")
			monitorCancel() // Cancel monitoring context
		}

		// Create and start tray manager
		trayManager := NewTrayManager(dm.provider, dm.devMode, shutdownCallback)

		// Start tray in goroutine so it doesn't block monitoring
		go func() {
			trayManager.StartTray()
		}()
	}

	// Wait for context cancellation (graceful shutdown)
	<-monitorCtx.Done()
	fmt.Println("🛑 Guardian daemon monitoring stopped")

	return nil
}

func (dm *Manager) enabledServiceCount() int {
	count := 0
	for _, service := range dm.config.Services {
		if service.Enabled {
			count++
		}
	}
	return count
}

// getLogDir returns the appropriate log directory for the platform
func (dm *Manager) getLogDir() string {
	switch runtime.GOOS {
	case "windows":
		// Use AppData/Local for Windows
		appData := os.Getenv("LOCALAPPDATA")
		if appData == "" {
			appData = os.Getenv("USERPROFILE") + "\\AppData\\Local"
		}
		return filepath.Join(appData, "Guardian", "logs")
	default:
		// Use /var/log for Unix-like systems, fallback to ~/.local/share
		if os.Getuid() == 0 {
			return "/var/log/guardian"
		}
		home := os.Getenv("HOME")
		return filepath.Join(home, ".local", "share", "guardian", "logs")
	}
}

// DaemonStatus represents detailed daemon status information
type DaemonStatus struct {
	Running   bool
	PID       int
	PIDFile   string
	Startable bool
	Error     error
	LogDir    string
	LogFiles  []string
}
