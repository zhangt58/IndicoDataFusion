package main

import (
	"embed"
	"flag"
	"os"

	"github.com/labstack/gommon/log"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed icons/icon.ico
var icon []byte

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Load configuration path
	cfgPath := flag.String("config", "", "path to config yaml")
	flag.Parse()

	if cfgPathEnv := os.Getenv(ConfEnvName); cfgPathEnv != "" {
		log.Printf("Using config path from env: %s", cfgPathEnv)
	} else {
		// Store config path for later use in startup
		os.Setenv(ConfEnvName, *cfgPath)
	}

	// Create application with options
	err := wails.Run(&options.App{
		Title:  AppName,
		Width:  1280,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: app.startup,
		Linux: &linux.Options{
			Icon:                icon,
			WindowIsTranslucent: true,
			WebviewGpuPolicy:    linux.WebviewGpuPolicyOnDemand,
			ProgramName:         AppName,
		},
		Mac: &mac.Options{
			TitleBar: mac.TitleBarDefault(),
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
		},
		Windows: &windows.Options{
			WebviewIsTransparent:              false,
			WindowIsTranslucent:               false,
			DisableFramelessWindowDecorations: false,
		},
		Bind: []interface{}{
			app,
		},
		EnableDefaultContextMenu: false,
	})
	if err != nil {
		println("Error:", err.Error())
	}
}
