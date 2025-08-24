//go:build !windows
// +build !windows

package platform

import (
	"github.com/sr-tamim/guardian/internal/core"
	"github.com/sr-tamim/guardian/internal/platform/mock"
	"github.com/sr-tamim/guardian/pkg/models"
)

// createWindowsProvider creates a mock provider for non-Windows platforms
func createWindowsProvider(config *models.Config) core.PlatformProvider {
	return mock.NewMockProvider(config)
}
