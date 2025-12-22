//go:build !windows

package utils

import (
	"os"
	"path/filepath"

	"IndicoDataFusion/backend/consts"
)

func GetWindowsDefaultPaths() string {
	// Fallback for non-Windows platforms, though this function should not be called
	if userdata := os.Getenv("USERPROFILE"); userdata != "" {
		return filepath.Join(userdata, consts.AppName, "config.yml")
	}
	return ""
}
