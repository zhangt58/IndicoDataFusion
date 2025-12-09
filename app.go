package main

import (
	"context"
	_ "embed"
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

//go:embed config/sample.yaml
var embeddedSample []byte

const (
	AppName     = "IndicoDataFusion"
	AppNameAbbr = "IDF"
	AppVersion  = "v1.0.0"
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
	ctx        context.Context
	handler    *backend.DataSourceHandler
	configPath string
	// DataSourceName caches the active data source name from the handler
	DataSourceName string
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

// DetermineConfigPath encapsulates the logic to choose or create a configuration file path.
// Priority: explicit flagPath > existing default paths (GetDefaultConfigPath) > attempt to create default from config/sample.yaml or a placeholder.
// Returns the chosen path (or empty string if none could be created).
func (a *App) DetermineConfigPath(flagPath string) string {
	// 1) If environment variable explicitly specifies a config path, use it.
	// Note: intentionally accept the value as-is (caller decides if it's valid).
	if env := os.Getenv(ConfEnvName); env != "" {
		log.Printf("Using config path from env: %s", env)
		return env
	}

	// 2) If explicit flag provided, use it
	if flagPath != "" {
		return flagPath
	}

	// 3) Look for an existing default config
	for _, p := range GetDefaultConfigPath() {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	// 4) No existing default – attempt to create one
	defaultPaths := GetDefaultConfigPath()
	var target string
	if len(defaultPaths) > 0 {
		target = defaultPaths[0]
	} else if home, err := os.UserHomeDir(); err == nil {
		target = filepath.Join(home, ".config", AppName, "config.yaml")
	} else {
		target = "config.yaml"
	}

	// Ensure parent dir exists
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		log.Printf("Warning: failed to create config dir: %v", err)
	}

	// Prefer writing from embedded sample if available
	if len(embeddedSample) > 0 {
		if err := os.WriteFile(target, embeddedSample, 0644); err != nil {
			log.Printf("Failed to write embedded default config to %s: %v", target, err)
		} else {
			log.Printf("Created default config at %s from embedded sample", target)
			return target
		}
	} else {
		// Fallback: if for some reason embedding is not present, try file system sample
		samplePath := "config/sample.yaml"
		if b, err := os.ReadFile(samplePath); err == nil {
			if werr := os.WriteFile(target, b, 0644); werr != nil {
				log.Printf("Failed to write default config to %s: %v", target, werr)
			} else {
				log.Printf("Created default config at %s from sample file", target)
				return target
			}
		} else {
			log.Printf("Sample file not found at %s: %v", samplePath, err)
		}
	}

	// If sample not available or write failed, write a placeholder
	placeholder := []byte("# IndicoDataFusion default configuration\n")
	if perr := os.WriteFile(target, placeholder, 0644); perr == nil {
		log.Printf("Created placeholder config at %s", target)
		return target
	} else {
		log.Printf("Failed to create placeholder config at %s: %v", target, perr)
	}

	// If creation failed, return empty string
	return ""
}

// startup is called when the app starts. It now requires an explicit config path
// to be provided by the caller.
func (a *App) startup(ctx context.Context, configPath string) {
	a.ctx = ctx

	if configPath == "" {
		log.Printf("Error: no config file path specified (startup requires an explicit path)")
		os.Exit(1)
	}

	handler, err := backend.NewDataSourceHandlerFromConfigFile(configPath)
	if err != nil {
		log.Printf("Error: Failed to initialize data handler from %s: %v\n", configPath, err)
		os.Exit(1)
	}
	a.handler = handler
	a.configPath = configPath
	// Cache the active data source name on startup
	if a.handler != nil {
		a.DataSourceName = a.handler.GetDataSourceName()
	}
	log.Printf("Data handler initialized from: %s\n", configPath)

	// Notify frontend of the active data source name so UI (titlebar) can display it
	if a.handler != nil {
		runtime.EventsEmit(a.ctx, "app:datasource", a.DataSourceName)
	}

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
	NameAbbr    string `json:"nameAbbr"`
	Version     string `json:"version"`
	Author      string `json:"author"`
	Company     string `json:"company"`
	AuthorEmail string `json:"authorEmail"`
	BuildDate   string `json:"buildDate"`
	DataSource  string `json:"dataSource"` // New field for data source name
}

// GetAppInfo returns application metadata
func (a *App) GetAppInfo() AppInfo {
	return AppInfo{
		Name:        AppName,
		NameAbbr:    AppNameAbbr,
		Version:     AppVersion,
		Author:      Author,
		Company:     Company,
		AuthorEmail: AuthorEmail,
		BuildDate:   BuildDate,
		DataSource:  a.DataSourceName,
	}
}

// GetConfigPath returns the current config path and whether it was from env.
func (a *App) GetConfigPath() backend.ConfigPathInfo {
	return backend.ConfigPathInfo{Path: a.configPath}
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
	// Cache the active data source name after reload
	if a.handler != nil {
		a.DataSourceName = a.handler.GetDataSourceName()
	}

	// Notify frontend of the active data source name after reload
	if a.handler != nil {
		runtime.EventsEmit(a.ctx, "app:datasource", a.DataSourceName)
	}

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
		Path: a.configPath,
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
	// Cache the active data source name after structured config reload
	if a.handler != nil {
		a.DataSourceName = a.handler.GetDataSourceName()
	}

	// Notify frontend of the active data source name after structured config reload
	if a.handler != nil {
		runtime.EventsEmit(a.ctx, "app:datasource", a.DataSourceName)
	}

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
		"key":              key,
		"action":           "refreshed",
		"data_source_name": a.DataSourceName,
	})

	return nil
}

// DeleteCacheEntry removes a specific entry from cache
func (a *App) DeleteCacheEntry(key string) error {
	if a.handler == nil {
		return errors.Errorf("data handler not initialized")
	}

	if err := a.handler.DeleteCacheEntry(key); err != nil {
		return errors.Wrap(err, "failed to delete cache entry")
	}

	// Emit event to notify frontend
	runtime.EventsEmit(a.ctx, "cache:updated", map[string]interface{}{
		"key":              key,
		"action":           "deleted",
		"data_source_name": a.DataSourceName,
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
		"action":           "cleared",
		"data_source_name": a.DataSourceName,
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

// registerCacheCallbacks sets up the handler callbacks to forward cache expiry/evict events to the frontend.
func (a *App) registerCacheCallbacks() {
	if a == nil || a.handler == nil {
		return
	}

	a.handler.SetCacheOnExpiry(func(fullKey string) {
		displayKey := fullKey
		dataSourceName := ""
		if idx := strings.Index(fullKey, ":"); idx != -1 {
			dataSourceName = fullKey[:idx]
			displayKey = fullKey[idx+1:]
		}
		runtime.EventsEmit(a.ctx, "cache:updated", map[string]interface{}{
			"key":              displayKey,
			"action":           "expired",
			"data_source_name": dataSourceName,
		})
	})

	a.handler.SetCacheOnEvict(func(fullKey string) {
		displayKey := fullKey
		dataSourceName := ""
		if idx := strings.Index(fullKey, ":"); idx != -1 {
			dataSourceName = fullKey[:idx]
			displayKey = fullKey[idx+1:]
		}
		runtime.EventsEmit(a.ctx, "cache:updated", map[string]interface{}{
			"key":              displayKey,
			"action":           "evicted",
			"data_source_name": dataSourceName,
		})
	})
}
