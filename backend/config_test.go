package backend

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
		BaseURL:  "https://example.test",
		APIToken: "secret",
		EventID:  42,
		Timeout:  Duration(7 * time.Second),
	}

	if err := SaveConfig(path, cfg); err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	loaded, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if loaded.BaseURL != cfg.BaseURL {
		t.Fatalf("BaseURL mismatch: got %q want %q", loaded.BaseURL, cfg.BaseURL)
	}
	if loaded.APIToken != cfg.APIToken {
		t.Fatalf("APIToken mismatch: got %q want %q", loaded.APIToken, cfg.APIToken)
	}
	if time.Duration(loaded.Timeout) != time.Duration(cfg.Timeout) {
		t.Fatalf("Timeout mismatch: got %v want %v", time.Duration(loaded.Timeout), time.Duration(cfg.Timeout))
	}
}
