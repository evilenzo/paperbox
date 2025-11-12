package app

import (
	"context"
	"fmt"

	"paperbox/internal/config"
	"paperbox/internal/config/requests"
)

// App represents the application logic
type App struct {
	ctx       context.Context
	configMgr *config.Manager
}

// New creates a new App instance
func New() *App {
	return &App{
		configMgr: config.NewManager(),
	}
}

// Startup initializes the application
func (a *App) Startup(ctx context.Context) error {
	a.ctx = ctx

	// Load all configurations
	if err := a.configMgr.LoadAll(); err != nil {
		return fmt.Errorf("failed to load configs: %w", err)
	}

	return nil
}

// GetRequests returns the requests from config
func (a *App) GetRequests() map[string]requests.Item {
	reqConfig := a.configMgr.GetRequests()
	if reqConfig == nil {
		return make(map[string]requests.Item)
	}
	return reqConfig.Values
}
