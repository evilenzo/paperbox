package main

import (
	"context"
	"fmt"

	"paperbox/internal/app"
	"paperbox/models"

	"github.com/wailsapp/wails/v2/pkg/logger"
)

// App is a thin wrapper for Wails bindings
type App struct {
	app *app.App
}

// NewApp creates a new App instance
func NewApp() *App {
	return &App{
		app: app.New(),
	}
}

// startup initializes the application (called by Wails)
func (a *App) startup(ctx context.Context) {
	if err := a.app.Startup(ctx); err != nil {
		logger.NewDefaultLogger().Fatal(err.Error())
	}
}

// GetRequests returns the requests for Wails bindings
func (a *App) GetRequests() models.Requests {
	requestsMap := a.app.GetRequests()
	if requestsMap == nil {
		return models.NewRequests()
	}
	return models.Requests{
		Values: requestsMap,
	}
}

// SetRequestsPatch applies a partial update to the requests configuration
func (a *App) SetRequestsPatch(patch map[string]interface{}) error {
	log := logger.NewDefaultLogger()
	log.Info("SetRequestsPatch called from Wails binding")
	if patch == nil {
		log.Error("SetRequestsPatch: patch is nil")
		return fmt.Errorf("patch is nil")
	}
	log.Info(fmt.Sprintf("SetRequestsPatch: patch has %d keys", len(patch)))
	if values, ok := patch["values"].(map[string]interface{}); ok {
		log.Info(fmt.Sprintf("SetRequestsPatch: values contains %d items", len(values)))
	} else {
		log.Error("SetRequestsPatch: values is not a map[string]interface{}")
	}
	
	err := a.app.SetRequestsPatch(patch)
	if err != nil {
		log.Error(fmt.Sprintf("SetRequestsPatch error: %v", err))
	} else {
		log.Info("SetRequestsPatch completed successfully")
	}
	return err
}

// AddRequest adds a new request to a parent folder
func (a *App) AddRequest(parentId string, name string, method string, path string) (string, error) {
	return a.app.AddRequest(parentId, name, method, path)
}

// AddFolder adds a new folder to a parent folder
func (a *App) AddFolder(parentId string, name string) (string, error) {
	return a.app.AddFolder(parentId, name)
}
