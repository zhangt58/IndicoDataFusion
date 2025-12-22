//go:build windows

package utils

import (
	"log"
	"os"
	"path/filepath"

	"IndicoDataFusion/backend/consts"

	"golang.org/x/sys/windows"
)

func GetWindowsDefaultPaths() string {
	docs, err := windows.KnownFolderPath(windows.FOLDERID_Documents, 0)
	if err != nil {
		log.Printf("Failed to get Documents folder: %v, falling back to USERPROFILE", err)
		if userdata := os.Getenv("USERPROFILE"); userdata != "" {
			return filepath.Join(userdata, consts.AppName, "config.yml")
		}
	} else {
		appDir := filepath.Join(docs, consts.AppName)
		return filepath.Join(appDir, "config.yml")
	}
	return ""
}
