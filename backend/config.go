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

// ActiveDataSource specifies which data source to use.
type ActiveDataSource struct {
	Use string `yaml:"use"`
}

// IndicoConfig holds Indico API configuration.
type IndicoConfig struct {
	BaseURL  string `yaml:"base_url" json:"baseUrl"`
	EventID  int    `yaml:"event_id" json:"eventId"`
	APIToken string `yaml:"api_token" json:"apiToken"`
	Timeout  string `yaml:"timeout,omitempty" json:"timeout,omitempty"`
}

// TestConfig holds test/local data configuration.
type TestConfig struct {
	DataDir   string `yaml:"data_dir" json:"dataDir"`
	EventInfo string `yaml:"event_info" json:"eventInfo"`
	Abstracts string `yaml:"abstracts" json:"abstracts"`
	Contribs  string `yaml:"contribs" json:"contribs"`
}

// DataSource represents a named data source configuration.
type DataSource struct {
	Name   string        `json:"name"`
	Type   string        `json:"type"` // "indico" or "test"
	Indico *IndicoConfig `json:"indico,omitempty"`
	Test   *TestConfig   `json:"test,omitempty"`
}

// Config holds the complete configuration with multiple data sources.
type Config struct {
	ActiveDataSource ActiveDataSource          `yaml:"data-source"`
	DataSources      map[string]map[string]any `yaml:",inline"`
}

// GetDataSource retrieves a specific data source by name.
func (c *Config) GetDataSource(name string) (*DataSource, error) {
	rawData, ok := c.DataSources[name]
	if !ok {
		return nil, os.ErrNotExist
	}

	ds := &DataSource{Name: name}

	// Check explicit indico field to determine type
	isIndico := false
	if indicoFlag, ok := rawData["indico"].(bool); ok {
		isIndico = indicoFlag
	}

	if isIndico {
		// Parse as IndicoConfig
		ds.Type = "indico"
		ic := &IndicoConfig{}
		if baseURL, ok := rawData["base_url"].(string); ok {
			ic.BaseURL = baseURL
		}
		if eventID, ok := rawData["event_id"].(int); ok {
			ic.EventID = eventID
		}
		if apiToken, ok := rawData["api_token"].(string); ok {
			ic.APIToken = apiToken
		}
		if timeout, ok := rawData["timeout"].(string); ok {
			ic.Timeout = timeout
		}
		ds.Indico = ic
	} else {
		// Parse as TestConfig
		ds.Type = "test"
		tc := &TestConfig{}
		if dataDir, ok := rawData["data_dir"].(string); ok {
			tc.DataDir = dataDir
		}
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
	}

	return ds, nil
}

// GetActiveDataSource retrieves the data source in use.
func (c *Config) GetActiveDataSource() (*DataSource, error) {
	return c.GetDataSource(c.ActiveDataSource.Use)
}

// LoadConfig reads and parses a YAML config file at path.
func LoadConfig(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return LoadConfigFromBytes(b)
}

// LoadConfigFromBytes reads and parses YAML content from bytes.
func LoadConfigFromBytes(b []byte) (*Config, error) {
	var rawConfig map[string]any
	if err := yaml.Unmarshal(b, &rawConfig); err != nil {
		return nil, err
	}

	cfg := &Config{
		DataSources: make(map[string]map[string]any),
	}

	// Extract data-source section
	if dataSourceSection, ok := rawConfig["data-source"].(map[string]any); ok {
		if use, ok := dataSourceSection["use"].(string); ok {
			cfg.ActiveDataSource.Use = use
		}
		delete(rawConfig, "data-source")
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

// ConfigPathInfo holds information about the config file location.
type ConfigPathInfo struct {
	Path       string `json:"path"`
	FromEnv    bool   `json:"fromEnv"`
	EnvVarName string `json:"envVarName"`
}

// ConfigData represents the structured configuration for the UI.
type ConfigDataUI struct {
	ActiveDataSourceName string         `json:"activeDataSourceName"`
	DataSources          []DataSource   `json:"dataSources"`
	PathInfo             ConfigPathInfo `json:"pathInfo"`
}

// GetStructuredConfigUI converts a Config to structured format for the UI.
func GetStructuredConfigUI(cfg *Config, pathInfo ConfigPathInfo) *ConfigDataUI {
	configData := &ConfigDataUI{
		ActiveDataSourceName: cfg.ActiveDataSource.Use,
		DataSources:          make([]DataSource, 0, len(cfg.DataSources)),
		PathInfo:             pathInfo,
	}

	// Use GetDataSource to build each entry
	for name := range cfg.DataSources {
		ds, err := cfg.GetDataSource(name)
		if err != nil {
			// skip invalid/unknown data sources
			continue
		}
		// append a copy of the resolved DataSource
		configData.DataSources = append(configData.DataSources, *ds)
	}

	return configData
}

// BuildConfigFromStructuredUI converts structured ConfigData back to Config.
func BuildConfigFromStructuredUI(configData *ConfigDataUI) *Config {
	cfg := &Config{
		ActiveDataSource: ActiveDataSource{
			Use: configData.ActiveDataSourceName,
		},
		DataSources: make(map[string]map[string]any),
	}

	// Convert each data source
	for _, ds := range configData.DataSources {
		rawData := make(map[string]any)

		if ds.Type == "indico" && ds.Indico != nil {
			rawData["indico"] = true
			rawData["base_url"] = ds.Indico.BaseURL
			rawData["event_id"] = ds.Indico.EventID
			rawData["api_token"] = ds.Indico.APIToken
			if ds.Indico.Timeout != "" {
				rawData["timeout"] = ds.Indico.Timeout
			}
		} else if ds.Type == "test" && ds.Test != nil {
			rawData["indico"] = false
			rawData["data_dir"] = ds.Test.DataDir
			rawData["event_info"] = ds.Test.EventInfo
			rawData["abstracts"] = ds.Test.Abstracts
			rawData["contribs"] = ds.Test.Contribs
		}

		cfg.DataSources[ds.Name] = rawData
	}

	return cfg
}
