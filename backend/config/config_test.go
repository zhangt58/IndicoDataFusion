package config

import (
	"path/filepath"
	"testing"
	"time"
)

func TestDurationTextMarshaling(t *testing.T) {
	d := Duration(5 * time.Second)
	text, err := d.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText error: %v", err)
	}
	if string(text) != "5s" {
		t.Fatalf("expected '5s', got %q", string(text))
	}

	var d2 Duration
	if err := d2.UnmarshalText(text); err != nil {
		t.Fatalf("UnmarshalText error: %v", err)
	}
	if time.Duration(d2) != 5*time.Second {
		t.Fatalf("expected 5s, got %v", time.Duration(d2))
	}

	// empty text -> zero duration
	var d3 Duration
	if err := d3.UnmarshalText([]byte("")); err != nil {
		t.Fatalf("UnmarshalText(empty) error: %v", err)
	}
	if time.Duration(d3) != 0 {
		t.Fatalf("expected zero duration, got %v", time.Duration(d3))
	}
}

func TestSaveLoadConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg := &Config{
		ActiveDataSource: ActiveDataSource{
			Use: "test",
		},
		APITokens: []APITokenEntry{
			{
				Name:     "bot",
				BaseURL:  "https://example.test",
				Username: "bot",
				Token:    "secret",
			},
		},
		DataSources: map[string]map[string]any{
			"indico": {
				"indico":         true,
				"base_url":       "https://example.test",
				"api_token_name": "bot",
				"event_id":       42,
				"timeout":        "7s",
			},
			"test": {
				"indico":     false,
				"data_dir":   "./testdata",
				"event_info": "info.json",
				"abstracts":  "abstracts.json",
				"contribs":   "contribs.json",
			},
		},
	}

	if err := SaveConfig(path, cfg); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	loaded, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if loaded.ActiveDataSource.Use != cfg.ActiveDataSource.Use {
		t.Fatalf("DataSourceSection.Use mismatch: got %q want %q", loaded.ActiveDataSource.Use, cfg.ActiveDataSource.Use)
	}

	// Test GetDataSource for Indico
	indicoDS, err := loaded.GetDataSource("indico")
	if err != nil {
		t.Fatalf("GetDataSource(indico) failed: %v", err)
	}
	if indicoDS.Type != "indico" {
		t.Fatalf("Indico Type mismatch: got %q want %q", indicoDS.Type, "indico")
	}
	if indicoDS.Indico == nil {
		t.Fatalf("expected Indico config, got nil")
	}
	if indicoDS.Indico.BaseURL != "https://example.test" {
		t.Fatalf("Indico BaseURL mismatch: got %q want %q", indicoDS.Indico.BaseURL, "https://example.test")
	}
	if indicoDS.Indico.EventID != 42 {
		t.Fatalf("Indico EventID mismatch: got %d want %d", indicoDS.Indico.EventID, 42)
	}
	if indicoDS.Indico.Timeout != "7s" {
		t.Fatalf("Indico Timeout mismatch: got %q want %q", indicoDS.Indico.Timeout, "7s")
	}
	if indicoDS.Indico.APITokenName != "bot" {
		t.Fatalf("Indico APITokenName mismatch: got %q want %q", indicoDS.Indico.APITokenName, "bot")
	}

	// Test GetDataSource for Test
	testDS, err := loaded.GetDataSource("test")
	if err != nil {
		t.Fatalf("GetDataSource(test) failed: %v", err)
	}
	if testDS.Type != "test" {
		t.Fatalf("Test Type mismatch: got %q want %q", testDS.Type, "test")
	}
	if testDS.Test == nil {
		t.Fatalf("expected Test config, got nil")
	}
	if testDS.Test.DataDir != "./testdata" {
		t.Fatalf("Test DataDir mismatch: got %q want %q", testDS.Test.DataDir, "./testdata")
	}
	if testDS.Test.EventInfo != "info.json" {
		t.Fatalf("Test EventInfo mismatch: got %q want %q", testDS.Test.EventInfo, "info.json")
	}

	// Test GetDefaultDataSource
	defaultDS, err := loaded.GetActiveDataSource()
	if err != nil {
		t.Fatalf("GetDefaultDataSource failed: %v", err)
	}
	if defaultDS.Name != "test" {
		t.Fatalf("Default data source name mismatch: got %q want %q", defaultDS.Name, "test")
	}
	if defaultDS.Test == nil {
		t.Fatalf("expected Test config for default data source, got nil")
	}
}

func TestLoadRealConfig(t *testing.T) {
	// Test loading the actual config.yaml file
	cfg, err := LoadConfig("config.yaml")
	if err != nil {
		t.Skipf("Skipping real config test (config.yaml not found): %v", err)
		return
	}

	if cfg.ActiveDataSource.Use == "" {
		t.Fatalf("DataSourceSection.Use is empty")
	}

	t.Logf("Default data source: %s", cfg.ActiveDataSource.Use)

	// Try to get the default data source
	ds, err := cfg.GetActiveDataSource()
	if err != nil {
		t.Fatalf("GetDefaultDataSource failed: %v", err)
	}

	t.Logf("Successfully loaded data source: %s", ds.Name)

	// Verify data source is properly parsed
	if ds.Indico == nil && ds.Test == nil {
		t.Fatalf("Data source has neither Indico nor Test configuration")
	}
}
