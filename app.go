package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"time"

	"IndicoDataFusion/backend"

	"github.com/pkg/errors"
	"github.com/wailsapp/wails/v2/pkg/runtime"
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
	ctx           context.Context
	handler       *backend.DataSourceHandler
	configPath    string
	configFromEnv bool
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

func GetDefaultConfigPath() []string {
	var defaultPaths []string
	switch goruntime.GOOS {
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
		a.configFromEnv = false
	} else {
		a.configFromEnv = true
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
	a.configPath = configPath
	log.Printf("Data handler initialized from: %s\n", configPath)

	a.registerCacheCallbacks()
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

// AppInfo holds application metadata
type AppInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Author      string `json:"author"`
	Company     string `json:"company"`
	AuthorEmail string `json:"authorEmail"`
	BuildDate   string `json:"buildDate"`
}

// GetAppInfo returns application metadata
func (a *App) GetAppInfo() AppInfo {
	return AppInfo{
		Name:        AppName,
		Version:     AppVersion,
		Author:      Author,
		Company:     Company,
		AuthorEmail: AuthorEmail,
		BuildDate:   BuildDate,
	}
}

// GetConfigPath returns the current config path and whether it was from env.
func (a *App) GetConfigPath() backend.ConfigPathInfo {
	return backend.ConfigPathInfo{Path: a.configPath, FromEnv: a.configFromEnv, EnvVarName: ConfEnvName}
}

// GetConfigYAML returns the current YAML content of the config file.
func (a *App) GetConfigYAML() (string, error) {
	if a.configPath == "" {
		return "", errors.Errorf("config path not set")
	}
	b, err := os.ReadFile(a.configPath)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ApplyConfigYAML saves the YAML to the config path and reloads the data handler.
func (a *App) ApplyConfigYAML(yamlContent string) error {
	if a.configPath == "" {
		return errors.Errorf("config path not set")
	}
	// Validate by parsing first
	cfg, err := backend.LoadConfigFromBytes([]byte(yamlContent))
	if err != nil {
		return errors.Wrap(err, "invalid config YAML")
	}
	// Marshal validated cfg back to YAML for normalized save
	if err := backend.SaveConfig(a.configPath, cfg); err != nil {
		return errors.Wrap(err, "failed to save config")
	}
	// Reload handler
	h, err := backend.NewDataSourceHandlerFromConfigFile(a.configPath)
	if err != nil {
		return errors.Wrap(err, "failed to reload handler")
	}
	a.handler = h

	a.registerCacheCallbacks()
	return nil
}

// GetStructuredConfigUI returns the configuration in a structured format for the UI.
func (a *App) GetStructuredConfigUI() (*backend.ConfigDataUI, error) {
	if a.configPath == "" {
		return nil, errors.Errorf("config path not set")
	}

	cfg, err := backend.LoadConfig(a.configPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load config")
	}

	pathInfo := backend.ConfigPathInfo{
		Path:       a.configPath,
		FromEnv:    a.configFromEnv,
		EnvVarName: ConfEnvName,
	}

	return backend.GetStructuredConfigUI(cfg, pathInfo), nil
}

// ApplyStructuredConfigUI applies the structured configuration from the UI.
func (a *App) ApplyStructuredConfigUI(configData *backend.ConfigDataUI) error {
	if a.configPath == "" {
		return errors.Errorf("config path not set")
	}

	// Build the config structure
	cfg := backend.BuildConfigFromStructuredUI(configData)

	// Save the config
	if err := backend.SaveConfig(a.configPath, cfg); err != nil {
		return errors.Wrap(err, "failed to save config")
	}

	// Reload handler
	h, err := backend.NewDataSourceHandlerFromConfigFile(a.configPath)
	if err != nil {
		return errors.Wrap(err, "failed to reload handler")
	}
	a.handler = h

	a.registerCacheCallbacks()
	return nil
}

// RefreshCache invalidates and refreshes a specific cache entry
func (a *App) RefreshCache(key string) error {
	if a.handler == nil {
		return errors.Errorf("data handler not initialized")
	}

	if err := a.handler.RefreshCache(a.ctx, key); err != nil {
		return errors.Wrap(err, "failed to refresh cache")
	}

	// Emit event to notify frontend
	runtime.EventsEmit(a.ctx, "cache:updated", map[string]interface{}{
		"key":    key,
		"action": "refreshed",
	})

	return nil
}

// ClearCache removes all entries from cache and deletes the cache file
func (a *App) ClearCache() error {
	if a.handler == nil {
		return errors.Errorf("data handler not initialized")
	}

	if err := a.handler.ClearCache(); err != nil {
		return errors.Wrap(err, "failed to clear cache")
	}

	// Emit event to notify frontend
	runtime.EventsEmit(a.ctx, "cache:updated", map[string]interface{}{
		"action": "cleared",
	})

	return nil
}

// GetCacheStats returns cache statistics
func (a *App) GetCacheStats() map[string]interface{} {
	if a.handler == nil {
		return map[string]interface{}{
			"error": "data handler not initialized",
		}
	}
	return a.handler.GetCacheStats()
}

// GetCacheKeys returns all available cache keys
func (a *App) GetCacheKeys() []string {
	if a.handler == nil {
		return []string{}
	}
	return a.handler.GetCacheKeys()
}

// IsTestMode returns true if the current data source is test mode (local files)
func (a *App) IsTestMode() bool {
	if a.handler == nil {
		return false
	}
	return a.handler.IsTestMode()
}

// GetCacheEntries returns all cache entries with metadata grouped by data source
func (a *App) GetCacheEntries() map[string][]*backend.CacheEntry {
	if a.handler == nil {
		return make(map[string][]*backend.CacheEntry)
	}
	return a.handler.GetCacheEntries()
}

// shutdown is called when the app is shutting down
func (a *App) shutdown(ctx context.Context) {
	if a.handler != nil {
		if err := a.handler.Shutdown(ctx); err != nil {
			log.Printf("Error during shutdown: %v", err)
		}
	}
}

// registerCacheCallbacks sets up the handler callbacks to forward cache delete/evict events to the frontend.
func (a *App) registerCacheCallbacks() {
	if a == nil || a.handler == nil {
		return
	}

	a.handler.SetCacheOnDelete(func(fullKey string) {
		displayKey := fullKey
		if idx := strings.Index(fullKey, ":"); idx != -1 {
			displayKey = fullKey[idx+1:]
		}
		runtime.EventsEmit(a.ctx, "cache:updated", map[string]interface{}{
			"key":    displayKey,
			"action": "expired",
		})
	})

	a.handler.SetCacheOnEvict(func(fullKey string) {
		displayKey := fullKey
		if idx := strings.Index(fullKey, ":"); idx != -1 {
			displayKey = fullKey[idx+1:]
		}
		runtime.EventsEmit(a.ctx, "cache:updated", map[string]interface{}{
			"key":    displayKey,
			"action": "evicted",
		})
	})
}
