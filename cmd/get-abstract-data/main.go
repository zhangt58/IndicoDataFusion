package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"IndicoDataFusion/backend"
)

func writeDataToFile(path string, data any) error {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal abstracts: %w", err)
	}
	if err := os.WriteFile(path, b, 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}

func main() {
	cfgPath := flag.String("config", "", "path to config yaml")
	out := flag.String("out", "abstracts-out.json", "Output JSON file")
	flag.Parse()

	if *cfgPath == "" {
		log.Fatalf("config path is required")
	}

	cfg, err := backend.LoadConfig(*cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	dataHandler, err := backend.NewDataSourceHandlerFromConfig(cfg)
	if err != nil {
		log.Fatalf("NewDataSourceHandlerFromConfig failed: %v", err)
	}

	abstractData, err := dataHandler.GetAbstracts(context.Background())
	if err != nil {
		log.Fatalf("GetAbstracts failed: %v", err)
	}

	if err := writeDataToFile(*out, abstractData); err != nil {
		log.Fatalf("write abstracts failed: %v", err)
	}
	fmt.Printf("Wrote %s\n", *out)
}
