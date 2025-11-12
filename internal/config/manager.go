package config

import (
	"fmt"

	"paperbox/internal/config/requests"
)

// Manager manages all application configurations
type Manager struct {
	requests *requests.RequestsConfig
}

// NewManager creates a new config manager
func NewManager() *Manager {
	return &Manager{}
}

// LoadAll loads all configurations
func (m *Manager) LoadAll() error {
	// Load requests config
	reqConfig, err := requests.Load()
	if err != nil {
		return fmt.Errorf("failed to load requests config: %w", err)
	}
	m.requests = reqConfig

	// TODO: Load other configs here (user config, etc.)

	return nil
}

// GetRequests returns the requests configuration
func (m *Manager) GetRequests() *requests.RequestsConfig {
	return m.requests
}

// SaveRequests saves the requests configuration
func (m *Manager) SaveRequests() error {
	if m.requests == nil {
		return fmt.Errorf("requests config is not loaded")
	}
	return requests.Save(m.requests)
}
