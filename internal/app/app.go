package app

import (
	"context"
	"fmt"

	"paperbox/internal/config"
	"paperbox/internal/config/requests"
	"github.com/wailsapp/wails/v2/pkg/logger"
)

// App represents the application logic
type App struct {
	ctx       context.Context
	logger    logger.Logger
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
	a.logger = logger.NewDefaultLogger()

	// Set context for config manager (needed for events)
	a.configMgr.SetContext(ctx, a.logger)

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

// SetRequestsPatch applies a partial update to the requests configuration
func (a *App) SetRequestsPatch(patch map[string]interface{}) error {
	a.logger.Info("SetRequestsPatch called in internal/app/app.go")
	if patch == nil {
		a.logger.Error("SetRequestsPatch: patch is nil")
		return fmt.Errorf("patch is nil")
	}
	a.logger.Info(fmt.Sprintf("SetRequestsPatch: calling Requests().Patch with %d keys", len(patch)))
	
	err := a.configMgr.Requests().Patch(patch)
	if err != nil {
		a.logger.Error(fmt.Sprintf("SetRequestsPatch error: %v", err))
	} else {
		a.logger.Info("SetRequestsPatch completed successfully in internal/app/app.go")
	}
	return err
}

// AddRequest adds a new request to a parent folder
func (a *App) AddRequest(parentId string, name string, method string, path string) (string, error) {
	return a.configMgr.Requests().AddRequest(parentId, name, method, path)
}

// AddFolder adds a new folder to a parent folder
func (a *App) AddFolder(parentId string, name string) (string, error) {
	return a.configMgr.Requests().AddFolder(parentId, name)
}
