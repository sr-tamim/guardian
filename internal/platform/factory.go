package platform

import (
	"runtime"

	"github.com/sr-tamim/guardian/internal/core"
	"github.com/sr-tamim/guardian/internal/platform/mock"
	"github.com/sr-tamim/guardian/pkg/models"
)

// Factory creates the appropriate platform provider based on the current OS
type Factory struct{}

// NewFactory creates a new platform factory
func NewFactory() *Factory {
	return &Factory{}
}

// CreateProvider creates the appropriate platform provider
// For now, always returns MockProvider. Later we'll add real platform detection.
func (f *Factory) CreateProvider(devMode bool, config *models.Config) (core.PlatformProvider, error) {
	if devMode {
		return mock.NewMockProvider(config), nil
	}

	// In production, detect the actual platform
	switch runtime.GOOS {
	case "windows":
		// TODO: return windows.NewWindowsProvider(config)
		return mock.NewMockProvider(config), nil // Use mock for now
	case "linux":
		// TODO: return linux.NewLinuxProvider(config)
		return mock.NewMockProvider(config), nil // Use mock for now
	case "darwin":
		// TODO: return darwin.NewDarwinProvider(config)
		return mock.NewMockProvider(config), nil // Use mock for now
	default:
		return mock.NewMockProvider(config), nil
	}
}

// DetectPlatform returns the current platform name
func (f *Factory) DetectPlatform() string {
	return runtime.GOOS
}

// IsPlatformSupported checks if the current platform is supported
func (f *Factory) IsPlatformSupported() bool {
	switch runtime.GOOS {
	case "windows", "linux", "darwin":
		return true
	default:
		return false
	}
}
