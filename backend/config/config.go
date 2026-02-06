package config

import (
	"os"
	"path/filepath"
	"strings"
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
// NOTE: APIToken (raw token value) has been removed. Use APITokenName to refer
// to an entry in the top-level API tokens list (identified by base URL + username).
type IndicoConfig struct {
	BaseURL      string `yaml:"base_url" json:"baseUrl"`
	EventID      int    `yaml:"event_id" json:"eventId"`
	APITokenName string `yaml:"api_token_name" json:"apiTokenName"`
	Timeout      string `yaml:"timeout,omitempty" json:"timeout,omitempty"`
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
	Name        string        `json:"name"`
	Type        string        `json:"type"` // "indico" or "test"
	Indico      *IndicoConfig `json:"indico,omitempty"`
	Test        *TestConfig   `json:"test,omitempty"`
	Favorite    bool          `json:"favorite,omitempty" yaml:"favorite,omitempty"`
	Description string        `json:"description,omitempty" yaml:"description,omitempty"`
	Tags        []string      `json:"tags,omitempty" yaml:"tags,omitempty"`
}

// CacheConfig holds cache configuration.
type CacheConfig struct {
	TTL      string `yaml:"ttl,omitempty" json:"ttl,omitempty"`            // Time-to-live (e.g., "24h", "1h30m")
	MaxSize  string `yaml:"max_size,omitempty" json:"maxSize,omitempty"`   // Max size (e.g., "100MB", "1GB")
	CacheDir string `yaml:"cache_dir,omitempty" json:"cacheDir,omitempty"` // Custom cache directory path
}

// APITokenEntry represents a stored API token referenced by base URL and username.
type APITokenEntry struct {
	Name     string `yaml:"name" json:"name"`
	BaseURL  string `yaml:"base_url" json:"baseUrl"`
	Username string `yaml:"username" json:"username"`
	Token    string `yaml:"token" json:"token"`
}

// Config holds the complete configuration with multiple data sources.
type Config struct {
	ActiveDataSource ActiveDataSource          `yaml:"data-source"`
	Cache            *CacheConfig              `yaml:"cache,omitempty"`
	APITokens        []APITokenEntry           `yaml:"api-tokens,omitempty"`
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
		// New: api_token_name references a token entry (username) stored in top-level api_tokens
		if apiTokenName, ok := rawData["api_token_name"].(string); ok {
			ic.APITokenName = apiTokenName
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

	// Optional common fields: favorite, description, tags
	if fav, ok := rawData["favorite"].(bool); ok {
		ds.Favorite = fav
	} else if favStr, ok := rawData["favorite"].(string); ok {
		// accept common string forms
		if strings.EqualFold(favStr, "true") {
			ds.Favorite = true
		} else {
			ds.Favorite = false
		}
	}

	if desc, ok := rawData["description"].(string); ok {
		ds.Description = desc
	}

	if tagsRaw, ok := rawData["tags"].([]any); ok {
		for _, t := range tagsRaw {
			if s, ok := t.(string); ok {
				ds.Tags = append(ds.Tags, s)
			}
		}
	} else if tagsStrSlice, ok := rawData["tags"].([]string); ok {
		// unlikely from yaml.Unmarshal into map[string]any, but handle just in case
		ds.Tags = append(ds.Tags, tagsStrSlice...)
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

	// Extract cache section
	if cacheSection, ok := rawConfig["cache"].(map[string]any); ok {
		cfg.Cache = &CacheConfig{}
		if ttl, ok := cacheSection["ttl"].(string); ok {
			cfg.Cache.TTL = ttl
		}
		if maxSize, ok := cacheSection["max_size"].(string); ok {
			cfg.Cache.MaxSize = maxSize
		}
		if cacheDir, ok := cacheSection["cache_dir"].(string); ok {
			cfg.Cache.CacheDir = cacheDir
		}
		delete(rawConfig, "cache")
	}

	// Extract api-tokens section (list)
	if apiTokensSection, ok := rawConfig["api-tokens"].([]any); ok {
		for _, item := range apiTokensSection {
			if m, ok := item.(map[string]any); ok {
				entry := APITokenEntry{}
				if n, ok := m["name"].(string); ok {
					entry.Name = n
				}
				if bu, ok := m["base_url"].(string); ok {
					entry.BaseURL = bu
				}
				if un, ok := m["username"].(string); ok {
					entry.Username = un
				}
				if tok, ok := m["token"].(string); ok {
					entry.Token = tok
				}
				cfg.APITokens = append(cfg.APITokens, entry)
			}
		}
		delete(rawConfig, "api-tokens")
	}

	// All remaining sections are data sources
	for name, val := range rawConfig {
		if section, ok := val.(map[string]any); ok {
			// NOTE: legacy `api_token` migration removed — we expect callers/UI to supply
			// named api tokens (api_token_name) and top-level `api-tokens` list.
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
		// explicitly handle returned errors to satisfy linters/static checks
		_ = f.Close()
		_ = os.Remove(tmp)
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
	Path string `json:"path"`
}

// ConfigDataUI napfrepresents the structured configuration for the UI.
type ConfigDataUI struct {
	ActiveDataSourceName string         `json:"activeDataSourceName"`
	DataSources          []DataSource   `json:"dataSources"`
	Cache                *CacheConfig   `json:"cache,omitempty"`
	PathInfo             ConfigPathInfo `json:"pathInfo"`
	// Include APITokens so the UI can present/manage named tokens
	APITokens []APITokenEntry `json:"apiTokens,omitempty"`
}

// GetStructuredConfigUI converts a Config to structured format for the UI.
func GetStructuredConfigUI(cfg *Config, pathInfo ConfigPathInfo) *ConfigDataUI {
	configData := &ConfigDataUI{
		ActiveDataSourceName: cfg.ActiveDataSource.Use,
		DataSources:          make([]DataSource, 0, len(cfg.DataSources)),
		Cache:                cfg.Cache,
		PathInfo:             pathInfo,
		APITokens:            cfg.APITokens,
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
		Cache:       configData.Cache,
		DataSources: make(map[string]map[string]any),
		APITokens:   configData.APITokens,
	}

	// Convert each data source
	for _, ds := range configData.DataSources {
		rawData := make(map[string]any)

		if ds.Type == "indico" && ds.Indico != nil {
			rawData["indico"] = true
			rawData["base_url"] = ds.Indico.BaseURL
			rawData["event_id"] = ds.Indico.EventID
			rawData["api_token_name"] = ds.Indico.APITokenName
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

		// Include optional fields so they are persisted
		rawData["favorite"] = ds.Favorite
		if ds.Description != "" {
			rawData["description"] = ds.Description
		}
		if len(ds.Tags) > 0 {
			rawData["tags"] = ds.Tags
		}

		cfg.DataSources[ds.Name] = rawData
	}

	return cfg
}

// Validate checks configuration consistency and returns an error listing problems.
func (c *Config) Validate() error {
	// Previously this function enforced that every Indico data source must specify
	// a non-empty api_token_name and that the name must exist in the top-level
	// api-tokens list. To allow the UI to manage tokens interactively (and to
	// tolerate configs where token names or the entire api-tokens section are
	// omitted), we relax that validation here: token presence will be checked at
	// runtime when initializing data sources and surfaced as non-fatal init
	// problems so the frontend can prompt the user.

	// Keep this function available for other future validations; currently there
	// are no strict errors to return here.
	return nil
}
