package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// CacheEntry represents a cached item with metadata
type CacheEntry struct {
	Data           interface{} `json:"data"`
	Timestamp      time.Time   `json:"timestamp"`
	Key            string      `json:"key"`
	ExpiresAt      time.Time   `json:"expires_at"`
	Size           int64       `json:"size"` // Approximate size in bytes
	DataSourceName string      `json:"data_source_name"`
}

// Cache provides thread-safe in-memory and disk-backed caching
type Cache struct {
	mu             sync.RWMutex
	entries        map[string]*CacheEntry
	cacheDir       string
	ttl            time.Duration
	maxSize        int64       // Maximum cache size in bytes
	currentSize    int64       // Current cache size in bytes
	saveQueue      chan string // Queue for async saves
	stopChan       chan struct{}
	dataSourceName string

	// Callbacks invoked when an entry is deleted or evicted. These are optional and run
	// asynchronously (called without holding cache locks).
	onDelete func(fullKey string)
	onEvict  func(fullKey string)
}

// CacheOptions configures cache behavior
type CacheOptions struct {
	CacheDir       string        // Directory for disk cache (default: ~/.cache/<app-name>)
	LoadOnStartup  bool          // Load cache from disk on creation
	TTL            time.Duration // Time-to-live for cache entries (0 = no expiration)
	MaxSize        int64         // Maximum cache size in bytes (0 = no limit)
	DataSourceName string        // Data source name to include in cache keys
}

// NewCache creates a new Cache instance
func NewCache(opts CacheOptions) (*Cache, error) {
	if opts.CacheDir == "" {
		// Use user home directory ~/.cache/<app-name>
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home dir: %w", err)
		}
		opts.CacheDir = filepath.Join(homeDir, ".cache", "IndicoDataFusion")
	}

	// Ensure cache directory exists
	if err := os.MkdirAll(opts.CacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Set default TTL if not specified
	if opts.TTL == 0 {
		opts.TTL = 24 * time.Hour // Default: 24 hours
	}

	// Set default max size if not specified
	if opts.MaxSize == 0 {
		opts.MaxSize = 100 * 1024 * 1024 // Default: 100 MB
	}

	cache := &Cache{
		entries:        make(map[string]*CacheEntry),
		cacheDir:       opts.CacheDir,
		ttl:            opts.TTL,
		maxSize:        opts.MaxSize,
		currentSize:    0,
		saveQueue:      make(chan string, 100), // Buffered channel for async saves
		stopChan:       make(chan struct{}),
		dataSourceName: opts.DataSourceName,
	}

	// Load existing cache from disk if requested
	if opts.LoadOnStartup {
		if err := cache.loadFromDisk(); err != nil {
			log.Printf("Warning: failed to load cache from disk: %v", err)
			// Don't fail - just start with empty cache
		}
	}

	// Start async save worker
	go cache.asyncSaveWorker()

	// Note: Expiration worker is disabled - expired entries are kept in cache
	// and only deleted manually or on refresh. Users are notified of expired data.

	return cache, nil
}

// expirationWorker periodically scans for expired entries and deletes them so onDelete callbacks fire.
func (c *Cache) expirationWorker() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			now := time.Now()
			var expired []string
			c.mu.Lock()
			for fullKey, entry := range c.entries {
				if !entry.ExpiresAt.IsZero() && now.After(entry.ExpiresAt) {
					expired = append(expired, fullKey)
				}
			}

			// Delete expired entries while holding lock
			for _, fk := range expired {
				if entry, exists := c.entries[fk]; exists {
					c.currentSize -= entry.Size
					delete(c.entries, fk)
				}
			}
			c.mu.Unlock()

			// Invoke callbacks after releasing lock
			if c.onDelete != nil && len(expired) > 0 {
				log.Printf("expirationWorker: expired %d entries", len(expired))
				for _, fk := range expired {
					go func(fullKey string) {
						defer func() {
							if r := recover(); r != nil {
								log.Printf("Recovered from panic in onDelete callback: %v", r)
							}
						}()
						c.onDelete(fullKey)
					}(fk)
				}
			}
		case <-c.stopChan:
			return
		}
	}
}

// asyncSaveWorker handles asynchronous saves to disk
func (c *Cache) asyncSaveWorker() {
	for {
		select {
		case <-c.saveQueue:
			// Debounce multiple save requests
			time.Sleep(100 * time.Millisecond)

			// Drain any additional save requests
			for len(c.saveQueue) > 0 {
				<-c.saveQueue
			}

			// Perform the save
			if err := c.SaveToDisk(); err != nil {
				log.Printf("Error during async save: %v", err)
			}
		case <-c.stopChan:
			return
		}
	}
}

// queueSave queues an asynchronous save operation
func (c *Cache) queueSave() {
	select {
	case c.saveQueue <- "save":
		// Queued successfully
	default:
		// Queue full, skip (will be saved on next operation)
	}
}

// makeCacheKey creates a cache key with data source prefix
func (c *Cache) makeCacheKey(key string) string {
	if c.dataSourceName != "" {
		return fmt.Sprintf("%s:%s", c.dataSourceName, key)
	}
	return key
}

// Get retrieves a value from cache, returning data even if expired
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	fullKey := c.makeCacheKey(key)
	entry, exists := c.entries[fullKey]
	if !exists {
		return nil, false
	}

	// Return data even if expired - caller can check expiry status separately
	return entry.Data, true
}

// GetWithMetadata retrieves a cache entry with its metadata
// Returns the entry even if expired
func (c *Cache) GetWithMetadata(key string) (*CacheEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	fullKey := c.makeCacheKey(key)
	entry, exists := c.entries[fullKey]

	return entry, exists
}

// IsExpired checks if a cache entry has expired
func (c *Cache) IsExpired(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	fullKey := c.makeCacheKey(key)
	entry, exists := c.entries[fullKey]
	if !exists {
		return false
	}

	return !entry.ExpiresAt.IsZero() && time.Now().After(entry.ExpiresAt)
}

// Set stores a value in cache
func (c *Cache) Set(key string, data interface{}) {
	c.setInternal(key, data, true)
}

// setInternal is the internal implementation of Set with option to skip disk save
func (c *Cache) setInternal(key string, data interface{}, saveToDisk bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	fullKey := c.makeCacheKey(key)

	// Estimate size of data
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Warning: failed to marshal data for size estimation: %v", err)
		return
	}
	dataSize := int64(len(jsonData))

	// Check if adding this entry would exceed max size
	if c.maxSize > 0 {
		// Remove old entry size if updating
		if oldEntry, exists := c.entries[fullKey]; exists {
			c.currentSize -= oldEntry.Size
		}

		// Evict entries if necessary
		for c.currentSize+dataSize > c.maxSize && len(c.entries) > 0 {
			c.evictOldest()
		}
	}

	// Calculate expiration time
	expiresAt := time.Time{}
	if c.ttl > 0 {
		expiresAt = time.Now().Add(c.ttl)
	}

	// Create new entry
	entry := &CacheEntry{
		Data:           data,
		Timestamp:      time.Now(),
		Key:            fullKey,
		ExpiresAt:      expiresAt,
		Size:           dataSize,
		DataSourceName: c.dataSourceName,
	}

	c.entries[fullKey] = entry
	c.currentSize += dataSize

	// Queue async save only if requested
	if saveToDisk {
		go c.queueSave()
	}
}

// evictOldest removes the oldest cache entry (must be called with lock held)
func (c *Cache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range c.entries {
		if oldestKey == "" || entry.Timestamp.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.Timestamp
		}
	}

	if oldestKey != "" {
		if entry, exists := c.entries[oldestKey]; exists {
			c.currentSize -= entry.Size
			delete(c.entries, oldestKey)
			log.Printf("Evicted cache entry: %s (age: %v)", oldestKey, time.Since(oldestTime))
			// Fire evict callback asynchronously if provided
			if c.onEvict != nil {
				go func(k string) {
					defer func() {
						if r := recover(); r != nil {
							log.Printf("Recovered from panic in onEvict callback: %v", r)
						}
					}()
					c.onEvict(k)
				}(oldestKey)
			}
		}
	}
}

// Delete removes a specific entry from cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	fullKey := c.makeCacheKey(key)
	if entry, exists := c.entries[fullKey]; exists {
		c.currentSize -= entry.Size
		delete(c.entries, fullKey)
	}
}

// Clear removes all entries from cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*CacheEntry)
	c.currentSize = 0
}

// Keys returns all cache keys (without data source prefix)
// Includes expired entries
func (c *Cache) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, len(c.entries))
	prefix := ""
	if c.dataSourceName != "" {
		prefix = c.dataSourceName + ":"
	}

	for k := range c.entries {
		// Include all entries, even expired ones

		// Strip data source prefix if present
		if prefix != "" && len(k) > len(prefix) && k[:len(prefix)] == prefix {
			keys = append(keys, k[len(prefix):])
		} else {
			keys = append(keys, k)
		}
	}
	return keys
}

// SaveToDisk writes the cache to disk as JSON
func (c *Cache) SaveToDisk() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.saveToDiskUnsafe()
}

// saveToDiskUnsafe performs the actual save without acquiring locks
// Caller must hold at least a read lock
func (c *Cache) saveToDiskUnsafe() error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic during cache save: %v", r)
		}
	}()

	cacheFile := filepath.Join(c.cacheDir, "cache.json")
	backupFile := cacheFile + ".bak"

	// Create backup of existing cache file
	if _, err := os.Stat(cacheFile); err == nil {
		if err := os.Rename(cacheFile, backupFile); err != nil {
			log.Printf("Warning: failed to create backup: %v", err)
		}
	}

	// Marshal cache data
	data, err := json.MarshalIndent(c.entries, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}

	// Write to disk
	if err := os.WriteFile(cacheFile, data, 0644); err != nil {
		// Try to restore backup
		if _, statErr := os.Stat(backupFile); statErr == nil {
			os.Rename(backupFile, cacheFile)
		}
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	// Remove backup on successful write
	os.Remove(backupFile)

	log.Printf("Cache saved to disk: %s", cacheFile)
	return nil
}

// loadFromDisk reads cache from disk
func (c *Cache) loadFromDisk() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic during cache load: %v", r)
			c.entries = make(map[string]*CacheEntry)
			c.currentSize = 0
		}
	}()

	cacheFile := filepath.Join(c.cacheDir, "cache.json")

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Not an error - just no cache yet
		}
		return fmt.Errorf("failed to read cache file: %w", err)
	}

	var entries map[string]*CacheEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		log.Printf("Warning: corrupted cache file, starting fresh: %v", err)
		return nil // Don't fail - just start with empty cache
	}

	// Load all entries including expired ones and recalculate size
	c.entries = make(map[string]*CacheEntry)
	c.currentSize = 0
	now := time.Now()
	expiredCount := 0

	for key, entry := range entries {
		// Count expired entries but keep them
		if !entry.ExpiresAt.IsZero() && now.After(entry.ExpiresAt) {
			expiredCount++
		}

		c.entries[key] = entry
		c.currentSize += entry.Size
	}

	log.Printf("Cache loaded from disk: %d entries (%d expired but kept)", len(c.entries), expiredCount)
	return nil
}

// Shutdown gracefully shuts down the cache, saving to disk
func (c *Cache) Shutdown(ctx context.Context) error {
	// Stop async save worker
	close(c.stopChan)

	// Final save to disk
	return c.SaveToDisk()
}

// GetCachePath returns the directory where cache files are stored
func (c *Cache) GetCachePath() string {
	return c.cacheDir
}

// DeleteCacheFile removes the cache file from disk
func (c *Cache) DeleteCacheFile() error {
	cacheFile := filepath.Join(c.cacheDir, "cache.json")
	backupFile := cacheFile + ".bak"

	// Remove main cache file
	if err := os.Remove(cacheFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove cache file: %w", err)
	}

	// Remove backup file if it exists
	if err := os.Remove(backupFile); err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: failed to remove backup file: %v", err)
	}

	log.Printf("Cache file deleted: %s", cacheFile)
	return nil
}

// GetStats returns cache statistics
func (c *Cache) GetStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Calculate size in MB
	sizeMB := float64(c.currentSize) / (1024 * 1024)
	maxSizeMB := float64(c.maxSize) / (1024 * 1024)

	return map[string]interface{}{
		"entries":          len(c.entries),
		"cache_dir":        c.cacheDir,
		"current_size":     c.currentSize,
		"current_size_mb":  fmt.Sprintf("%.2f MB", sizeMB),
		"max_size":         c.maxSize,
		"max_size_mb":      fmt.Sprintf("%.2f MB", maxSizeMB),
		"ttl":              c.ttl.String(),
		"data_source_name": c.dataSourceName,
	}
}

// GetAllEntriesWithMetadata returns metadata for all cache entries grouped by data source
// Includes expired entries so they can be displayed and managed in the UI
func (c *Cache) GetAllEntriesWithMetadata() map[string][]*CacheEntry {
	c.mu.RLock()
	defer c.mu.RUnlock()

	grouped := make(map[string][]*CacheEntry)

	for fullKey, entry := range c.entries {
		// Include all entries, even expired ones

		// Determine the data source name
		dataSourceName := entry.DataSourceName
		if dataSourceName == "" {
			// Fallback to extracting from key if not set
			if c.dataSourceName != "" {
				dataSourceName = c.dataSourceName
			} else {
				dataSourceName = "unknown"
			}
		}

		// Strip data source prefix from key for display
		displayKey := fullKey
		prefix := dataSourceName + ":"
		if len(fullKey) > len(prefix) && fullKey[:len(prefix)] == prefix {
			displayKey = fullKey[len(prefix):]
		}

		// Create a copy of the entry with the display key
		entryCopy := &CacheEntry{
			Data:           nil, // Don't include data in metadata response
			Timestamp:      entry.Timestamp,
			Key:            displayKey,
			ExpiresAt:      entry.ExpiresAt,
			Size:           entry.Size,
			DataSourceName: dataSourceName,
		}

		grouped[dataSourceName] = append(grouped[dataSourceName], entryCopy)
	}

	return grouped
}

// SetOnDelete registers a callback that will be called when an entry is deleted (including expired deletes).
// The callback receives the full cache key (including data source prefix if present).
func (c *Cache) SetOnDelete(cb func(fullKey string)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onDelete = cb
}

// SetOnEvict registers a callback that will be called when an entry is evicted due to size limits.
func (c *Cache) SetOnEvict(cb func(fullKey string)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onEvict = cb
}
