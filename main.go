package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"runtime"

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

	// Load configuration path (optional)
	cfgPath := flag.String("config", "", "path to config yaml")
	flag.Parse()

	// Determine which config path to use (flag > existing default > create from sample)
	chosenConfig := app.DetermineConfigPath(*cfgPath)
	if chosenConfig == "" {
		fmt.Printf("No config path determined; startup will require an explicit path\n")
	}

	isMacOS := runtime.GOOS == "darwin"

	// Create application with options
	err := wails.Run(&options.App{
		Title:     AppName,
		Width:     1280,
		Height:    800,
		Frameless: !isMacOS,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		// startup called with explicit config path (no env)
		OnStartup:  func(ctx context.Context) { app.startup(ctx, chosenConfig) },
		OnShutdown: app.shutdown,
		Linux: &linux.Options{
			Icon:                icon,
			WindowIsTranslucent: true,
			WebviewGpuPolicy:    linux.WebviewGpuPolicyOnDemand,
			ProgramName:         AppName,
		},
		Mac: &mac.Options{
			TitleBar:             mac.TitleBarDefault(),
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
