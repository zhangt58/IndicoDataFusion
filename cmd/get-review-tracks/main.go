package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"IndicoDataFusion/backend/config"
	"IndicoDataFusion/backend/data"
)

func main() {
	var cfgPath string
	var outFile string
	var sourceName string
	flag.StringVar(&cfgPath, "config", "", "path to YAML config file")
	flag.StringVar(&outFile, "out", "review_tracks.json", "output JSON file")
	flag.StringVar(&sourceName, "source", "", "optional data source name (overrides config active data source)")
	flag.Parse()

	if cfgPath == "" {
		fmt.Fprintln(os.Stderr, "usage: get-review-tracks -config /path/to/config.yml [-source name] [-out file]")
		os.Exit(2)
	}

	// Load config
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config %s: %v\n", cfgPath, err)
		os.Exit(1)
	}

	// Resolve data source: explicit -source overrides active data source
	var ds *config.DataSource
	if sourceName != "" {
		ds, err = cfg.GetDataSource(sourceName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to get data source %s: %v\n", sourceName, err)
			os.Exit(1)
		}
	} else {
		ds, err = cfg.GetActiveDataSource()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to get active data source from config: %v\n", err)
			os.Exit(1)
		}
	}

	// Create data handler from resolved data source (resolves api tokens and constructs Indico client)
	handler, err := data.NewDataSourceHandler(ds, cfg.Cache, cfg.APITokens)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create data handler from config: %v\n", err)
		os.Exit(1)
	}

	// Print any non-fatal init problems
	if probs := handler.GetInitProblems(); len(probs) > 0 {
		fmt.Fprintln(os.Stderr, "init warnings:")
		for _, p := range probs {
			fmt.Fprintf(os.Stderr, " - %s\n", p)
		}
	}

	// Fetch review tracks
	reviewTracks, err := handler.GetReviewTracks(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching review tracks: %v\n", err)
		os.Exit(1)
	}

	// Convert review tracks to JSON
	jsonData, err := json.MarshalIndent(reviewTracks, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error converting review tracks to JSON: %v\n", err)
		os.Exit(1)
	}

	// Write JSON data to a file
	err = os.WriteFile(outFile, jsonData, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing JSON to file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Review tracks successfully written to %s\n", outFile)
}
