package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DataSourceHandler provides a high-level interface for accessing event data
// from different sources (Indico API or local test files).
type DataSourceHandler struct {
	config     *DataSource
	client     *IndicoClient
	dataDir    string
	isTestMode bool
	cache      *Cache
}

// NewDataSourceHandler creates a new data source handler from a DataSource configuration.
func NewDataSourceHandler(ds *DataSource) (*DataSourceHandler, error) {
	return NewDataSourceHandlerWithCache(ds, nil)
}

// NewDataSourceHandlerWithCache creates a new data source handler with optional cache config
func NewDataSourceHandlerWithCache(ds *DataSource, cacheConfig *CacheConfig) (*DataSourceHandler, error) {
	handler := &DataSourceHandler{
		config: ds,
	}

	// Parse cache configuration
	ttl := 24 * time.Hour               // Default: 24 hours
	maxSize := int64(100 * 1024 * 1024) // Default: 100 MB
	cacheDir := ""                      // Use default if not specified

	if cacheConfig != nil {
		if cacheConfig.TTL != "" {
			if parsedTTL, err := time.ParseDuration(cacheConfig.TTL); err == nil {
				ttl = parsedTTL
			} else {
				log.Printf("Warning: invalid cache TTL '%s', using default 24h", cacheConfig.TTL)
			}
		}
		if cacheConfig.MaxSize != "" {
			if parsedSize, err := parseSize(cacheConfig.MaxSize); err == nil {
				maxSize = parsedSize
			} else {
				log.Printf("Warning: invalid cache max_size '%s', using default 100MB", cacheConfig.MaxSize)
			}
		}
		if cacheConfig.CacheDir != "" {
			cacheDir = cacheConfig.CacheDir
		}
	}

	// Initialize cache
	cache, err := NewCache(CacheOptions{
		CacheDir:       cacheDir,
		LoadOnStartup:  true,
		TTL:            ttl,
		MaxSize:        maxSize,
		DataSourceName: ds.Name,
	})
	if err != nil {
		log.Printf("Warning: failed to initialize cache: %v", err)
		// Continue without cache
	}
	handler.cache = cache

	if ds.Indico != nil {
		// Initialize Indico client
		client := NewIndicoClient(
			ds.Indico.BaseURL,
			ds.Indico.EventID,
			ds.Indico.APIToken,
		)
		if ds.Indico.Timeout != "" {
			if timeout, err := time.ParseDuration(ds.Indico.Timeout); err == nil {
				client.Timeout = timeout
			}
		}
		handler.client = client
		handler.isTestMode = false
	} else if ds.Test != nil {
		// Test mode with local files
		handler.dataDir, _ = filepath.Abs(ds.Test.DataDir)
		handler.isTestMode = true
	} else {
		return nil, fmt.Errorf("data source %s has no valid configuration", ds.Name)
	}

	return handler, nil
}

// parseSize parses size strings like "100MB", "1GB", "512KB"
func parseSize(sizeStr string) (int64, error) {
	sizeStr = strings.ToUpper(strings.TrimSpace(sizeStr))

	// Check suffixes in order from longest to shortest to avoid "MB" matching "B"
	suffixes := []struct {
		suffix     string
		multiplier int64
	}{
		{"GB", 1024 * 1024 * 1024},
		{"MB", 1024 * 1024},
		{"KB", 1024},
		{"B", 1},
	}

	for _, s := range suffixes {
		if strings.HasSuffix(sizeStr, s.suffix) {
			numStr := strings.TrimSpace(strings.TrimSuffix(sizeStr, s.suffix))
			var num float64
			if _, err := fmt.Sscanf(numStr, "%f", &num); err != nil {
				return 0, fmt.Errorf("invalid size format: %s", sizeStr)
			}
			return int64(num * float64(s.multiplier)), nil
		}
	}

	return 0, fmt.Errorf("invalid size format: %s (use B, KB, MB, or GB)", sizeStr)
}

// NewDataSourceHandlerFromConfig creates a handler from a full Config using the default data source.
func NewDataSourceHandlerFromConfig(cfg *Config) (*DataSourceHandler, error) {
	ds, err := cfg.GetActiveDataSource()
	if err != nil {
		return nil, fmt.Errorf("failed to get active data source: %w", err)
	}
	return NewDataSourceHandlerWithCache(ds, cfg.Cache)
}

// NewDataSourceHandlerFromConfigFile creates a handler by loading a config file.
func NewDataSourceHandlerFromConfigFile(configPath string) (*DataSourceHandler, error) {
	cfg, err := LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return NewDataSourceHandlerFromConfig(cfg)
}

// Shutdown gracefully shuts down the handler, saving cache to disk
func (h *DataSourceHandler) Shutdown(ctx context.Context) error {
	if h.cache != nil {
		return h.cache.Shutdown(ctx)
	}
	return nil
}

// GetInfo retrieves event information from the configured data source.
func (h *DataSourceHandler) GetInfo(ctx context.Context) (*Event, error) {
	if h.isTestMode {
		return h.getInfoFromFile()
	}

	// Check cache first in API mode
	if h.cache != nil {
		if cached, found := h.cache.Get("event_info"); found {
			log.Printf("Using cached event info")
			if event, ok := cached.(*Event); ok {
				return event, nil
			}
		}
	}

	// Fetch from API
	event, err := h.getInfoFromAPI(ctx)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if h.cache != nil {
		h.cache.Set("event_info", event)
	}

	return event, nil
}

// getInfoFromFile reads event info from a local JSON file (test mode).
func (h *DataSourceHandler) getInfoFromFile() (*Event, error) {
	if h.config.Test == nil {
		return nil, fmt.Errorf("test configuration not available")
	}

	filePath := filepath.Join(h.dataDir, h.config.Test.EventInfo)
	log.Printf("Reading event info from: %v\n", filePath)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", filePath, err)
	}

	var ev EventAPIResponse
	if err := json.Unmarshal(data, &ev); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", filePath, err)
	}

	return &ev.Results[0], nil
}

// getInfoFromAPI fetches event info from the Indico API.
func (h *DataSourceHandler) getInfoFromAPI(ctx context.Context) (*Event, error) {
	if h.client == nil {
		return nil, fmt.Errorf("indico client not initialized")
	}
	log.Printf("Reading event info from Indico API\n")
	return h.client.GetEventInfo()
}

// GetAbstracts retrieves abstract data from the configured data source.
func (h *DataSourceHandler) GetAbstracts(ctx context.Context) ([]AbstractData, error) {
	if h.isTestMode {
		return h.getAbstractsFromFile()
	}

	// Check cache first in API mode
	if h.cache != nil {
		if cached, found := h.cache.Get("abstracts"); found {
			log.Printf("Using cached abstracts")
			if abstracts, ok := cached.([]AbstractData); ok {
				return abstracts, nil
			}
		}
	}

	// Fetch from API
	abstracts, err := h.getAbstractsFromAPI(ctx)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if h.cache != nil {
		h.cache.Set("abstracts", abstracts)
	}

	return abstracts, nil
}

// getAbstractsFromFile reads abstracts from a local JSON file (test mode).
func (h *DataSourceHandler) getAbstractsFromFile() ([]AbstractData, error) {
	if h.config.Test == nil {
		return nil, fmt.Errorf("test configuration not available")
	}

	filePath := filepath.Join(h.dataDir, h.config.Test.Abstracts)
	log.Printf("Reading abstract data from: %v\n", filePath)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", filePath, err)
	}

	var response AbstractsResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", filePath, err)
	}

	return response.Abstracts, nil
}

// getAbstractsFromAPI fetches abstracts from the Indico API.
// This is a placeholder for future implementation that would fetch from the live API.
func (h *DataSourceHandler) getAbstractsFromAPI(ctx context.Context) ([]AbstractData, error) {
	if h.client == nil {
		return nil, fmt.Errorf("indico client not initialized")
	}

	log.Printf("Reading abstract data from Indico API\n")

	// Fetch the abstracts list page to get IDs and CSRF token
	ids, csrfToken, err := h.client.GetAbstractIDsAndCSRFFromList(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get abstract IDs: %w", err)
	}

	if len(ids) == 0 {
		return []AbstractData{}, nil
	}

	// Fetch the abstracts data
	rawData, err := h.client.FetchAbstractsData(ctx, ids, csrfToken)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch abstracts data: %w", err)
	}

	// Parse the response
	abstractsIface, ok := rawData["abstracts"]
	if !ok {
		return nil, fmt.Errorf("no abstracts field in response")
	}

	// Convert to JSON and back to properly deserialize
	jsonData, err := json.Marshal(map[string]any{"abstracts": abstractsIface})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal abstracts: %w", err)
	}

	var response AbstractsResponse
	if err := json.Unmarshal(jsonData, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal abstracts: %w", err)
	}

	return response.Abstracts, nil
}

// GetContributions retrieves contribution data from the configured data source.
func (h *DataSourceHandler) GetContributions(ctx context.Context) ([]ContributionData, error) {
	if h.isTestMode {
		return h.getContributionsFromFile()
	}

	// Check cache first in API mode
	if h.cache != nil {
		if cached, found := h.cache.Get("contributions"); found {
			log.Printf("Using cached contributions")
			if contribs, ok := cached.([]ContributionData); ok {
				return contribs, nil
			}
		}
	}

	// Fetch from API
	contribs, err := h.getContributionsFromAPI(ctx)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if h.cache != nil {
		h.cache.Set("contributions", contribs)
	}

	return contribs, nil
}

// getContributionsFromFile reads contributions from a local JSON file (test mode).
func (h *DataSourceHandler) getContributionsFromFile() ([]ContributionData, error) {
	if h.config.Test == nil {
		return nil, fmt.Errorf("test configuration not available")
	}

	filePath := filepath.Join(h.dataDir, h.config.Test.Contribs)
	log.Printf("Reading contribution data from: %v\n", filePath)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", filePath, err)
	}

	var response ContributionsAPIResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", filePath, err)
	}

	if len(response.Results) == 0 {
		return []ContributionData{}, nil
	}

	return response.Results[0].Contributions, nil
}

// getContributionsFromAPI fetches contributions from the Indico API.
func (h *DataSourceHandler) getContributionsFromAPI(ctx context.Context) ([]ContributionData, error) {
	if h.client == nil {
		return nil, fmt.Errorf("indico client not initialized")
	}

	log.Printf("Reading contribution data from Indico API\n")

	// Construct the API path for contributions
	path := fmt.Sprintf("/export/event/%d.json", h.client.EventID)

	queryValues := url.Values{}
	queryValues.Set("detail", "contributions")

	// Fetch the contribution data
	var response ContributionsAPIResponse
	if err := h.client.doGet(ctx, path, queryValues, &response); err != nil {
		return nil, fmt.Errorf("failed to fetch contributions: %w", err)
	}

	if len(response.Results) == 0 {
		return []ContributionData{}, nil
	}

	return response.Results[0].Contributions, nil
}

// GetContributionByID retrieves a specific contribution by ID.
func (h *DataSourceHandler) GetContributionByID(ctx context.Context, id string) (*ContributionData, error) {
	contributions, err := h.GetContributions(ctx)
	if err != nil {
		return nil, err
	}

	for _, contrib := range contributions {
		if contrib.ID == id {
			return &contrib, nil
		}
	}

	return nil, fmt.Errorf("contribution with ID %s not found", id)
}

// GetAbstractByID retrieves a specific abstract by ID.
func (h *DataSourceHandler) GetAbstractByID(ctx context.Context, id int) (*AbstractData, error) {
	abstracts, err := h.GetAbstracts(ctx)
	if err != nil {
		return nil, err
	}

	for _, abstract := range abstracts {
		if abstract.ID == id {
			return &abstract, nil
		}
	}

	return nil, fmt.Errorf("abstract with ID %d not found", id)
}

// GetAbstractsByState filters abstracts by their state.
func (h *DataSourceHandler) GetAbstractsByState(ctx context.Context, state string) ([]AbstractData, error) {
	abstracts, err := h.GetAbstracts(ctx)
	if err != nil {
		return nil, err
	}

	var filtered []AbstractData
	for _, abstract := range abstracts {
		if abstract.State == state {
			filtered = append(filtered, abstract)
		}
	}

	return filtered, nil
}

// GetContributionsBySession filters contributions by session.
func (h *DataSourceHandler) GetContributionsBySession(ctx context.Context, session string) ([]ContributionData, error) {
	contributions, err := h.GetContributions(ctx)
	if err != nil {
		return nil, err
	}

	var filtered []ContributionData
	for _, contrib := range contributions {
		if contrib.Session == session {
			filtered = append(filtered, contrib)
		}
	}

	return filtered, nil
}

// GetContributionsByTrack filters contributions by track.
func (h *DataSourceHandler) GetContributionsByTrack(ctx context.Context, track string) ([]ContributionData, error) {
	contributions, err := h.GetContributions(ctx)
	if err != nil {
		return nil, err
	}

	var filtered []ContributionData
	for _, contrib := range contributions {
		if contrib.Track == track {
			filtered = append(filtered, contrib)
		}
	}

	return filtered, nil
}

// RefreshCache invalidates and refreshes a specific cache entry
func (h *DataSourceHandler) RefreshCache(ctx context.Context, key string) error {
	if h.cache == nil {
		return fmt.Errorf("cache not initialized")
	}

	// Invalidate the cache entry
	h.cache.Delete(key)

	// Immediately fetch fresh data based on key
	switch key {
	case "event_info":
		event, err := h.getInfoFromAPI(ctx)
		if err != nil {
			return fmt.Errorf("failed to refresh event info: %w", err)
		}
		h.cache.Set(key, event)
	case "abstracts":
		abstracts, err := h.getAbstractsFromAPI(ctx)
		if err != nil {
			return fmt.Errorf("failed to refresh abstracts: %w", err)
		}
		h.cache.Set(key, abstracts)
	case "contributions":
		contribs, err := h.getContributionsFromAPI(ctx)
		if err != nil {
			return fmt.Errorf("failed to refresh contributions: %w", err)
		}
		h.cache.Set(key, contribs)
	default:
		return fmt.Errorf("unknown cache key: %s", key)
	}

	log.Printf("Cache refreshed for key: %s", key)
	return nil
}

// ClearCache removes all entries from cache and deletes the cache file
func (h *DataSourceHandler) ClearCache() error {
	if h.cache == nil {
		return fmt.Errorf("cache not initialized")
	}

	// Clear in-memory cache
	h.cache.Clear()

	// Delete the cache file
	if err := h.cache.DeleteCacheFile(); err != nil {
		log.Printf("Warning: failed to delete cache file: %v", err)
		// Don't fail - the in-memory cache is cleared
	}

	log.Printf("Cache cleared and file removed")
	return nil
}

// GetCacheStats returns cache statistics
func (h *DataSourceHandler) GetCacheStats() map[string]interface{} {
	if h.cache == nil {
		return map[string]interface{}{
			"enabled": false,
		}
	}
	stats := h.cache.GetStats()
	stats["enabled"] = true
	return stats
}

// GetCacheKeys returns all available cache keys
func (h *DataSourceHandler) GetCacheKeys() []string {
	if h.cache == nil {
		return []string{}
	}
	return h.cache.Keys()
}

// IsTestMode returns true if the data source is in test mode (local files)
func (h *DataSourceHandler) IsTestMode() bool {
	return h.isTestMode
}

// GetCacheEntries returns all cache entries with metadata
func (h *DataSourceHandler) GetCacheEntries() []*CacheEntry {
	if h.cache == nil {
		return []*CacheEntry{}
	}
	return h.cache.GetAllEntriesWithMetadata()
}
