// fetch-abstracts-data retrieves raw abstract data from an Indico data source
// using the IndicoClient.FetchAbstractsData API and writes the response to a
// JSON file.  The output file can be fed back to the application by setting
// the data source `abstracts_file` in the YAML config or by selecting the file
// via the application's UI to operate in file-override (review) mode during
// development / testing.
//
// Usage:
//
//	fetch-abstracts-data --config config/dev.yaml --source <data-source-name> [--out abstracts-data.json]
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"IndicoDataFusion/backend/config"
	"IndicoDataFusion/backend/data"
)

func writeJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}

func main() {
	cfgPath := flag.String("config", "", "path to config yaml (required)")
	sourceName := flag.String("source", "", "data source name to use; defaults to the active data source in config")
	out := flag.String("out", "abstracts-data.json", "output JSON file path")
	flag.Parse()

	if *cfgPath == "" {
		log.Fatalf("--config is required")
	}

	cfg, err := config.LoadConfig(*cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	// Resolve target data source – fall back to the active one when --source is not set.
	name := *sourceName
	if name == "" {
		name = cfg.ActiveDataSource.Use
	}
	if name == "" {
		log.Fatalf("no data source specified and no active data source configured")
	}

	ds, err := cfg.GetDataSource(name)
	if err != nil {
		log.Fatalf("data source %q not found in config: %v", name, err)
	}

	if ds.Type != "indico" || ds.Indico == nil {
		log.Fatalf("data source %q is not an Indico source (type=%q); only Indico sources are supported", name, ds.Type)
	}

	// Initialize a DataSourceHandler which resolves API tokens / keyring and builds the Indico client.
	handler, err := data.NewDataSourceHandler(ds, cfg.Cache, cfg.APITokens)
	if err != nil {
		log.Fatalf("failed to initialize data handler for %q: %v", name, err)
	}

	ctx := context.Background()

	log.Printf("Fetching abstracts from data source %q (%s, event %d)…", name, ds.Indico.BaseURL, ds.Indico.EventID)

	rawData, err := handler.FetchAbstractsRaw(ctx)
	if err != nil {
		log.Fatalf("FetchAbstractsRaw: %v", err)
	}

	if rawData == nil {
		// Defensive: write an empty payload if nil returned
		rawData = map[string]any{"abstracts": []any{}, "questions": []any{}}
	}

	if err := writeJSON(*out, rawData); err != nil {
		log.Fatalf("write output: %v", err)
	}
	fmt.Printf("Wrote %s\n", *out)
}
