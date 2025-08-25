package tui

import (
	"context"
	"fmt"
	"os"
	"time"

	"fyne.io/systray"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/sr-tamim/guardian/internal/core"
)

// ServiceManager handles TUI and background service integration
type ServiceManager struct {
	provider   core.PlatformProvider
	devMode    bool
	monitoring bool
	ctx        context.Context
	cancel     context.CancelFunc
	startup    StartupManager
}

// NewServiceManager creates a new service manager
func NewServiceManager(provider core.PlatformProvider, devMode bool) *ServiceManager {
	ctx, cancel := context.WithCancel(context.Background())

	startup := CreateStartupManager() // Use platform-agnostic creation

	return &ServiceManager{
		provider: provider,
		devMode:  devMode,
		ctx:      ctx,
		cancel:   cancel,
		startup:  startup,
	}
}

// StartWithTraySupport starts the service with TUI and system tray integration
func (sm *ServiceManager) StartWithTraySupport() error {
	// Start system tray first - this will block until tray exits
	// TUI will be launched from within the tray menu
	sm.initSystemTray()
	return nil
}

// initSystemTray sets up the system tray functionality
func (sm *ServiceManager) initSystemTray() {
	systray.Run(sm.onTrayReady, sm.onTrayExit)
}

// onTrayReady is called when system tray is ready
func (sm *ServiceManager) onTrayReady() {
	// Initialize tray display (platform-specific)
	initializeTrayDisplay()

	// Create menu items
	mShowDashboard := systray.AddMenuItem("Show Dashboard", "Open Guardian TUI Dashboard")
	mStatus := systray.AddMenuItem("Status", "Show Guardian Status")
	systray.AddSeparator()
	mStartStop := systray.AddMenuItem("Start Monitoring", "Start/Stop Guardian monitoring")
	systray.AddSeparator()

	// Platform-aware auto-start menu
	autoStartDescription := fmt.Sprintf("Enable startup (%s)", sm.startup.GetDescription())
	mAutoStart := systray.AddMenuItem("Auto-start on boot", autoStartDescription)
	if sm.startup.IsEnabled() {
		mAutoStart.Check()
	}

	systray.AddSeparator()
	mExit := systray.AddMenuItem("Exit Guardian", "Stop Guardian and exit")

	// Automatically show TUI dashboard on first launch
	go func() {
		sm.showTUI()
	}()

	// Handle menu actions
	go func() {
		for {
			select {
			case <-mShowDashboard.ClickedCh:
				sm.showTUI()

			case <-mStatus.ClickedCh:
				sm.showStatus()

			case <-mStartStop.ClickedCh:
				sm.toggleMonitoring()
				if sm.monitoring {
					mStartStop.SetTitle("Stop Monitoring")
				} else {
					mStartStop.SetTitle("Start Monitoring")
				}

			case <-mAutoStart.ClickedCh:
				if mAutoStart.Checked() {
					sm.disableAutoStart()
					mAutoStart.Uncheck()
				} else {
					sm.enableAutoStart()
					mAutoStart.Check()
				}

			case <-mExit.ClickedCh:
				sm.exitApplication()
				return
			}
		}
	}()
}

// onTrayExit is called when system tray is exiting
func (sm *ServiceManager) onTrayExit() {
	sm.cancel()
}

// showTUI brings TUI to foreground (called from tray)
func (sm *ServiceManager) showTUI() {
	// Launch TUI in a separate goroutine so it doesn't block the tray
	go func() {
		dashboard := NewDashboard()
		dashboard.SetProvider(sm.provider, sm.devMode)

		p := tea.NewProgram(
			dashboard,
			tea.WithAltScreen(),
			tea.WithMouseCellMotion(),
		)

		// Run TUI - when it exits, return to tray (don't exit the whole app)
		_, err := p.Run()
		if err != nil {
			fmt.Printf("TUI error: %v\n", err)
		}

		// TUI closed - Guardian continues in system tray
		fmt.Println("📱 Guardian minimized to system tray - protection continues")
	}()
}

// showStatus shows a status notification
func (sm *ServiceManager) showStatus() {
	status := "Guardian Status:\n"
	if sm.monitoring {
		status += "🟢 Monitoring: Active\n"
	} else {
		status += "🔴 Monitoring: Stopped\n"
	}
	status += fmt.Sprintf("Platform: %s\n", sm.provider.Name())
	if sm.devMode {
		status += "Mode: Development"
	} else {
		status += "Mode: Production"
	}

	// This would show a native notification in a real implementation
	fmt.Println(status)
}

// toggleMonitoring starts/stops the monitoring service
func (sm *ServiceManager) toggleMonitoring() {
	if sm.monitoring {
		sm.stopMonitoring()
	} else {
		sm.startMonitoring()
	}
}

// startMonitoring begins the monitoring process
func (sm *ServiceManager) startMonitoring() {
	sm.monitoring = true
	updateTrayTooltip("Guardian - Protection Active")

	// Start background monitoring
	go sm.runBackgroundMonitoring()
}

// stopMonitoring stops the monitoring process
func (sm *ServiceManager) stopMonitoring() {
	sm.monitoring = false
	updateTrayTooltip("Guardian - Protection Stopped")
}

// runBackgroundMonitoring runs the actual monitoring service
func (sm *ServiceManager) runBackgroundMonitoring() {
	for sm.monitoring {
		select {
		case <-sm.ctx.Done():
			return
		default:
			// Simulate monitoring work
			if sm.provider != nil {
				// In real implementation: check logs, detect threats, block IPs
				time.Sleep(5 * time.Second)
			}
		}
	}
}

// enableAutoStart adds Guardian to system startup
func (sm *ServiceManager) enableAutoStart() error {
	err := sm.startup.Enable()
	if err != nil {
		fmt.Printf("❌ Failed to enable auto-start: %v\n", err)
		return err
	}

	fmt.Printf("✅ Auto-start enabled (%s)\n", sm.startup.GetDescription())
	return nil
}

// disableAutoStart removes Guardian from system startup
func (sm *ServiceManager) disableAutoStart() error {
	err := sm.startup.Disable()
	if err != nil {
		fmt.Printf("❌ Failed to disable auto-start: %v\n", err)
		return err
	}

	fmt.Printf("❌ Auto-start disabled (%s)\n", sm.startup.GetDescription())
	return nil
}

// exitApplication cleanly shuts down Guardian
func (sm *ServiceManager) exitApplication() {
	sm.monitoring = false
	sm.cancel()
	systray.Quit()
	os.Exit(0)
}

// IsMonitoring returns current monitoring status
func (sm *ServiceManager) IsMonitoring() bool {
	return sm.monitoring
}

// GetProvider returns the platform provider
func (sm *ServiceManager) GetProvider() core.PlatformProvider {
	return sm.provider
}
