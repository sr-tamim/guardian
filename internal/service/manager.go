package service

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/kardianos/service"
	"github.com/sr-tamim/guardian/internal/daemon"
	"github.com/sr-tamim/guardian/internal/platform"
	"github.com/sr-tamim/guardian/pkg/logger"
	"github.com/sr-tamim/guardian/pkg/models"
)

// Manager configures and controls the Guardian service.
type Manager struct {
	configLoader func() (*models.Config, error)
	devMode      *bool
	configPath   *string
}

// NewManager creates a new service manager.
func NewManager(configLoader func() (*models.Config, error), devMode *bool, configPath *string) *Manager {
	return &Manager{configLoader: configLoader, devMode: devMode, configPath: configPath}
}

// NewService creates the platform service instance.
func (m *Manager) NewService() (service.Service, error) {
	program := &guardianProgram{
		configLoader: m.configLoader,
		devMode:      m.devMode,
		configPath:   m.configPath,
		done:         make(chan struct{}),
	}

	args := []string{"service", "run"}
	if m.devMode != nil && *m.devMode {
		args = append(args, "--dev")
	}
	if m.configPath != nil && *m.configPath != "" {
		args = append(args, "--config", *m.configPath)
	}

	config := &service.Config{
		Name:        "Guardian",
		DisplayName: "Guardian",
		Description: "Guardian Intrusion Prevention System",
		Arguments:   args,
	}

	return service.New(program, config)
}

type guardianProgram struct {
	configLoader func() (*models.Config, error)
	devMode      *bool
	configPath   *string
	cancel       context.CancelFunc
	done         chan struct{}
}

func (p *guardianProgram) Start(_ service.Service) error {
	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel

	go p.run(ctx)
	return nil
}

func (p *guardianProgram) run(ctx context.Context) {
	defer func() {
		if recovered := recover(); recovered != nil {
			logger.Error("Guardian service panic", "panic", recovered)
		}
		close(p.done)
	}()

	config, err := p.configLoader()
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		return
	}

	factory := platform.NewFactory()
	provider, err := factory.CreateProvider(*p.devMode, config)
	if err != nil {
		logger.Error("Failed to create platform provider", "error", err)
		return
	}

	configPath := ""
	if p.configPath != nil {
		configPath = *p.configPath
	}
	manager := daemon.NewManager(config, provider, *p.devMode, configPath)

	if runtime.GOOS == "windows" {
		logger.Info("Guardian service starting (Windows)")
	} else {
		logger.Info("Guardian service starting")
	}

	if err := manager.RunMonitorInCurrentProcess(ctx); err != nil {
		logger.Error("Guardian service stopped with error", "error", err)
		return
	}
}

func (p *guardianProgram) Stop(_ service.Service) error {
	if p.cancel != nil {
		p.cancel()
	}

	select {
	case <-p.done:
		return nil
	case <-time.After(10 * time.Second):
		return fmt.Errorf("timeout stopping Guardian service")
	}
}
