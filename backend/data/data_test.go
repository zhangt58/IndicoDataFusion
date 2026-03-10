package data

import (
	"IndicoDataFusion/backend/config"
	"testing"
)

func TestNewDataSourceHandler(t *testing.T) {
	// Test with Indico config
	indicoDS := &config.DataSource{
		Name: "indico",
		Type: "indico",
		Indico: &config.IndicoConfig{
			BaseURL:      "https://example.com",
			EventID:      123,
			APITokenName: "token",
			Timeout:      "10s",
		},
	}

	// Provide api token entries so the token name can be resolved
	tokens := []config.APITokenEntry{{Name: "token", BaseURL: "https://example.com", Token: "token"}}

	handler, err := NewDataSourceHandler(indicoDS, nil, tokens)
	if err != nil {
		t.Fatalf("NewDataSourceHandler(indico) failed: %v", err)
	}
	if handler.client == nil {
		t.Fatalf("Expected client to be initialized")
	}

	// Test with invalid config (no Indico block)
	invalidDS := &config.DataSource{Name: "invalid"}
	_, err = NewDataSourceHandler(invalidDS, nil, nil)
	if err == nil {
		t.Fatalf("Expected error for invalid data source, got nil")
	}
}

func TestNewDataSourceHandlerFromConfig(t *testing.T) {
	// Create a temporary config file
	dir := t.TempDir()
	configPath := dir + "/config.yaml"

	cfg := &config.Config{
		ActiveDataSource: config.ActiveDataSource{
			Use: "indico",
		},
		APITokens: []config.APITokenEntry{
			{Name: "bot", BaseURL: "https://example.test", Token: "token"},
		},
		DataSources: map[string]map[string]any{
			"indico": {
				"indico":         true,
				"base_url":       "https://example.test",
				"api_token_name": "bot",
				"event_id":       42,
			},
		},
	}

	if err := config.SaveConfig(configPath, cfg); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	handler, err := NewDataSourceHandlerFromConfigFile(configPath)
	if err != nil {
		t.Fatalf("NewDataSourceHandlerFromConfigFile failed: %v", err)
	}

	if handler.config.Name != "indico" {
		t.Fatalf("Expected data source name 'indico', got %q", handler.config.Name)
	}
}
