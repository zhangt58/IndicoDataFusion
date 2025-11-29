package backend

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestNewDataSourceHandler(t *testing.T) {
	// Test with Indico config
	indicoDS := &DataSource{
		Name: "indico",
		Type: "indico",
		Indico: &IndicoConfig{
			BaseURL:  "https://example.com",
			EventID:  123,
			APIToken: "token",
			Timeout:  "10s",
		},
	}

	handler, err := NewDataSourceHandler(indicoDS)
	if err != nil {
		t.Fatalf("NewDataSourceHandler(indico) failed: %v", err)
	}
	if handler.isTestMode {
		t.Fatalf("Expected API mode, got test mode")
	}
	if handler.client == nil {
		t.Fatalf("Expected client to be initialized")
	}

	// Test with Test config
	testDS := &DataSource{
		Name: "test",
		Type: "test",
		Test: &TestConfig{
			DataDir:   "./testdata",
			EventInfo: "info.json",
			Abstracts: "abstracts.json",
			Contribs:  "contribs.json",
		},
	}

	handler, err = NewDataSourceHandler(testDS)
	if err != nil {
		t.Fatalf("NewDataSourceHandler(test) failed: %v", err)
	}
	if !handler.isTestMode {
		t.Fatalf("Expected test mode, got API mode")
	}
	// dataDir is converted to absolute path, so check it's absolute and contains testdata
	if !filepath.IsAbs(handler.dataDir) {
		t.Fatalf("Expected absolute dataDir, got %q", handler.dataDir)
	}

	// Test with invalid config
	invalidDS := &DataSource{Name: "invalid"}
	_, err = NewDataSourceHandler(invalidDS)
	if err == nil {
		t.Fatalf("Expected error for invalid data source, got nil")
	}
}

func TestDataSourceHandlerGetInfo(t *testing.T) {
	// Create test config pointing to testdata directory
	testDS := &DataSource{
		Name: "test",
		Test: &TestConfig{
			DataDir:   "../testdata",
			EventInfo: "info.json",
			Abstracts: "abstracts.json",
			Contribs:  "contribs.json",
		},
	}

	handler, err := NewDataSourceHandler(testDS)
	if err != nil {
		t.Fatalf("NewDataSourceHandler failed: %v", err)
	}

	ctx := context.Background()
	event, err := handler.GetInfo(ctx)
	if err != nil {
		// Check if testdata exists
		if os.IsNotExist(err) {
			t.Skipf("Skipping test: testdata/info.json not found")
			return
		}
		t.Fatalf("GetInfo failed: %v", err)
	}

	if event.Title == "" {
		t.Fatalf("Expected non-empty title")
	}

	t.Logf("Event title: %s", event.Title)
}

func TestDataSourceHandlerGetAbstracts(t *testing.T) {
	// Create test config pointing to testdata directory
	testDS := &DataSource{
		Name: "test",
		Test: &TestConfig{
			DataDir:   "../testdata",
			EventInfo: "info.json",
			Abstracts: "abstracts.json",
			Contribs:  "contribs.json",
		},
	}

	handler, err := NewDataSourceHandler(testDS)
	if err != nil {
		t.Fatalf("NewDataSourceHandler failed: %v", err)
	}

	ctx := context.Background()
	abstracts, err := handler.GetAbstracts(ctx)
	if err != nil {
		// Check if testdata exists
		if os.IsNotExist(err) {
			t.Skipf("Skipping test: testdata/abstracts.json not found")
			return
		}
		t.Fatalf("GetAbstracts failed: %v", err)
	}

	if len(abstracts) == 0 {
		t.Logf("Warning: no abstracts found")
	}

	t.Logf("Found %d abstracts", len(abstracts))
}

func TestDataSourceHandlerGetContributions(t *testing.T) {
	// Create test config pointing to testdata directory
	testDS := &DataSource{
		Name: "test",
		Test: &TestConfig{
			DataDir:   "../testdata",
			EventInfo: "info.json",
			Abstracts: "abstracts.json",
			Contribs:  "contribs.json",
		},
	}

	handler, err := NewDataSourceHandler(testDS)
	if err != nil {
		t.Fatalf("NewDataSourceHandler failed: %v", err)
	}

	ctx := context.Background()
	contribs, err := handler.GetContributions(ctx)
	if err != nil {
		// Check if testdata exists
		if os.IsNotExist(err) {
			t.Skipf("Skipping test: testdata/contribs.json not found")
			return
		}
		t.Fatalf("GetContributions failed: %v", err)
	}

	if len(contribs) == 0 {
		t.Logf("Warning: no contributions found")
	}

	t.Logf("Found %d contributions", len(contribs))
}

func TestNewDataSourceHandlerFromConfig(t *testing.T) {
	// Create a temporary config file
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	cfg := &Config{
		ActiveDataSource: ActiveDataSource{
			Use: "test",
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

	if err := SaveConfig(configPath, cfg); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	handler, err := NewDataSourceHandlerFromConfigFile(configPath)
	if err != nil {
		t.Fatalf("NewDataSourceHandlerFromConfigFile failed: %v", err)
	}

	if !handler.isTestMode {
		t.Fatalf("Expected test mode")
	}
	if handler.config.Name != "test" {
		t.Fatalf("Expected data source name 'test', got %q", handler.config.Name)
	}
}

func TestGetAbstractsByState(t *testing.T) {
	// Create test config pointing to testdata directory
	testDS := &DataSource{
		Name: "test",
		Test: &TestConfig{
			DataDir:   "../testdata",
			EventInfo: "info.json",
			Abstracts: "abstracts.json",
			Contribs:  "contribs.json",
		},
	}

	handler, err := NewDataSourceHandler(testDS)
	if err != nil {
		t.Fatalf("NewDataSourceHandler failed: %v", err)
	}

	ctx := context.Background()
	filtered, err := handler.GetAbstractsByState(ctx, "accepted")
	if err != nil {
		// Check if testdata exists
		if os.IsNotExist(err) {
			t.Skipf("Skipping test: testdata not found")
			return
		}
		t.Fatalf("GetAbstractsByState failed: %v", err)
	}

	t.Logf("Found %d abstracts with state 'accepted'", len(filtered))
}

func TestGetContributionsBySession(t *testing.T) {
	// Create test config pointing to testdata directory
	testDS := &DataSource{
		Name: "test",
		Test: &TestConfig{
			DataDir:   "../testdata",
			EventInfo: "info.json",
			Abstracts: "abstracts.json",
			Contribs:  "contribs.json",
		},
	}

	handler, err := NewDataSourceHandler(testDS)
	if err != nil {
		t.Fatalf("NewDataSourceHandler failed: %v", err)
	}

	ctx := context.Background()
	contribs, err := handler.GetContributions(ctx)
	if err != nil {
		// Check if testdata exists
		if os.IsNotExist(err) {
			t.Skipf("Skipping test: testdata not found")
			return
		}
		t.Fatalf("GetContributions failed: %v", err)
	}

	// If there are contributions, test filtering by the first one's session
	if len(contribs) > 0 && contribs[0].Session != "" {
		session := contribs[0].Session
		filtered, err := handler.GetContributionsBySession(ctx, session)
		if err != nil {
			t.Fatalf("GetContributionsBySession failed: %v", err)
		}
		t.Logf("Found %d contributions in session %q", len(filtered), session)
		if len(filtered) == 0 {
			t.Fatalf("Expected at least one contribution in session %q", session)
		}
	} else {
		t.Logf("No contributions with sessions found, skipping filter test")
	}
}
