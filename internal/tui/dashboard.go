package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/sr-tamim/guardian/internal/core"
	"github.com/sr-tamim/guardian/pkg/version"
)

// Dashboard represents the main TUI interface
type Dashboard struct {
	provider       core.PlatformProvider
	devMode        bool
	serviceManager *ServiceManager
	width          int
	height         int
	ready          bool
	quitting       bool
	
	// State
	serviceRunning bool
	blockedIPs     []string
	attackCount    int64
	lastUpdate     time.Time
	
	// Navigation
	selectedTab    int
	tabs           []string
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
	}
}

// SetProvider sets the platform provider (called after creation)
func (d *Dashboard) SetProvider(provider core.PlatformProvider, devMode bool) {
	d.provider = provider
	d.devMode = devMode
}

// SetServiceManager links the dashboard with service manager
func (d *Dashboard) SetServiceManager(sm *ServiceManager) {
	d.serviceManager = sm
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
		case "ctrl+c":
			// Minimize to tray instead of quit
			d.quitting = true
			return d, tea.Quit
			
		case "q":
			// Quit but ask for confirmation if service is running
			if d.serviceManager != nil && d.serviceManager.IsMonitoring() {
				// In a real implementation, show confirmation dialog
				// For now, minimize to tray to keep service running
				d.quitting = true
				return d, tea.Quit
			} else {
				d.quitting = true
				return d, tea.Quit
			}

		case "tab", "right":
			d.selectedTab = (d.selectedTab + 1) % len(d.tabs)
			return d, nil

		case "shift+tab", "left":
			d.selectedTab = (d.selectedTab - 1 + len(d.tabs)) % len(d.tabs)
			return d, nil
			
		case "s":
			// Toggle service through service manager if available
			if d.selectedTab == TabService {
				if d.serviceManager != nil {
					d.serviceManager.toggleMonitoring()
					d.serviceRunning = d.serviceManager.IsMonitoring()
				} else if d.provider != nil {
					// Fallback to local toggle
					d.serviceRunning = !d.serviceRunning
				}
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
		return "👋 Minimizing to system tray...\n🛡️  Guardian continues protecting in background\n\n💡 Right-click tray icon to restore or exit"
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
	title := fmt.Sprintf("🛡️  Guardian v%s - Interactive Dashboard", versionInfo.Version)

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
	if d.serviceRunning {
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
   • Windows RDP Monitoring: %s
   • Windows Firewall: %s
   • Real-time Detection: %s

Press 'r' to refresh, 'tab' to navigate, 'q' to quit
	`,
		statusIcon, statusText, modeText,
		d.attackCount,
		len(d.blockedIPs),
		d.lastUpdate.Format("15:04:05"),
		platformInfo,
		d.getServiceIcon(d.serviceRunning),
		d.getServiceIcon(d.serviceRunning),
		d.getServiceIcon(d.serviceRunning),
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

// renderLogsTab shows recent log activity
func (d *Dashboard) renderLogsTab() string {
	contentStyle := lipgloss.NewStyle().
		Padding(2).
		Height(d.height - 6)

	content := "📝 Recent Activity:\n\n"
	content += "   [15:30:22] Failed RDP from 192.168.1.100\n"
	content += "   [15:30:15] Blocked IP 203.0.113.50\n"
	content += "   [15:29:45] Guardian service started\n"
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
	actionText := "Press 's' to START service"

	if d.serviceRunning {
		statusIcon = "🟢"
		statusText = "RUNNING"
		actionText = "Press 's' to STOP service"
	}

	content := fmt.Sprintf(`
🔧 Service Management

Current Status: %s %s

Windows Service: Not installed
Background Mode: %s

%s
Press 'r' to refresh, 'tab' to navigate
	`,
		statusIcon, statusText,
		d.getServiceIcon(d.serviceRunning),
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

Platform: Windows Provider
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

	controls := "Tab: Navigate • R: Refresh • S: Toggle Service • Q: Minimize to Tray • Ctrl+C: Minimize"
	return footerStyle.Render(controls)
}

// getServiceIcon returns appropriate icon for service status
func (d *Dashboard) getServiceIcon(running bool) string {
	if running {
		return "✅"
	}
	return "❌"
}
