package cache

import (
	"context"
	"testing"
	"time"
)

// TestUpdateTTL verifies that UpdateTTL recalculates ExpiresAt for all entries
func TestUpdateTTL(t *testing.T) {
	tmpDir := t.TempDir()

	// Create cache with 1 hour TTL
	c, err := NewCache(CacheOptions{
		CacheDir:       tmpDir,
		LoadOnStartup:  false,
		TTL:            1 * time.Hour,
		MaxSize:        100 * 1024 * 1024,
		DataSourceName: "test",
	})
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}
	defer c.Shutdown(context.Background())

	// Add a test entry
	testData := "test data"
	c.Set("test_key", testData)

	// Get the entry and check its expiry
	entry, found := c.GetWithMetadata("test_key")
	if !found {
		t.Fatal("Entry not found after Set")
	}

	originalExpiry := entry.ExpiresAt
	originalTimestamp := entry.Timestamp

	// Verify original expiry is approximately 1 hour from now
	expectedExpiry := originalTimestamp.Add(1 * time.Hour)
	if originalExpiry.Sub(expectedExpiry).Abs() > time.Second {
		t.Errorf("Original expiry not correct. Expected ~%v, got %v", expectedExpiry, originalExpiry)
	}

	// Update TTL to 2 hours
	c.UpdateTTL(2 * time.Hour)

	// Get the entry again and check its new expiry
	entry, found = c.GetWithMetadata("test_key")
	if !found {
		t.Fatal("Entry not found after UpdateTTL")
	}

	newExpiry := entry.ExpiresAt

	// Verify new expiry is approximately 2 hours from original timestamp
	expectedNewExpiry := originalTimestamp.Add(2 * time.Hour)
	if newExpiry.Sub(expectedNewExpiry).Abs() > time.Second {
		t.Errorf("New expiry not correct. Expected ~%v, got %v", expectedNewExpiry, newExpiry)
	}

	// Verify expiry changed
	if newExpiry.Equal(originalExpiry) {
		t.Error("Expiry did not change after UpdateTTL")
	}
}

// TestUpdateTTLZero verifies that UpdateTTL(0) removes expiration
func TestUpdateTTLZero(t *testing.T) {
	tmpDir := t.TempDir()

	// Create cache with 1 hour TTL
	c, err := NewCache(CacheOptions{
		CacheDir:       tmpDir,
		LoadOnStartup:  false,
		TTL:            1 * time.Hour,
		MaxSize:        100 * 1024 * 1024,
		DataSourceName: "test",
	})
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}
	defer c.Shutdown(context.Background())

	// Add a test entry
	testData := "test data"
	c.Set("test_key", testData)

	// Get the entry and verify it has expiry
	entry, found := c.GetWithMetadata("test_key")
	if !found {
		t.Fatal("Entry not found after Set")
	}

	if entry.ExpiresAt.IsZero() {
		t.Error("Entry should have expiry initially")
	}

	// Update TTL to 0 (no expiration)
	c.UpdateTTL(0)

	// Get the entry again and check its new expiry
	entry, found = c.GetWithMetadata("test_key")
	if !found {
		t.Fatal("Entry not found after UpdateTTL")
	}

	if !entry.ExpiresAt.IsZero() {
		t.Errorf("Entry should have no expiry after UpdateTTL(0), got %v", entry.ExpiresAt)
	}
}

// TestLoadFromDiskRecalculatesExpiry verifies that loading from disk recalculates ExpiresAt
func TestLoadFromDiskRecalculatesExpiry(t *testing.T) {
	tmpDir := t.TempDir()

	// Create cache with 1 hour TTL and add an entry
	c1, err := NewCache(CacheOptions{
		CacheDir:       tmpDir,
		LoadOnStartup:  false,
		TTL:            1 * time.Hour,
		MaxSize:        100 * 1024 * 1024,
		DataSourceName: "test",
	})
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	// Add a test entry
	testData := "test data"
	c1.Set("test_key", testData)

	// Get the original timestamp
	entry1, _ := c1.GetWithMetadata("test_key")
	originalTimestamp := entry1.Timestamp
	originalExpiry := entry1.ExpiresAt

	// Save to disk and shut down
	if err := c1.SaveToDisk(); err != nil {
		t.Fatalf("Failed to save cache: %v", err)
	}
	c1.Shutdown(context.Background())

	// Create a new cache with 2 hour TTL, loading from disk
	c2, err := NewCache(CacheOptions{
		CacheDir:       tmpDir,
		LoadOnStartup:  true,
		TTL:            2 * time.Hour, // Different TTL!
		MaxSize:        100 * 1024 * 1024,
		DataSourceName: "test",
	})
	if err != nil {
		t.Fatalf("Failed to create cache with LoadOnStartup: %v", err)
	}
	defer c2.Shutdown(context.Background())

	// Get the entry and check its expiry was recalculated
	entry2, found := c2.GetWithMetadata("test_key")
	if !found {
		t.Fatal("Entry not found after loading from disk")
	}

	newExpiry := entry2.ExpiresAt

	// Verify the timestamp stayed the same
	if !entry2.Timestamp.Equal(originalTimestamp) {
		t.Errorf("Timestamp should not change. Expected %v, got %v", originalTimestamp, entry2.Timestamp)
	}

	// Verify new expiry is approximately 2 hours from original timestamp (not 1 hour)
	expectedNewExpiry := originalTimestamp.Add(2 * time.Hour)
	if newExpiry.Sub(expectedNewExpiry).Abs() > time.Second {
		t.Errorf("New expiry not recalculated correctly. Expected ~%v, got %v", expectedNewExpiry, newExpiry)
	}

	// Verify expiry is different from the original
	if newExpiry.Equal(originalExpiry) {
		t.Error("Expiry should be recalculated when loading with different TTL")
	}
}
