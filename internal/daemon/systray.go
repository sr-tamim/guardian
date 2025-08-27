package daemon

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"syscall"

	"fyne.io/systray"

	"github.com/sr-tamim/guardian/internal/core"
	"github.com/sr-tamim/guardian/pkg/version"
)

// TrayManager handles system tray functionality for the daemon
type TrayManager struct {
	provider   core.PlatformProvider
	devMode    bool
	ctx        context.Context
	cancel     context.CancelFunc
	onShutdown func() // Callback for daemon shutdown
}

// NewTrayManager creates a new system tray manager for the daemon
func NewTrayManager(provider core.PlatformProvider, devMode bool, onShutdown func()) *TrayManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &TrayManager{
		provider:   provider,
		devMode:    devMode,
		ctx:        ctx,
		cancel:     cancel,
		onShutdown: onShutdown,
	}
}

// StartTray initializes and runs the system tray
func (tm *TrayManager) StartTray() {
	systray.Run(tm.onTrayReady, tm.onTrayExit)
}

// onTrayReady sets up the system tray menu and functionality
func (tm *TrayManager) onTrayReady() {
	// Set tray title and icon
	tm.initializeTrayDisplay()

	// Create menu items
	mStatus := systray.AddMenuItem("Show Status", "View Guardian daemon status")
	mShowDashboard := systray.AddMenuItem("Open Dashboard", "Launch Guardian TUI Dashboard")
	systray.AddSeparator()

	mLogs := systray.AddMenuItem("View Logs", "Open daemon log file")
	systray.AddSeparator()

	mStop := systray.AddMenuItem("Stop Daemon", "Stop Guardian daemon")
	systray.AddSeparator()
	mExit := systray.AddMenuItem("Exit", "Stop daemon and exit")

	// Handle menu actions in separate goroutines
	go tm.handleMenuActions(mStatus, mShowDashboard, mLogs, mStop, mExit)
}

// handleMenuActions processes system tray menu clicks
func (tm *TrayManager) handleMenuActions(mStatus, mShowDashboard, mLogs, mStop, mExit *systray.MenuItem) {
	for {
		select {
		case <-mStatus.ClickedCh:
			tm.showStatus()

		case <-mShowDashboard.ClickedCh:
			tm.launchDashboard()

		case <-mLogs.ClickedCh:
			tm.openLogs()

		case <-mStop.ClickedCh:
			tm.stopDaemon()

		case <-mExit.ClickedCh:
			tm.exitDaemon()
			return

		case <-tm.ctx.Done():
			return
		}
	}
}

// showStatus displays daemon status in a notification or dialog
func (tm *TrayManager) showStatus() {
	pidManager := NewPIDManager()

	var message string
	if pid, running := pidManager.GetRunningPID(); running {
		message = fmt.Sprintf("Guardian Daemon Status:\n✅ Running (PID: %d)\n🛡️ Monitoring active", pid)
	} else {
		message = "Guardian Daemon Status:\n❌ Not running"
	}

	// For now, we'll use a simple approach - in a full implementation,
	// you'd show a native notification or dialog
	fmt.Printf("📊 %s\n", message)
}

// launchDashboard opens the Guardian TUI dashboard
func (tm *TrayManager) launchDashboard() {
	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		fmt.Printf("❌ Failed to get executable path: %v\n", err)
		return
	}

	// Build TUI command
	args := []string{"tui"}
	if tm.devMode {
		args = append(args, "--dev")
	}

	// Launch TUI in new process
	cmd := exec.Command(execPath, args...)
	cmd.Env = os.Environ()

	// On Windows, show the TUI window
	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			CreationFlags: 0x00000010, // CREATE_NEW_CONSOLE
		}
	}

	if err := cmd.Start(); err != nil {
		fmt.Printf("❌ Failed to launch dashboard: %v\n", err)
		return
	}

	fmt.Println("🖥️ Guardian dashboard launched")
}

// openLogs opens the daemon log file with system default application
func (tm *TrayManager) openLogs() {
	logPath := tm.getLogPath()

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", logPath)
	case "darwin":
		cmd = exec.Command("open", logPath)
	default: // linux
		cmd = exec.Command("xdg-open", logPath)
	}

	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Failed to open logs: %v\n", err)
	} else {
		fmt.Println("📝 Log file opened")
	}
}

// stopDaemon stops the daemon process
func (tm *TrayManager) stopDaemon() {
	if tm.onShutdown != nil {
		fmt.Println("🛑 Stopping Guardian daemon...")
		tm.onShutdown()
	}
}

// exitDaemon stops the daemon and exits the tray
func (tm *TrayManager) exitDaemon() {
	if tm.onShutdown != nil {
		tm.onShutdown()
	}
	systray.Quit()
}

// onTrayExit is called when the system tray is exited
func (tm *TrayManager) onTrayExit() {
	tm.cancel()
}

// initializeTrayDisplay sets up platform-specific tray display
func (tm *TrayManager) initializeTrayDisplay() {
	versionInfo := version.Get()
	title := fmt.Sprintf("Guardian v%s", versionInfo.Version)

	if tm.devMode {
		title += " (Dev)"
	}

	systray.SetTemplateIcon(tm.getIconData(), tm.getIconData())
	systray.SetTitle(title)
	systray.SetTooltip("Guardian Intrusion Prevention System")
}

// getIconData returns the system tray icon data
func (tm *TrayManager) getIconData() []byte {
	// Simple Guardian shield icon as bytes (placeholder)
	// In a full implementation, you'd embed actual icon files
	return []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D,
		0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x10,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x91, 0x68, 0x36, 0x00, 0x00, 0x00,
		// Minimal PNG data for a simple icon
	}
}

// getLogPath returns the path to the daemon log file
func (tm *TrayManager) getLogPath() string {
	switch runtime.GOOS {
	case "windows":
		if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			return fmt.Sprintf("%s\\Guardian\\logs\\guardian-daemon.log", localAppData)
		}
		return "C:\\Windows\\Temp\\guardian-daemon.log"
	case "darwin":
		if home := os.Getenv("HOME"); home != "" {
			return fmt.Sprintf("%s/Library/Logs/Guardian/guardian-daemon.log", home)
		}
		return "/tmp/guardian-daemon.log"
	default: // linux
		return "/var/log/guardian/guardian-daemon.log"
	}
}

// Stop gracefully stops the tray manager
func (tm *TrayManager) Stop() {
	tm.cancel()
}
