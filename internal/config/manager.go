package config

import (
	"context"
	"fmt"

	"paperbox/internal/config/requests"
	"paperbox/internal/config/storage"
	"paperbox/internal/config/user"

	"github.com/wailsapp/wails/v2/pkg/logger"
)

// Manager manages all application configurations
// It aggregates all config managers and provides a unified interface
type Manager struct {
	managers []ManagerInterface
	requests *requests.Manager
	user     *user.Manager
}

// NewManager creates a new config manager
func NewManager() *Manager {
	// Create shared storage coordinator for all configs
	fileStorage := storage.NewFileStorage()
	coordinator := storage.NewStorageCoordinator(fileStorage, nil, nil)

	reqMgr := requests.NewManager(coordinator)
	userMgr := user.NewManager(coordinator)

	return &Manager{
		managers: []ManagerInterface{reqMgr, userMgr},
		requests: reqMgr,
		user:     userMgr,
	}
}

// LoadAll loads all configurations
func (m *Manager) LoadAll() error {
	for _, mgr := range m.managers {
		if err := mgr.Load(); err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
	}
	return nil
}

// SetContext sets the Wails runtime context for all config managers
func (m *Manager) SetContext(ctx context.Context, log logger.Logger) {
	for _, mgr := range m.managers {
		mgr.SetContext(ctx, log)
	}
}

// Requests returns the requests config manager
func (m *Manager) Requests() *requests.Manager {
	return m.requests
}

// User returns the user config manager
func (m *Manager) User() *user.Manager {
	return m.user
}

// GetRequests returns the requests configuration (for backward compatibility)
func (m *Manager) GetRequests() *requests.RequestsConfig {
	return m.requests.GetRequestsConfig()
}
