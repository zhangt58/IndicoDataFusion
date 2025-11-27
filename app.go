package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"IndicoDataFusion/backend"

	"github.com/pkg/errors"
)

const (
	AppName     = "IndicoDataFusion"
	AppVersion  = "v0.1.0"
	Author      = "Tong Zhang"
	Company     = "Michigan State University"
	AuthorEmail = "zhangt@frib.msu.edu"
	ConfEnvName = "INDICO_DATA_FUSION_CONFIG_PATH"
)

var (
	BuildDate = time.Now().Format("January 2, 2006")
)

// App struct
type App struct {
	ctx     context.Context
	handler *backend.DataSourceHandler
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

func GetDefaultConfigPath() []string {
	var defaultPaths []string
	switch runtime.GOOS {
	case "windows":
		if appdata := os.Getenv("APPDATA"); appdata != "" {
			defaultPaths = append(defaultPaths,
				filepath.Join(appdata, AppName, "config.yaml"),
				filepath.Join(appdata, AppName, "config.yml"))
		}
	case "darwin":
		// macOS (Library/Application Support)
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			defaultPaths = append(defaultPaths,
				filepath.Join(home, "Library", "Application Support", AppName, "config.yaml"),
				filepath.Join(home, "Library", "Application Support", AppName, "config.yml"))
		}
	default: // linux and others
		if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
			defaultPaths = append(defaultPaths,
				filepath.Join(xdg, ".config", AppName, "config.yaml"),
				filepath.Join(xdg, ".config", AppName, "config.yml"))
		} else if home, err := os.UserHomeDir(); err == nil && home != "" {
			defaultPaths = append(defaultPaths,
				filepath.Join(home, ".config", AppName, "config.yaml"),
				filepath.Join(home, ".config", AppName, "config.yml"))
		}
	}
	return defaultPaths
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Initialize the data source handler from config
	// Check for config path from environment variable first
	configPath := os.Getenv(ConfEnvName)
	if configPath == "" {
		for _, path := range GetDefaultConfigPath() {
			if _, err := os.Stat(path); err == nil {
				configPath = path
				break
			}
		}
	}

	if configPath == "" {
		log.Printf("Error: no config file path specified (%s not set and no default available)\n", ConfEnvName)
		os.Exit(1)
	}

	handler, err := backend.NewDataSourceHandlerFromConfigFile(configPath)
	if err != nil {
		log.Printf("Error: Failed to initialize data handler from %s: %v\n", configPath, err)
		os.Exit(1)
	}
	a.handler = handler
	log.Printf("Data handler initialized from: %s\n", configPath)
}

// GetEventInfo retrieves event information from the configured data source
func (a *App) GetEventInfo() (*backend.Event, error) {
	if a.handler == nil {
		return nil, errors.Errorf("data handler not initialized")
	}
	return a.handler.GetInfo(a.ctx)
}

// GetAbstracts retrieves all abstracts from the configured data source
func (a *App) GetAbstracts() ([]backend.AbstractData, error) {
	if a.handler == nil {
		return nil, errors.Errorf("data handler not initialized")
	}
	return a.handler.GetAbstracts(a.ctx)
}

// GetContributions retrieves all contributions from the configured data source
func (a *App) GetContributions() ([]backend.ContributionData, error) {
	if a.handler == nil {
		return nil, errors.Errorf("data handler not initialized")
	}
	return a.handler.GetContributions(a.ctx)
}

// GetAbstractByID retrieves a specific abstract by ID
func (a *App) GetAbstractByID(id int) (*backend.AbstractData, error) {
	if a.handler == nil {
		return nil, errors.Errorf("data handler not initialized")
	}
	return a.handler.GetAbstractByID(a.ctx, id)
}

// GetAbstractsByState filters abstracts by their state
func (a *App) GetAbstractsByState(state string) ([]backend.AbstractData, error) {
	if a.handler == nil {
		return nil, errors.Errorf("data handler not initialized")
	}
	return a.handler.GetAbstractsByState(a.ctx, state)
}

// GetContributionByID retrieves a specific contribution by ID
func (a *App) GetContributionByID(id string) (*backend.ContributionData, error) {
	if a.handler == nil {
		return nil, errors.Errorf("data handler not initialized")
	}
	return a.handler.GetContributionByID(a.ctx, id)
}

// GetContributionsBySession filters contributions by session
func (a *App) GetContributionsBySession(session string) ([]backend.ContributionData, error) {
	if a.handler == nil {
		return nil, errors.Errorf("data handler not initialized")
	}
	return a.handler.GetContributionsBySession(a.ctx, session)
}

// GetContributionsByTrack filters contributions by track
func (a *App) GetContributionsByTrack(track string) ([]backend.ContributionData, error) {
	if a.handler == nil {
		return nil, errors.Errorf("data handler not initialized")
	}
	return a.handler.GetContributionsByTrack(a.ctx, track)
}
