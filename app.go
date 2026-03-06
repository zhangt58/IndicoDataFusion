package main

import (
	"IndicoDataFusion/backend/cache"
	"IndicoDataFusion/backend/config"
	"IndicoDataFusion/backend/consts"
	"IndicoDataFusion/backend/data"
	"IndicoDataFusion/backend/indico"
	"IndicoDataFusion/backend/reviewmode"
	"IndicoDataFusion/backend/utils"
	"context"
	_ "embed"
	"log"
	"net/url"
	"os"
	"os/exec"
	goruntime "runtime"
	"strings"
	"time"

	"github.com/pkg/errors"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed config/sample.yml
var embeddedSample []byte

var (
	BuildDate = time.Now().Format("January 2, 2006")
)

// App struct
type App struct {
	ctx        context.Context
	handler    *data.DataSourceHandler
	configPath string
	// abstractsFile is the optional path from the --abstracts-file CLI flag.
	// When non-empty, all GetAbstracts calls read from this file instead of the
	// configured data source.
	abstractsFile string
	// DataSourceName caches the active data source name from the handler
	DataSourceName string
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

func (a *App) DetermineConfigPath(flagPath string) string {
	return utils.DetermineConfigPath(flagPath, embeddedSample)
}

// startup is called when the app starts. It now requires an explicit config path
// to be provided by the caller.
func (a *App) startup(ctx context.Context, configPath string) {
	a.ctx = ctx

	if configPath == "" {
		log.Printf("Error: no config file path specified (startup requires an explicit path)")
		os.Exit(1)
	}

	handler, err := data.NewDataSourceHandlerFromConfigFile(configPath)
	if err != nil {
		log.Printf("Error: Failed to initialize data handler from %s: %v\n", configPath, err)
		os.Exit(1)
	}
	a.handler = handler
	a.configPath = configPath

	// Apply the abstracts file override if one was provided via --abstracts-file.
	if a.abstractsFile != "" {
		log.Printf("Abstracts override file: %s\n", a.abstractsFile)
		a.handler.SetAbstractsFile(a.abstractsFile)
	}

	// Cache the active data source name on startup
	if a.handler != nil {
		a.DataSourceName = a.handler.GetDataSourceName()
	}
	log.Printf("Data handler initialized from: %s\n", configPath)

	// Notify frontend of the active data source name so UI (titlebar) can display it
	if a.handler != nil {
		wailsruntime.EventsEmit(a.ctx, "app:datasource", a.DataSourceName)
		// Emit init problems so UI can show token-related issues
		wailsruntime.EventsEmit(a.ctx, "app:initproblems", a.GetInitProblems())
	}

	a.registerCacheCallbacks()
}

// GetEventInfo retrieves event information from the configured data source
func (a *App) GetEventInfo() (*indico.Event, error) {
	if a.handler == nil {
		return nil, errors.Errorf("data handler not initialized")
	}
	return a.handler.GetInfo(a.ctx)
}

// GetAbstracts retrieves all abstracts from the configured data source
func (a *App) GetAbstracts() ([]indico.AbstractData, error) {
	if a.handler == nil {
		return nil, errors.Errorf("data handler not initialized")
	}
	return a.handler.GetAbstracts(a.ctx)
}

// GetContributions retrieves all contributions from the configured data source
func (a *App) GetContributions() ([]indico.ContributionData, error) {
	if a.handler == nil {
		return nil, errors.Errorf("data handler not initialized")
	}
	return a.handler.GetContributions(a.ctx)
}

// GetAbstractByID retrieves a specific abstract by ID
func (a *App) GetAbstractByID(id int) (*indico.AbstractData, error) {
	if a.handler == nil {
		return nil, errors.Errorf("data handler not initialized")
	}
	return a.handler.GetAbstractByID(a.ctx, id)
}

// RefreshAbstractByID fetches fresh data for a single abstract from the API
func (a *App) RefreshAbstractByID(id int) (*indico.AbstractData, error) {
	if a.handler == nil {
		return nil, errors.Errorf("data handler not initialized")
	}
	return a.handler.RefreshAbstractByID(a.ctx, id)
}

// GetAbstractsByState filters abstracts by their state
func (a *App) GetAbstractsByState(state string) ([]indico.AbstractData, error) {
	if a.handler == nil {
		return nil, errors.Errorf("data handler not initialized")
	}
	return a.handler.GetAbstractsByState(a.ctx, state)
}

// GetContributionByID retrieves a specific contribution by ID
func (a *App) GetContributionByID(id string) (*indico.ContributionData, error) {
	if a.handler == nil {
		return nil, errors.Errorf("data handler not initialized")
	}
	return a.handler.GetContributionByID(a.ctx, id)
}

// GetContributionsBySession filters contributions by session
func (a *App) GetContributionsBySession(session string) ([]indico.ContributionData, error) {
	if a.handler == nil {
		return nil, errors.Errorf("data handler not initialized")
	}
	return a.handler.GetContributionsBySession(a.ctx, session)
}

// GetContributionsByTrack filters contributions by track
func (a *App) GetContributionsByTrack(track string) ([]indico.ContributionData, error) {
	if a.handler == nil {
		return nil, errors.Errorf("data handler not initialized")
	}
	return a.handler.GetContributionsByTrack(a.ctx, track)
}

// GetReviewTracks returns the review tracks assigned to the current user/data source.
func (a *App) GetReviewTracks() (*indico.ReviewTracks, error) {
	if a.handler == nil {
		return nil, errors.Errorf("data handler not initialized")
	}
	return a.handler.GetReviewTracks(a.ctx)
}

// GetAssignedReviewCount returns the number of unique abstracts assigned to the current user across all review tracks.
func (a *App) GetAssignedReviewCount() (int, error) {
	if a.handler == nil {
		return 0, errors.Errorf("data handler not initialized")
	}
	return a.handler.GetAssignedReviewCount(a.ctx)
}

// GetReviewAbstractIDs returns the list of abstract IDs under a specific review track.
func (a *App) GetReviewAbstractIDs(reviewTrackID int) ([]int, error) {
	if a.handler == nil {
		return nil, errors.Errorf("data handler not initialized")
	}
	return a.handler.GetReviewAbstractIDs(a.ctx, reviewTrackID)
}

// GetVoteStats returns vote statistics per review track for the current reviewer.
// A "vote" is defined as casting a first- or second-priority "yes" answer on any abstract.
// Each track has a maximum of data.MaxVotesPerTrack votes allowed.
func (a *App) GetVoteStats() (*data.VoteStats, error) {
	if a.handler == nil {
		return nil, errors.Errorf("data handler not initialized")
	}
	return a.handler.GetVoteStats(a.ctx)
}

// SubmitAbstractReview submits a new review for an abstract.
// Parameters:
//   - abstractID: the database ID of the abstract
//   - trackID: the track ID to review for
//   - firstPriorityValue: rating value (0 or 1) for first priority question
//   - secondPriorityValue: rating value (0 or 1) for second priority question
//   - proposedAction: the proposed action (accept, reject, change_tracks, mark_as_duplicate, merge)
//   - proposedContribTypeID: proposed contribution type ID (nil for __None)
//   - proposedTrackIDs: proposed track IDs for change_tracks action
//   - proposedRelatedAbstractID: related abstract ID for mark_as_duplicate/merge actions
//   - comment: review comment
func (a *App) SubmitAbstractReview(
	abstractID int,
	trackID int,
	firstPriorityValue int,
	secondPriorityValue int,
	proposedAction string,
	proposedContribTypeID *int,
	proposedTrackIDs []int,
	proposedRelatedAbstractID *int,
	comment string,
) error {
	if a.handler == nil {
		return errors.Errorf("data handler not initialized")
	}

	// Get the abstract
	abstract, err := a.handler.GetAbstractByID(a.ctx, abstractID)
	if err != nil {
		return errors.Wrap(err, "failed to get abstract")
	}

	// Get the client
	client := a.handler.GetClient()
	if client == nil {
		return errors.Errorf("no Indico client available - test mode not supported")
	}

	// Submit the review
	return abstract.SubmitNewReview(
		a.ctx,
		client,
		trackID,
		firstPriorityValue,
		secondPriorityValue,
		proposedAction,
		proposedContribTypeID,
		proposedTrackIDs,
		proposedRelatedAbstractID,
		comment,
	)
}

// UpdateAbstractReview updates an existing review for an abstract.
// Parameters:
//   - abstractID: the database ID of the abstract
//   - reviewID: the review ID to update
//   - trackID: the track ID being reviewed
//   - firstPriorityValue: rating value (0 or 1) for first priority question
//   - secondPriorityValue: rating value (0 or 1) for second priority question
//   - proposedAction: the proposed action (accept, reject, change_tracks, mark_as_duplicate, merge)
//   - proposedContribTypeID: proposed contribution type ID (nil for __None)
//   - proposedTrackIDs: proposed track IDs for change_tracks action
//   - proposedRelatedAbstractID: related abstract ID for mark_as_duplicate/merge actions
//   - comment: review comment
func (a *App) UpdateAbstractReview(
	abstractID int,
	reviewID int,
	trackID int,
	firstPriorityValue int,
	secondPriorityValue int,
	proposedAction string,
	proposedContribTypeID *int,
	proposedTrackIDs []int,
	proposedRelatedAbstractID *int,
	comment string,
) error {
	if a.handler == nil {
		return errors.Errorf("data handler not initialized")
	}

	// Get the abstract
	abstract, err := a.handler.GetAbstractByID(a.ctx, abstractID)
	if err != nil {
		return errors.Wrap(err, "failed to get abstract")
	}

	// Get the client
	client := a.handler.GetClient()
	if client == nil {
		return errors.Errorf("no Indico client available - test mode not supported")
	}

	// Update the review
	return abstract.UpdateReview(
		a.ctx,
		client,
		reviewID,
		trackID,
		firstPriorityValue,
		secondPriorityValue,
		proposedAction,
		proposedContribTypeID,
		proposedTrackIDs,
		proposedRelatedAbstractID,
		comment,
	)
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
		Name:        consts.AppName,
		NameAbbr:    consts.AppNameAbbr,
		Version:     consts.AppVersion,
		Author:      consts.Author,
		Company:     consts.Company,
		AuthorEmail: consts.AuthorEmail,
		BuildDate:   BuildDate,
		DataSource:  a.DataSourceName,
	}
}

// GetConfigPath returns the current config path and whether it was from env.
func (a *App) GetConfigPath() config.ConfigPathInfo {
	return config.ConfigPathInfo{Path: a.configPath}
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
	cfg, err := config.LoadConfigFromBytes([]byte(yamlContent))
	if err != nil {
		return errors.Wrap(err, "invalid config YAML")
	}
	// Marshal validated cfg back to YAML for normalized save
	if err := config.SaveConfig(a.configPath, cfg); err != nil {
		return errors.Wrap(err, "failed to save config")
	}
	// Reload handler
	h, err := data.NewDataSourceHandlerFromConfigFile(a.configPath)
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
		wailsruntime.EventsEmit(a.ctx, "app:datasource", a.DataSourceName)
		// Emit init problems so UI can show token-related issues
		wailsruntime.EventsEmit(a.ctx, "app:initproblems", a.GetInitProblems())
	}

	a.registerCacheCallbacks()
	return nil
}

// GetStructuredConfigUI returns the configuration in a structured format for the UI.
func (a *App) GetStructuredConfigUI() (*config.ConfigDataUI, error) {
	if a.configPath == "" {
		return nil, errors.Errorf("config path not set")
	}

	cfg, err := config.LoadConfig(a.configPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load config")
	}

	pathInfo := config.ConfigPathInfo{
		Path: a.configPath,
	}

	return config.GetStructuredConfigUI(cfg, pathInfo), nil
}

// ApplyStructuredConfigUI applies the structured configuration from the UI.
func (a *App) ApplyStructuredConfigUI(configData *config.ConfigDataUI) error {
	if a.configPath == "" {
		return errors.Errorf("config path not set")
	}

	// Build the config structure
	cfg := config.BuildConfigFromStructuredUI(configData)

	// Save the config
	if err := config.SaveConfig(a.configPath, cfg); err != nil {
		return errors.Wrap(err, "failed to save config")
	}

	// Shutdown the old handler to stop its expiry notification worker
	if a.handler != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := a.handler.Shutdown(shutdownCtx); err != nil {
			log.Printf("Warning: error shutting down old handler: %v", err)
		}
	}

	// Reload handler
	h, err := data.NewDataSourceHandlerFromConfigFile(a.configPath)
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
		wailsruntime.EventsEmit(a.ctx, "app:datasource", a.DataSourceName)
		// Emit init problems so UI can show token-related issues
		wailsruntime.EventsEmit(a.ctx, "app:initproblems", a.GetInitProblems())
	}

	a.registerCacheCallbacks()
	return nil
}

// ExportConfig exports the current configuration with API tokens encrypted with a password.
// Returns the encrypted JSON data as a base64-encoded string.
func (a *App) ExportConfig(password string) (string, error) {
	if a.configPath == "" {
		return "", errors.Errorf("config path not set")
	}

	// Load current config
	cfg, err := config.LoadConfig(a.configPath)
	if err != nil {
		return "", errors.Wrap(err, "failed to load config")
	}

	// Export with encryption, passing token retriever to fetch secrets from keyring
	data, err := config.ExportConfig(cfg, password, utils.GetAPITokenSecret)
	if err != nil {
		return "", errors.Wrap(err, "failed to export config")
	}

	return string(data), nil
}

// ImportConfig imports and decrypts a configuration file, then applies it.
// The data parameter should be the encrypted JSON export.
func (a *App) ImportConfig(encryptedData string, password string) error {
	if a.configPath == "" {
		return errors.Errorf("config path not set")
	}

	// Decrypt and parse the config, passing token storer to save secrets to keyring
	cfg, err := config.ImportConfig([]byte(encryptedData), password, utils.SetAPITokenSecret)
	if err != nil {
		return errors.Wrap(err, "failed to import config")
	}

	// Save the imported config (tokens are now in keyring, config has empty token fields)
	if err := config.SaveConfig(a.configPath, cfg); err != nil {
		return errors.Wrap(err, "failed to save imported config")
	}

	// Reload handler
	h, err := data.NewDataSourceHandlerFromConfigFile(a.configPath)
	if err != nil {
		return errors.Wrap(err, "failed to reload handler after import")
	}
	a.handler = h
	// Cache the active data source name after import
	if a.handler != nil {
		a.DataSourceName = a.handler.GetDataSourceName()
	}

	// Notify frontend of the active data source name after import
	if a.handler != nil {
		wailsruntime.EventsEmit(a.ctx, "app:datasource", a.DataSourceName)
		// Emit init problems so UI can show token-related issues
		wailsruntime.EventsEmit(a.ctx, "app:initproblems", a.GetInitProblems())
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
	wailsruntime.EventsEmit(a.ctx, "cache:updated", map[string]interface{}{
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
	wailsruntime.EventsEmit(a.ctx, "cache:updated", map[string]interface{}{
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
	wailsruntime.EventsEmit(a.ctx, "cache:updated", map[string]interface{}{
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

// ReviewMode returns true when abstract data is being served from
// an --abstracts-file override. In this mode certain UI elements should be
// hidden (e.g., priority ratings, submission tab, and some review analytics).
func (a *App) ReviewMode() bool {
	if a.handler == nil {
		return false
	}
	return a.handler.ReviewMode()
}

// GetVisibilityConfig returns the visibility configuration based on
// whether the app is in review mode or not
func (a *App) GetVisibilityConfig() *reviewmode.VisibilityConfig {
	if a.ReviewMode() {
		return reviewmode.ReviewModeVisibilityConfig()
	}
	return reviewmode.DefaultVisibilityConfig()
}

// GetCacheEntries returns all cache entries with metadata grouped by data source
func (a *App) GetCacheEntries() map[string][]*cache.CacheEntry {
	if a.handler == nil {
		return make(map[string][]*cache.CacheEntry)
	}
	return a.handler.GetCacheEntries()
}

// GetCacheEntryMetadata retrieves metadata for a specific cache entry
func (a *App) GetCacheEntryMetadata(key string) *cache.CacheEntry {
	if a.handler == nil {
		return nil
	}
	entry, found := a.handler.GetCacheEntryMetadata(key)
	if !found {
		return nil
	}
	return entry
}

// AddAPIToken stores the token secret in OS keyring and updates the config metadata
// (without storing the raw token in YAML).
func (a *App) AddAPIToken(entry config.APITokenEntry, rawToken string) error {
	// store in keyring
	if err := utils.SetAPITokenSecret(entry.Name, rawToken); err != nil {
		return errors.Wrap(err, "failed to store token in keyring")
	}

	// load existing structured config
	cfgData, err := a.GetStructuredConfigUI()
	if err != nil {
		return errors.Wrap(err, "failed to load structured config")
	}
	// ensure apiTokens list exists and replace/add entry (but clear token field)
	if cfgData.APITokens == nil {
		cfgData.APITokens = []config.APITokenEntry{}
	}
	found := false
	for i := range cfgData.APITokens {
		if cfgData.APITokens[i].Name == entry.Name {
			cfgData.APITokens[i].BaseURL = entry.BaseURL
			cfgData.APITokens[i].Username = entry.Username
			// clear token in persisted metadata
			cfgData.APITokens[i].Token = ""
			found = true
			break
		}
	}
	if !found {
		entry.Token = ""
		cfgData.APITokens = append(cfgData.APITokens, entry)
	}

	// Save via ApplyStructuredConfigUI which will persist the metadata
	if err := a.ApplyStructuredConfigUI(cfgData); err != nil {
		return errors.Wrap(err, "failed to persist API token metadata")
	}
	return nil
}

// DeleteAPIToken removes the secret from keyring and the metadata from config
func (a *App) DeleteAPIToken(name string) error {
	if name == "" {
		return errors.Errorf("token name required")
	}
	if err := utils.DeleteAPITokenSecret(name); err != nil {
		return errors.Wrap(err, "failed to delete token from keyring")
	}
	cfgData, err := a.GetStructuredConfigUI()
	if err != nil {
		return errors.Wrap(err, "failed to load structured config")
	}
	var newList []config.APITokenEntry
	for _, e := range cfgData.APITokens {
		if e.Name != name {
			newList = append(newList, e)
		}
	}
	cfgData.APITokens = newList
	if err := a.ApplyStructuredConfigUI(cfgData); err != nil {
		return errors.Wrap(err, "failed to persist API token metadata after delete")
	}
	return nil
}

// HasAPITokenSecret checks whether a token with the given name exists in keyring (without returning the secret)
func (a *App) HasAPITokenSecret(name string) (bool, error) {
	if name == "" {
		return false, errors.Errorf("token name required")
	}
	_, err := utils.GetAPITokenSecret(name)
	if err != nil {
		// keyring returns "secret not found" style errors; treat any error as absent
		return false, nil
	}
	return true, nil
}

// RevealAPIToken returns the token value for a given name. Use with care (UI should prompt the user).
func (a *App) RevealAPIToken(name string) (string, error) {
	if name == "" {
		return "", errors.Errorf("token name required")
	}
	tok, err := utils.GetAPITokenSecret(name)
	if err != nil {
		return "", errors.Wrap(err, "failed to read token from keyring")
	}
	return tok, nil
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
		wailsruntime.EventsEmit(a.ctx, "cache:updated", map[string]interface{}{
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
		wailsruntime.EventsEmit(a.ctx, "cache:updated", map[string]interface{}{
			"key":              displayKey,
			"action":           "evicted",
			"data_source_name": dataSourceName,
		})
	})
}

// GetInitProblems returns any non-fatal initialization problems encountered when creating the handler.
func (a *App) GetInitProblems() []string {
	if a == nil || a.handler == nil {
		return []string{}
	}
	return a.handler.GetInitProblems()
}

// OpenSafeURL validates and opens an external URL using the OS default browser.
func (a *App) OpenSafeURL(rawURL string) error {
	if rawURL == "" {
		return errors.Errorf("empty url")
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return errors.Wrap(err, "invalid url")
	}

	u.RawQuery = u.Query().Encode()

	// Use the system default application to open the URL.
	// Cross-platform handling: Linux -> xdg-open, macOS -> open, Windows -> start via cmd.
	s := u.String()
	switch goruntime.GOOS {
	case "linux":
		if err := exec.Command("xdg-open", s).Start(); err != nil {
			return errors.Wrap(err, "failed to open url with xdg-open")
		}
	case "darwin":
		if err := exec.Command("open", s).Start(); err != nil {
			return errors.Wrap(err, "failed to open url with open")
		}
	case "windows":
		// Use cmd /c start which delegates to the default handler. Use 'rundll32' as fallback.
		cmd := exec.Command("cmd", "/c", "start", "", s)
		if err := cmd.Start(); err != nil {
			// Try rundll32 as fallback
			if err2 := exec.Command("rundll32", "url.dll,FileProtocolHandler", s).Start(); err2 != nil {
				return errors.Wrap(err, "failed to open url on windows")
			}
		}
	default:
		// Fallback to Wails browser open if available
		wailsruntime.BrowserOpenURL(a.ctx, s)
	}

	return nil
}

// OpenCacheDirectory opens the cache directory in the system file browser
func (a *App) OpenCacheDirectory() error {
	stats := a.GetCacheStats()
	cacheDir, ok := stats["cache_dir"].(string)
	if !ok || cacheDir == "" {
		return errors.New("cache directory not available")
	}

	// Cross-platform handling: Linux -> xdg-open, macOS -> open, Windows -> explorer
	switch goruntime.GOOS {
	case "linux":
		if err := exec.Command("xdg-open", cacheDir).Start(); err != nil {
			return errors.Wrap(err, "failed to open cache directory with xdg-open")
		}
	case "darwin":
		if err := exec.Command("open", cacheDir).Start(); err != nil {
			return errors.Wrap(err, "failed to open cache directory with open")
		}
	case "windows":
		if err := exec.Command("explorer", cacheDir).Start(); err != nil {
			return errors.Wrap(err, "failed to open cache directory with explorer")
		}
	default:
		return errors.New("unsupported platform")
	}

	return nil
}

// GetWordFrequencies computes word frequencies from input text
func (a *App) GetWordFrequencies(text string, minLength int, topN int, enablePluralNorm bool, customExcludedWords []string) []data.WordFrequency {
	return data.GetWordFrequencies(text, minLength, topN, enablePluralNorm, customExcludedWords)
}
