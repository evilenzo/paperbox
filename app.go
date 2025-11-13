package main

import (
	"context"
	"fmt"
	"os"

	"paperbox/internal/config"
	"paperbox/models"
)

// App is a thin wrapper for Wails bindings
type App struct {
	ctx       context.Context
	configMgr *config.Manager
}

// NewApp creates a new App instance
func NewApp() *App {
	return &App{
		configMgr: config.NewManager(),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Set context for config manager (needed for events)
	a.configMgr.SetContext(ctx, nil)

	// Load all configurations
	if err := a.configMgr.LoadAll(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to startup application: %v\n", err)
		os.Exit(1)
	}
}

// GetRequests returns the requests for Wails bindings
func (a *App) GetRequests() models.Requests {
	reqConfig := a.configMgr.GetRequests()
	if reqConfig == nil {
		return models.NewRequests()
	}
	return models.Requests{
		Values:    reqConfig.Values,
		RootOrder: reqConfig.RootOrder,
	}
}

// SetRequestsPatch applies a partial update to the requests configuration
func (a *App) SetRequestsPatch(patch models.RequestsPatch) error {
	return a.configMgr.Requests().PatchValues(patch.Values)
}

// AddRequest adds a new request to a parent folder
func (a *App) AddRequest(parentId string, name string, method string, path string) (string, error) {
	return a.configMgr.Requests().AddRequest(parentId, name, method, path)
}

// AddFolder adds a new folder to a parent folder
func (a *App) AddFolder(parentId string, name string) (string, error) {
	return a.configMgr.Requests().AddFolder(parentId, name)
}

// AddRootFolder adds a new root-level folder (without parent)
func (a *App) AddRootFolder(name string) (string, error) {
	return a.configMgr.Requests().AddRootFolder(name)
}

// DeleteItem deletes an item from the requests configuration
func (a *App) DeleteItem(itemId string) error {
	return a.configMgr.Requests().DeleteItem(itemId)
}
