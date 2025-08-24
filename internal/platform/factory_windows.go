//go:build windows
// +build windows

package platform

import (
	"github.com/sr-tamim/guardian/internal/core"
	"github.com/sr-tamim/guardian/internal/platform/windows"
	"github.com/sr-tamim/guardian/pkg/models"
)

// createWindowsProvider creates the Windows-specific provider
func createWindowsProvider(config *models.Config) core.PlatformProvider {
	return windows.NewWindowsProvider(config)
}
