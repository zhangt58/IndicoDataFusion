package utils

import (
	"log"
	"os"
	"path/filepath"
	goruntime "runtime"

	"IndicoDataFusion/backend/consts"
)

func GetDefaultConfigPath() string {
	switch goruntime.GOOS {
	case "windows":
		return GetWindowsDefaultPaths()
	case "darwin":
		// macOS (Library/Application Support)
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			return filepath.Join(home, "Library", "Application Support", consts.AppName, "config.yml")
		}
	default: // linux and others
		if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
			return filepath.Join(xdg, ".config", consts.AppName, "config.yml")
		} else if home, err := os.UserHomeDir(); err == nil && home != "" {
			return filepath.Join(home, ".config", consts.AppName, "config.yml")
		}
	}
	return ""
}

// DetermineConfigPath encapsulates the logic to choose or create a configuration file path.
// Priority: explicit flagPath > existing default paths (GetDefaultConfigPath) > attempt to create default from config/sample.yml or a placeholder.
// Returns the chosen path (or empty string if none could be created).
func DetermineConfigPath(flagPath string, embeddedSample []byte) string {
	// 1) If environment variable explicitly specifies a config path, use it.
	// Note: intentionally accept the value as-is (caller decides if it's valid).
	if env := os.Getenv(consts.ConfEnvName); env != "" {
		log.Printf("Using config path from env: %s", env)
		return env
	}

	// 2) If explicit flag provided, use it
	if flagPath != "" {
		return flagPath
	}

	// 3) Look for an existing default config
	p := GetDefaultConfigPath()
	if p != "" {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	// 4) No existing default – attempt to create one
	target := p
	if target == "" {
		if home, err := os.UserHomeDir(); err == nil {
			target = filepath.Join(home, ".config", consts.AppName, "config.yml")
		} else {
			target = "config.yml"
		}
	}

	// Ensure parent dir exists
	if err := os.MkdirAll(filepath.Dir(target), 0500); err != nil {
		log.Printf("Warning: failed to create config dir: %v", err)
	}

	// Prefer writing from embedded sample if available
	if len(embeddedSample) > 0 {
		if err := os.WriteFile(target, embeddedSample, 0400); err != nil {
			log.Printf("Failed to write embedded default config to %s: %v", target, err)
		} else {
			log.Printf("Created default config at %s from embedded sample", target)
			return target
		}
	} else {
		// Fallback: if for some reason embedding is not present, try file system sample
		samplePath := "config/sample.yml"
		if b, err := os.ReadFile(samplePath); err == nil {
			if werr := os.WriteFile(target, b, 0400); werr != nil {
				log.Printf("Failed to write default config to %s: %v", target, werr)
			} else {
				log.Printf("Created default config at %s from sample file", target)
				return target
			}
		} else {
			log.Printf("Sample file not found at %s: %v", samplePath, err)
		}
	}

	// If sample not available or write failed, write a placeholder
	placeholder := []byte("# Indico Data Fusion default configuration\n")
	if perr := os.WriteFile(target, placeholder, 0400); perr == nil {
		log.Printf("Created placeholder config at %s", target)
		return target
	} else {
		log.Printf("Failed to create placeholder config at %s: %v", target, perr)
	}

	// If creation failed, return empty string
	return ""
}
