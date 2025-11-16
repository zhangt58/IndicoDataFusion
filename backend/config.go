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

// Config holds simple client configuration (YAML).
type Config struct {
	BaseURL  string   `yaml:"base_url,omitempty"`
	APIToken string   `yaml:"api_token,omitempty"`
	EventID  int      `yaml:"event_id,omitempty"`
	Timeout  Duration `yaml:"timeout,omitempty"`
}

// LoadConfig reads and parses a YAML config file at path.
func LoadConfig(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
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
