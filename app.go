package main

import (
	"context"
	"log"

	"paperbox/internal/app"
	"paperbox/models"
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
		log.Fatalf("Failed to startup app: %v", err)
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
