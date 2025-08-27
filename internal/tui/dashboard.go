package tui

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/sr-tamim/guardian/internal/autostart"
	"github.com/sr-tamim/guardian/internal/core"
	"github.com/sr-tamim/guardian/internal/daemon"
	"github.com/sr-tamim/guardian/pkg/version"
)

// Dashboard represents the main TUI interface
type Dashboard struct {
	provider   core.PlatformProvider
	devMode    bool
	pidManager *daemon.PIDManager
	width      int
	height     int
	ready      bool
	quitting   bool

	// State
	daemonRunning    bool
	daemonPID        int
	autostartEnabled bool
	blockedIPs       []string
	attackCount      int64
	lastUpdate       time.Time
	recentLogs       []string

	// Navigation
	selectedTab int
	tabs        []string
}

// Tab constants
const (
	TabDashboard = iota
	TabBlocked
	TabLogs
	TabService
	TabSettings
)

// NewDashboard creates a new dashboard instance
func NewDashboard() *Dashboard {
	return &Dashboard{
		tabs:       []string{"Dashboard", "Blocked IPs", "Logs", "Service", "Settings"},
		lastUpdate: time.Now(),
		pidManager: daemon.NewPIDManager(),
		recentLogs: make([]string, 0),
	}
}

// SetProvider sets the platform provider (called after creation)
func (d *Dashboard) SetProvider(provider core.PlatformProvider, devMode bool) {
	d.provider = provider
	d.devMode = devMode
}

// Init initializes the dashboard
func (d *Dashboard) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		d.tickCmd(),
	)
}

// tickCmd returns a command that sends a tick message every second
func (d *Dashboard) tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg{Time: t}
	})
}

// TickMsg represents a periodic update
type TickMsg struct {
	Time time.Time
}

// Update handles messages and updates the model
func (d *Dashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.height = msg.Height
		d.width = msg.Width
		d.ready = true
		return d, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			// Simply quit the dashboard
			d.quitting = true
			return d, tea.Quit

		case "tab", "right":
			d.selectedTab = (d.selectedTab + 1) % len(d.tabs)
			return d, nil

		case "shift+tab", "left":
			d.selectedTab = (d.selectedTab - 1 + len(d.tabs)) % len(d.tabs)
			return d, nil

		case "s":
			// Show daemon control options in service tab
			if d.selectedTab == TabService {
				// For now, just refresh daemon status
				// In the future, could show start/stop daemon options
				if pid, running := d.pidManager.GetRunningPID(); running {
					d.daemonRunning = true
					d.daemonPID = pid
				} else {
					d.daemonRunning = false
					d.daemonPID = 0
				}
				d.updateAutostartStatus()
				d.lastUpdate = time.Now()
				return d, nil
			}

		case "r":
			// Refresh data
			if d.provider != nil {
				// Get real blocked IPs count (mock implementation for now)
				// In real implementation: d.blockedIPs = d.provider.GetBlockedIPs()
				d.attackCount++ // Simulate activity
			}
			d.lastUpdate = time.Now()
			return d, nil
		}

	case TickMsg:
		d.lastUpdate = msg.Time
		// Update daemon status
		if pid, running := d.pidManager.GetRunningPID(); running {
			d.daemonRunning = true
			d.daemonPID = pid
		} else {
			d.daemonRunning = false
			d.daemonPID = 0
		}
		// Update autostart status
		d.updateAutostartStatus()
		// Update recent logs
		d.updateRecentLogs()
		return d, d.tickCmd()
	}

	return d, nil
}

// View renders the dashboard
func (d *Dashboard) View() string {
	if !d.ready {
		return "Initializing Guardian TUI..."
	}

	if d.quitting {
		return "👋 Guardian Dashboard closed\n� Daemon status monitoring ended\n\n💡 Use 'guardian status' for quick status checks"
	}

	// Header
	header := d.renderHeader()

	// Tab navigation
	tabs := d.renderTabs()

	// Content based on selected tab
	var content string
	switch d.selectedTab {
	case TabDashboard:
		content = d.renderDashboardTab()
	case TabBlocked:
		content = d.renderBlockedTab()
	case TabLogs:
		content = d.renderLogsTab()
	case TabService:
		content = d.renderServiceTab()
	case TabSettings:
		content = d.renderSettingsTab()
	}

	// Footer
	footer := d.renderFooter()

	// Combine all sections
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		tabs,
		content,
		footer,
	)
}

// renderHeader creates the main header with Guardian branding
func (d *Dashboard) renderHeader() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00FF88")).
		Background(lipgloss.Color("#1a1a1a")).
		Padding(0, 2).
		Width(d.width)

	versionInfo := version.Get()
	title := fmt.Sprintf("🛡️  Guardian v%s - Daemon Status Viewer", versionInfo.Version)

	return titleStyle.Render(title)
}

// renderTabs creates the tab navigation bar
func (d *Dashboard) renderTabs() string {
	var tabs []string

	for i, tab := range d.tabs {
		style := lipgloss.NewStyle().Padding(0, 2)

		if i == d.selectedTab {
			// Active tab
			style = style.
				Bold(true).
				Foreground(lipgloss.Color("#ffffff")).
				Background(lipgloss.Color("#6366f1"))
		} else {
			// Inactive tab
			style = style.
				Foreground(lipgloss.Color("#6b7280")).
				Background(lipgloss.Color("#374151"))
		}

		tabs = append(tabs, style.Render(tab))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
}

// renderDashboardTab shows the main dashboard overview
func (d *Dashboard) renderDashboardTab() string {
	contentStyle := lipgloss.NewStyle().
		Padding(2).
		Height(d.height - 6)

	statusIcon := "🔴"
	statusText := "STOPPED"
	if d.daemonRunning {
		statusIcon = "🟢"
		statusText = "RUNNING"
	}

	// Provider mode indicator
	modeText := "Production Mode"
	if d.devMode {
		modeText = "Development Mode"
	}

	// Platform info
	platformInfo := "Unknown Platform"
	if d.provider != nil {
		platformInfo = "Platform Provider Active"
	}

	content := fmt.Sprintf(`
%s Service Status: %s (%s)

📊 Statistics:
   • Total Attacks Blocked: %d
   • Currently Blocked IPs: %d
   • Last Update: %s

🛡️  Protection Status:
   • %s
   • Platform Monitoring: %s
   • Platform Firewall: %s
   • Real-time Detection: %s
   • Auto-startup: %s

	content += "
Press 'r' to refresh, 'tab' to navigate"
	`,
		statusIcon, statusText, modeText,
		d.attackCount,
		len(d.blockedIPs),
		d.lastUpdate.Format("15:04:05"),
		platformInfo,
		d.getServiceIcon(d.daemonRunning),
		d.getServiceIcon(d.daemonRunning),
		d.getServiceIcon(d.daemonRunning),
		d.getAutostartIcon(),
	)

	return contentStyle.Render(strings.TrimSpace(content))
}

// renderBlockedTab shows blocked IPs
func (d *Dashboard) renderBlockedTab() string {
	contentStyle := lipgloss.NewStyle().
		Padding(2).
		Height(d.height - 6)

	content := "🚫 Currently Blocked IPs:\n\n"

	if len(d.blockedIPs) == 0 {
		content += "   No IPs currently blocked\n\n"
	} else {
		for i, ip := range d.blockedIPs {
			if i >= 10 { // Limit display
				content += fmt.Sprintf("   ... and %d more\n", len(d.blockedIPs)-10)
				break
			}
			content += fmt.Sprintf("   • %s\n", ip)
		}
	}

	content += "\nPress 'r' to refresh, 'tab' to navigate"

	return contentStyle.Render(content)
}

// renderLogsTab shows recent log activity from daemon logs
func (d *Dashboard) renderLogsTab() string {
	contentStyle := lipgloss.NewStyle().
		Padding(2).
		Height(d.height - 6)

	content := "📝 Recent Daemon Activity:\n\n"

	if len(d.recentLogs) == 0 {
		if d.daemonRunning {
			content += "   Loading daemon logs...\n"
		} else {
			content += "   No daemon running. Start with: guardian monitor -d\n"
		}
	} else {
		for _, logLine := range d.recentLogs {
			if logLine != "" {
				content += fmt.Sprintf("   %s\n", logLine)
			}
		}
	}

	content += "\nPress 'r' to refresh, 'tab' to navigate"

	return contentStyle.Render(content)
}

// renderServiceTab shows service management
func (d *Dashboard) renderServiceTab() string {
	contentStyle := lipgloss.NewStyle().
		Padding(2).
		Height(d.height - 6)

	statusIcon := "🔴"
	statusText := "STOPPED"
	actionText := "Daemon not running - use 'guardian monitor -d' to start"

	if d.daemonRunning {
		statusIcon = "🟢"
		statusText = fmt.Sprintf("RUNNING (PID: %d)", d.daemonPID)
		actionText = "Use 'guardian stop' to stop daemon"
	}

	autostartIcon := "🔴"
	autostartText := "DISABLED"
	if d.autostartEnabled {
		autostartIcon = "🟢"
		autostartText = "ENABLED"
	}

	content := fmt.Sprintf(`
🔧 Service Management

Current Status: %s %s

System Service: Not installed
Background Mode: %s
Auto-startup: %s %s

%s
Press 'r' to refresh, 'tab' to navigate
	`,
		statusIcon, statusText,
		d.getServiceIcon(d.daemonRunning),
		autostartIcon, autostartText,
		actionText,
	)

	return contentStyle.Render(strings.TrimSpace(content))
}

// renderSettingsTab shows configuration options
func (d *Dashboard) renderSettingsTab() string {
	contentStyle := lipgloss.NewStyle().
		Padding(2).
		Height(d.height - 6)

	content := `
⚙️  Guardian Settings

Configuration:
   • Block Duration: 20 hours
   • Failure Threshold: 5 attempts
   • Auto-cleanup: Enabled
   • Log Level: INFO

Platform: Platform Provider
Build: ` + version.GetBuildTime() + `

Press 'r' to refresh, 'tab' to navigate
	`

	return contentStyle.Render(strings.TrimSpace(content))
}

// renderFooter creates the bottom status bar
func (d *Dashboard) renderFooter() string {
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6b7280")).
		Background(lipgloss.Color("#1f2937")).
		Padding(0, 2).
		Width(d.width)

	controls := "Tab: Navigate • R: Refresh • Q: Quit • Ctrl+C: Quit • TUI is Daemon Status Viewer"
	return footerStyle.Render(controls)
}

// getServiceIcon returns appropriate icon for service status
func (d *Dashboard) getServiceIcon(running bool) string {
	if running {
		return "✅"
	}
	return "❌"
}

// getAutostartIcon returns appropriate icon for autostart status
func (d *Dashboard) getAutostartIcon() string {
	if d.autostartEnabled {
		return "✅ Enabled"
	}
	return "❌ Disabled"
}

// updateRecentLogs reads recent log entries from daemon log file
func (d *Dashboard) updateRecentLogs() {
	// Get log file path based on platform
	var logPath string
	switch runtime.GOOS {
	case "windows":
		if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			logPath = filepath.Join(localAppData, "Guardian", "logs", "guardian-daemon.log")
		}
	case "darwin":
		if home := os.Getenv("HOME"); home != "" {
			logPath = filepath.Join(home, "Library", "Logs", "Guardian", "guardian-daemon.log")
		}
	default: // linux
		logPath = "/var/log/guardian/guardian-daemon.log"
		if _, err := os.Stat(logPath); os.IsNotExist(err) {
			if home := os.Getenv("HOME"); home != "" {
				logPath = filepath.Join(home, ".local", "share", "Guardian", "logs", "guardian-daemon.log")
			}
		}
	}

	if logPath == "" {
		d.recentLogs = []string{"Log path not available"}
		return
	}

	// Read last few lines from log file
	if file, err := os.Open(logPath); err == nil {
		defer file.Close()

		scanner := bufio.NewScanner(file)
		var lines []string

		// Read all lines
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}

		// Keep only last 10 lines
		start := len(lines) - 10
		if start < 0 {
			start = 0
		}

		d.recentLogs = lines[start:]
	} else {
		d.recentLogs = []string{"Unable to read daemon logs: " + err.Error()}
	}
}

// updateAutostartStatus checks and updates the autostart status
func (d *Dashboard) updateAutostartStatus() {
	if execPath, err := autostart.GetExecutablePath(); err == nil {
		autoStart := autostart.New("Guardian", execPath)
		d.autostartEnabled = autoStart.IsEnabled()
	} else {
		d.autostartEnabled = false
	}
}
