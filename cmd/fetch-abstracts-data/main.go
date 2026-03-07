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
	"net/http"
	"os"
	"time"

	"IndicoDataFusion/backend/config"
	"IndicoDataFusion/backend/indico"
	"IndicoDataFusion/backend/utils"
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

	// Resolve the API token.
	apiToken := ""
	if ds.Indico.APITokenName == "" {
		log.Fatalf("data source %q has no api_token_name configured", name)
	}
	var matched *config.APITokenEntry
	for i := range cfg.APITokens {
		if cfg.APITokens[i].Name == ds.Indico.APITokenName {
			matched = &cfg.APITokens[i]
			break
		}
	}
	if matched == nil {
		log.Fatalf("api token %q not found in config api-tokens for data source %q", ds.Indico.APITokenName, name)
	}
	apiToken = matched.Token
	if apiToken == "" {
		// Fall back to keyring.
		secret, err := utils.GetAPITokenSecret(matched.Name)
		if err != nil {
			log.Fatalf("api token %q has no token in config and keyring lookup failed: %v", matched.Name, err)
		}
		apiToken = secret
	}

	// Build the Indico client.
	client := indico.NewIndicoClient(ds.Indico.BaseURL, ds.Indico.EventID, apiToken)
	if ds.Indico.Timeout != "" {
		if d, err := time.ParseDuration(ds.Indico.Timeout); err == nil {
			client.Timeout = d
			client.Client = &http.Client{Timeout: d}
		} else {
			log.Printf("warning: invalid timeout %q, using default: %v", ds.Indico.Timeout, err)
		}
	}

	ctx := context.Background()

	// Step 1 – fetch the abstract list page to get IDs and cache the CSRF token.
	log.Printf("Fetching abstract IDs from data source %q (%s, event %d)…",
		name, ds.Indico.BaseURL, ds.Indico.EventID)
	ids, err := client.GetAbstractIDsAndCSRFFromList(ctx)
	if err != nil {
		log.Fatalf("GetAbstractIDsAndCSRFFromList: %v", err)
	}
	log.Printf("Found %d abstract(s)", len(ids))

	if len(ids) == 0 {
		log.Printf("No abstracts found; writing empty result to %s", *out)
		if err := writeJSON(*out, map[string]any{"abstracts": []any{}, "questions": []any{}}); err != nil {
			log.Fatalf("write output: %v", err)
		}
		fmt.Printf("Wrote %s\n", *out)
		return
	}

	// Step 2 – fetch the full abstracts payload.
	log.Printf("Fetching abstracts data…")
	rawData, err := client.FetchAbstractsData(ctx, ids)
	if err != nil {
		log.Fatalf("FetchAbstractsData: %v", err)
	}

	// Step 3 – write the raw payload to disk.
	if err := writeJSON(*out, rawData); err != nil {
		log.Fatalf("write output: %v", err)
	}
	fmt.Printf("Wrote %s\n", *out)
}
