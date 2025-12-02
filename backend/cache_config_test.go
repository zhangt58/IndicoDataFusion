package backend

import (
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
	cacheConfig := &CacheConfig{
		TTL:      "1h",
		MaxSize:  "50MB",
		CacheDir: customCacheDir,
	}

	// Create a test data source
	ds := &DataSource{
		Name: "test-custom-cache",
		Type: "test",
		Test: &TestConfig{
			DataDir:   "../testdata",
			EventInfo: "info.json",
			Abstracts: "abstracts.json",
			Contribs:  "contribs.json",
		},
	}

	// Create handler with custom cache config
	handler, err := NewDataSourceHandlerWithCache(ds, cacheConfig)
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}
	defer handler.Shutdown(nil)

	// Verify cache directory was created
	if _, err := os.Stat(customCacheDir); os.IsNotExist(err) {
		t.Errorf("Custom cache directory was not created: %s", customCacheDir)
	}

	// Verify cache stats reflect custom configuration
	stats := handler.GetCacheStats()
	if stats["cache_dir"] != customCacheDir {
		t.Errorf("Expected cache_dir %s, got %s", customCacheDir, stats["cache_dir"])
	}
	if stats["ttl"] != "1h0m0s" {
		t.Errorf("Expected TTL 1h0m0s, got %s", stats["ttl"])
	}
}

// TestCacheConfigDefaults tests cache initialization with default values
func TestCacheConfigDefaults(t *testing.T) {
	ds := &DataSource{
		Name: "test-defaults",
		Type: "test",
		Test: &TestConfig{
			DataDir:   "../testdata",
			EventInfo: "info.json",
			Abstracts: "abstracts.json",
			Contribs:  "contribs.json",
		},
	}

	// Create handler with nil cache config (should use defaults)
	handler, err := NewDataSourceHandlerWithCache(ds, nil)
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}
	defer handler.Shutdown(nil)

	// Verify cache stats use defaults
	stats := handler.GetCacheStats()
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

	ds := &DataSource{
		Name: "test-entries",
		Type: "test",
		Test: &TestConfig{
			DataDir:   "../testdata",
			EventInfo: "info.json",
			Abstracts: "abstracts.json",
			Contribs:  "contribs.json",
		},
	}

	// Create handler with custom cache directory
	cacheConfig := &CacheConfig{
		CacheDir: tmpDir,
		TTL:      "24h",
		MaxSize:  "100MB",
	}

	handler, err := NewDataSourceHandlerWithCache(ds, cacheConfig)
	if err != nil {
		t.Fatalf("Failed to create handler: %v", err)
	}
	defer handler.Shutdown(nil)

	// Manually populate cache for testing (since test mode doesn't use cache)
	if handler.cache != nil {
		handler.cache.Set("event_info", map[string]string{"test": "data1"})
		handler.cache.Set("abstracts", map[string]string{"test": "data2"})
		handler.cache.Set("contributions", map[string]string{"test": "data3"})
	}

	// Get cache entries
	// Prefer using the underlying cache's grouped metadata so the test does not depend
	// on DataSourceHandler.GetCacheEntries() signature. Flatten to a slice for assertions.
	var entriesMap map[string][]*CacheEntry
	if handler.cache != nil {
		entriesMap = handler.cache.GetAllEntriesWithMetadata()
	} else {
		entriesMap = make(map[string][]*CacheEntry)
	}

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
	cfg := &Config{
		ActiveDataSource: ActiveDataSource{Use: "test"},
		Cache: &CacheConfig{
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

	pathInfo := ConfigPathInfo{
		Path: "/test/config.yaml",
	}

	// Convert to UI format
	uiConfig := GetStructuredConfigUI(cfg, pathInfo)

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
	rebuiltConfig := BuildConfigFromStructuredUI(uiConfig)

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
