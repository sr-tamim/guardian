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
func (f *Factory) CreateProvider(devMode bool, config *models.Config) (core.PlatformProvider, error) {
	if devMode {
		return mock.NewMockProvider(config), nil
	}

	// In production, detect the actual platform
	switch runtime.GOOS {
	case "windows":
		return createWindowsProvider(config), nil
	case "linux":
		// TODO: return createLinuxProvider(config), nil
		return mock.NewMockProvider(config), nil // Use mock for now
	case "darwin":
		// TODO: return createDarwinProvider(config), nil
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
