package main

import (
	"IndicoDataFusion/backend"
	"embed"
	"flag"
	"os"

	"github.com/labstack/gommon/log"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	//"IndicoDataFusion/backend"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Load configuration
	cfgPath := flag.String("config", "", "path to config yaml")
	flag.Parse()

	if cfgPathEnv := os.Getenv("CONFIG_PATH"); cfgPathEnv != "" {
		log.Printf("Using config path: %s", cfgPathEnv)
		*cfgPath = cfgPathEnv
	} else if *cfgPath == "" {
		log.Errorf("Config path must be provided via -config flag")
		return
	}
	cfg, err := backend.LoadConfig(*cfgPath)
	if err != nil {
		log.Errorf("Failed to load config: %v", err)
		return
	}

	indicoClient := backend.NewIndicoClient(cfg.BaseURL, cfg.EventID, cfg.APIToken)

	// Create application with options
	err = wails.Run(&options.App{
		Title:  "IndicoDataFusion",
		Width:  1280,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: app.startup,
		Bind: []interface{}{
			app,
			indicoClient,
		},
		EnableDefaultContextMenu: false,
	})
	if err != nil {
		println("Error:", err.Error())
	}
}
