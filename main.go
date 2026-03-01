package main

import (
	"IndicoDataFusion/backend/consts"
	"context"
	"embed"
	"flag"
	"fmt"
	"os"

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
	abstractsFile := flag.String("abstracts-file", "", "path to a pre-processed abstracts JSON file; overrides the data source for all GetAbstracts calls")
	flag.Parse()

	// Determine which config path to use (flag > existing default > create from sample)
	chosenConfig := app.DetermineConfigPath(*cfgPath)
	if chosenConfig == "" {
		fmt.Printf("No config path determined; startup will require an explicit path\n")
	}

	// Store the abstracts file override so startup can apply it to the handler
	if env := os.Getenv("IDF_ABSTRACTS_FILE"); env != "" {
		fmt.Printf("Using abstracts file from env: %s\n", env)
		app.abstractsFile = env
	} else {
		app.abstractsFile = *abstractsFile
	}

	// Create application with options
	err := wails.Run(&options.App{
		Title:     consts.AppName,
		Width:     1280,
		Height:    800,
		Frameless: true,
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
			ProgramName:         consts.AppName,
		},
		Mac: &mac.Options{
			TitleBar:             &mac.TitleBar{HideTitleBar: true},
			WebviewIsTransparent: true,
			WindowIsTranslucent:  true,
		},
		Windows: &windows.Options{
			WebviewIsTransparent:              true,
			WindowIsTranslucent:               true,
			BackdropType:                      windows.Mica,
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
