package backend

import (
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Duration wraps time.Duration to (un)marshal as a string like "5s".
type Duration time.Duration

// MarshalText implements encoding.TextMarshaler for YAML support.
func (d Duration) MarshalText() ([]byte, error) {
	return []byte(time.Duration(d).String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler for YAML support.
func (d *Duration) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*d = Duration(0)
		return nil
	}
	dur, err := time.ParseDuration(string(text))
	if err != nil {
		return err
	}
	*d = Duration(dur)
	return nil
}

// DefaultSection specifies which data source to use.
type DefaultSection struct {
	DataSource string `yaml:"data_source"`
}

// IndicoConfig holds Indico API configuration.
type IndicoConfig struct {
	BaseURL  string   `yaml:"base_url"`
	EventID  int      `yaml:"event_id"`
	APIToken string   `yaml:"api_token"`
	Timeout  Duration `yaml:"timeout,omitempty"`
}

// TestConfig holds test/local data configuration.
type TestConfig struct {
	DataDir   string `yaml:"data_dir"`
	EventInfo string `yaml:"event_info"`
	Abstracts string `yaml:"abstracts"`
	Contribs  string `yaml:"contribs"`
}

// DataSource represents a named data source configuration.
type DataSource struct {
	Name   string
	Indico *IndicoConfig
	Test   *TestConfig
}

// Config holds the complete configuration with multiple data sources.
type Config struct {
	Default     DefaultSection            `yaml:"default"`
	DataSources map[string]map[string]any `yaml:",inline"`
}

// GetDataSource retrieves a specific data source by name.
func (c *Config) GetDataSource(name string) (*DataSource, error) {
	rawData, ok := c.DataSources[name]
	if !ok {
		return nil, os.ErrNotExist
	}

	ds := &DataSource{Name: name}

	// Try to parse as IndicoConfig
	if baseURL, ok := rawData["base_url"].(string); ok {
		ic := &IndicoConfig{BaseURL: baseURL}
		if eventID, ok := rawData["event_id"].(int); ok {
			ic.EventID = eventID
		}
		if apiToken, ok := rawData["api_token"].(string); ok {
			ic.APIToken = apiToken
		}
		if timeout, ok := rawData["timeout"].(string); ok {
			var d Duration
			if err := d.UnmarshalText([]byte(timeout)); err == nil {
				ic.Timeout = d
			}
		}
		ds.Indico = ic
		return ds, nil
	}

	// Try to parse as TestConfig
	if dataDir, ok := rawData["data_dir"].(string); ok {
		tc := &TestConfig{DataDir: dataDir}
		if eventInfo, ok := rawData["event_info"].(string); ok {
			tc.EventInfo = eventInfo
		}
		if abstracts, ok := rawData["abstracts"].(string); ok {
			tc.Abstracts = abstracts
		}
		if contribs, ok := rawData["contribs"].(string); ok {
			tc.Contribs = contribs
		}
		ds.Test = tc
		return ds, nil
	}

	return nil, os.ErrInvalid
}

// GetDefaultDataSource retrieves the default data source.
func (c *Config) GetDefaultDataSource() (*DataSource, error) {
	return c.GetDataSource(c.Default.DataSource)
}

// LoadConfig reads and parses a YAML config file at path.
func LoadConfig(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var rawConfig map[string]any
	if err := yaml.Unmarshal(b, &rawConfig); err != nil {
		return nil, err
	}

	cfg := &Config{
		DataSources: make(map[string]map[string]any),
	}

	// Extract default section
	if defaultSection, ok := rawConfig["default"].(map[string]any); ok {
		if dataSource, ok := defaultSection["data_source"].(string); ok {
			cfg.Default.DataSource = dataSource
		}
		delete(rawConfig, "default")
	}

	// All remaining sections are data sources
	for name, val := range rawConfig {
		if section, ok := val.(map[string]any); ok {
			cfg.DataSources[name] = section
		}
	}

	return cfg, nil
}

// SaveConfig writes the config to path atomically.
func SaveConfig(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	f, err := os.CreateTemp(dir, "cfgtmp-*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	// ensure cleanup on error
	defer func() {
		f.Close()
		os.Remove(tmp)
	}()
	if _, err := f.Write(data); err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
