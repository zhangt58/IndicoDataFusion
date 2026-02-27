package cache

import (
	"context"
	"testing"
	"time"
)

// TestStopExpiryWorker tests that the expiry worker can be stopped cleanly
func TestStopExpiryWorker(t *testing.T) {
	dir := t.TempDir()

	// Create a cache with a data source name to trigger expiry worker
	cache, err := NewCache(CacheOptions{
		CacheDir:       dir,
		TTL:            1 * time.Second,
		MaxSize:        1024 * 1024,
		LoadOnStartup:  false,
		DataSourceName: "test-source",
	})
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Add an entry
	cache.Set("test-key", "test-value")

	// Give the expiry worker a moment to start
	time.Sleep(100 * time.Millisecond)

	// Stop the expiry worker
	cache.StopExpiryWorker()

	// Verify we can call it again without panic
	cache.StopExpiryWorker()

	// Shutdown the cache
	if err := cache.Shutdown(context.Background()); err != nil {
		t.Fatalf("Failed to shutdown cache: %v", err)
	}
}

// TestMultipleCachesExpiryWorkers tests that multiple caches can have their own expiry workers
// and they can be stopped independently
func TestMultipleCachesExpiryWorkers(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	// Create first cache
	cache1, err := NewCache(CacheOptions{
		CacheDir:       dir1,
		TTL:            1 * time.Second,
		MaxSize:        1024 * 1024,
		LoadOnStartup:  false,
		DataSourceName: "source1",
	})
	if err != nil {
		t.Fatalf("Failed to create cache1: %v", err)
	}

	// Create second cache
	cache2, err := NewCache(CacheOptions{
		CacheDir:       dir2,
		TTL:            1 * time.Second,
		MaxSize:        1024 * 1024,
		LoadOnStartup:  false,
		DataSourceName: "source2",
	})
	if err != nil {
		t.Fatalf("Failed to create cache2: %v", err)
	}

	// Add entries
	cache1.Set("key1", "value1")
	cache2.Set("key2", "value2")

	// Give workers time to start
	time.Sleep(100 * time.Millisecond)

	// Stop first cache's worker
	cache1.StopExpiryWorker()

	// Second cache should still be running - verify by adding more data
	cache2.Set("key3", "value3")

	// Stop second cache's worker
	cache2.StopExpiryWorker()

	// Shutdown both
	if err := cache1.Shutdown(context.Background()); err != nil {
		t.Fatalf("Failed to shutdown cache1: %v", err)
	}
	if err := cache2.Shutdown(context.Background()); err != nil {
		t.Fatalf("Failed to shutdown cache2: %v", err)
	}
}

// TestExpiryWorkerNotStartedForGlobalCache tests that expiry worker is not started for global caches
func TestExpiryWorkerNotStartedForGlobalCache(t *testing.T) {
	dir := t.TempDir()

	// Create a cache without a data source name (global cache)
	cache, err := NewCache(CacheOptions{
		CacheDir:      dir,
		TTL:           1 * time.Second,
		MaxSize:       1024 * 1024,
		LoadOnStartup: false,
		// DataSourceName is empty - this is a global cache
	})
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Add an entry
	cache.Set("test-key", "test-value")

	// StopExpiryWorker should be safe to call even if worker wasn't started
	cache.StopExpiryWorker()

	// Shutdown should work fine
	if err := cache.Shutdown(context.Background()); err != nil {
		t.Fatalf("Failed to shutdown cache: %v", err)
	}
}
