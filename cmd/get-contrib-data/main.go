package main

import (
	"IndicoDataFusion/backend/data"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"IndicoDataFusion/backend/config"
)

func writeDataToFile(path string, data any) error {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal contribs: %w", err)
	}
	if err := os.WriteFile(path, b, 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}

func main() {
	cfgPath := flag.String("config", "", "path to config yaml")
	out := flag.String("out", "contribs-out.json", "Output JSON file")
	flag.Parse()

	if *cfgPath == "" {
		log.Fatalf("config path is required")
	}

	cfg, err := config.LoadConfig(*cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	dataHandler, err := data.NewDataSourceHandlerFromConfig(cfg)
	if err != nil {
		log.Fatalf("NewDataSourceHandlerFromConfig failed: %v", err)
	}

	eventData, err := dataHandler.GetContributions(context.Background())
	if err != nil {
		log.Fatalf("GetInfo failed: %v", err)
	}

	if err := writeDataToFile(*out, eventData); err != nil {
		log.Fatalf("write events failed: %v", err)
	}
	fmt.Printf("Wrote %s\n", *out)
}
