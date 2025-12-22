package cache

import (
	"IndicoDataFusion/backend/config"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestCacheConfigWithCustomDirectory tests cache initialization with custom directory
func TestCacheConfigWithCustomDirectory(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	customCacheDir := filepath.Join(tmpDir, "custom_cache")

	// Create a cache config with custom directory
	cacheConfig := &config.CacheConfig{
		TTL:      "1h",
		MaxSize:  "50MB",
		CacheDir: customCacheDir,
	}

	// Create a Cache directly using equivalent options
	opts := CacheOptions{
		CacheDir:       cacheConfig.CacheDir,
		LoadOnStartup:  true,
		TTL:            time.Hour,
		MaxSize:        50 * 1024 * 1024,
		DataSourceName: "test-custom-cache",
	}

	c, err := NewCache(opts)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}
	defer c.Shutdown(context.Background())

	// Verify cache directory was created
	if _, err := os.Stat(customCacheDir); os.IsNotExist(err) {
		t.Errorf("Custom cache directory was not created: %s", customCacheDir)
	}

	// Verify cache stats reflect custom configuration
	stats := c.GetStats()
	if stats["cache_dir"] != customCacheDir {
		t.Errorf("Expected cache_dir %s, got %s", customCacheDir, stats["cache_dir"])
	}
	if stats["ttl"] != "1h0m0s" {
		t.Errorf("Expected TTL 1h0m0s, got %s", stats["ttl"])
	}
}

// TestCacheConfigDefaults tests cache initialization with default values
func TestCacheConfigDefaults(t *testing.T) {
	tmpDir := t.TempDir()

	// Create cache with defaults but point cache dir to temp to avoid touching user cache
	opts := CacheOptions{
		CacheDir:      tmpDir,
		LoadOnStartup: true,
		// TTL and MaxSize left zero to trigger defaults inside NewCache
	}

	c, err := NewCache(opts)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}
	defer c.Shutdown(context.Background())

	// Verify cache stats use defaults
	stats := c.GetStats()
	if stats["ttl"] != "24h0m0s" {
		t.Errorf("Expected default TTL 24h0m0s, got %s", stats["ttl"])
	}
	if stats["max_size_mb"] != "100.00 MB" {
		t.Errorf("Expected default max_size 100.00 MB, got %s", stats["max_size_mb"])
	}
}

// TestGetCacheEntries tests retrieving cache entries with metadata
func TestGetCacheEntries(t *testing.T) {
	// Use a temporary cache directory to avoid loading old cache
	tmpDir := t.TempDir()

	// Create cache directly with known data source name
	opts := CacheOptions{
		CacheDir:       tmpDir,
		TTL:            24 * time.Hour,
		MaxSize:        100 * 1024 * 1024,
		DataSourceName: "test-entries",
	}

	c, err := NewCache(opts)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}
	defer c.Shutdown(context.Background())

	// Populate cache for testing
	c.Set("event_info", map[string]string{"test": "data1"})
	c.Set("abstracts", map[string]string{"test": "data2"})
	c.Set("contributions", map[string]string{"test": "data3"})

	// Get cache entries
	entriesMap := c.GetAllEntriesWithMetadata()

	var entries []*CacheEntry
	for _, group := range entriesMap {
		entries = append(entries, group...)
	}

	// Should have 3 entries
	if len(entries) != 3 {
		t.Errorf("Expected 3 cache entries, got %d", len(entries))
	}

	// Verify each entry has required fields
	foundKeys := make(map[string]bool)
	for _, entry := range entries {
		if entry.Key == "" {
			t.Error("Cache entry has empty key")
		}
		foundKeys[entry.Key] = true

		if entry.Timestamp.IsZero() {
			t.Errorf("Cache entry %s has zero timestamp", entry.Key)
		}
		if entry.Size <= 0 {
			t.Errorf("Cache entry %s has invalid size %d", entry.Key, entry.Size)
		}

		// Verify timestamp is recent (within last minute)
		if time.Since(entry.Timestamp) > time.Minute {
			t.Errorf("Cache entry %s has old timestamp: %v", entry.Key, entry.Timestamp)
		}
	}

	// Verify expected keys are present
	expectedKeys := []string{"event_info", "abstracts", "contributions"}
	for _, key := range expectedKeys {
		if !foundKeys[key] {
			t.Errorf("Expected cache entry %s not found", key)
		}
	}
}

// TestConfigDataUIWithCache tests that ConfigDataUI includes cache configuration
func TestConfigDataUIWithCache(t *testing.T) {
	cfg := &config.Config{
		ActiveDataSource: config.ActiveDataSource{Use: "test"},
		Cache: &config.CacheConfig{
			TTL:      "12h",
			MaxSize:  "200MB",
			CacheDir: "/custom/path",
		},
		DataSources: map[string]map[string]any{
			"test": {
				"indico":     false,
				"data_dir":   "./testdata",
				"event_info": "info.json",
				"abstracts":  "abstracts.json",
				"contribs":   "contribs.json",
			},
		},
	}

	pathInfo := config.ConfigPathInfo{
		Path: "/test/config.yaml",
	}

	// Convert to UI format
	uiConfig := config.GetStructuredConfigUI(cfg, pathInfo)

	// Verify cache config is included
	if uiConfig.Cache == nil {
		t.Fatal("Cache config is nil in UI data")
	}

	if uiConfig.Cache.TTL != "12h" {
		t.Errorf("Expected TTL 12h, got %s", uiConfig.Cache.TTL)
	}
	if uiConfig.Cache.MaxSize != "200MB" {
		t.Errorf("Expected MaxSize 200MB, got %s", uiConfig.Cache.MaxSize)
	}
	if uiConfig.Cache.CacheDir != "/custom/path" {
		t.Errorf("Expected CacheDir /custom/path, got %s", uiConfig.Cache.CacheDir)
	}

	// Convert back to Config
	rebuiltConfig := config.BuildConfigFromStructuredUI(uiConfig)

	// Verify cache config is preserved
	if rebuiltConfig.Cache == nil {
		t.Fatal("Cache config is nil after rebuild")
	}
	if rebuiltConfig.Cache.TTL != cfg.Cache.TTL {
		t.Error("Cache TTL not preserved")
	}
	if rebuiltConfig.Cache.MaxSize != cfg.Cache.MaxSize {
		t.Error("Cache MaxSize not preserved")
	}
	if rebuiltConfig.Cache.CacheDir != cfg.Cache.CacheDir {
		t.Error("Cache CacheDir not preserved")
	}
}
