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

	"IndicoDataFusion/backend"
)

func main() {
	cfgPath := flag.String("config", "backend/config.yaml", "path to config yaml")
	out := flag.String("out", "abstracts.json", "Output JSON file")
	flag.Parse()

	cfg, err := backend.LoadConfig(*cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	client := backend.NewIndicoClient(cfg.BaseURL, cfg.EventID, cfg.APIToken)
	// apply configured timeout if present
	client.Timeout = time.Duration(cfg.Timeout)
	client.Client = &http.Client{Timeout: time.Duration(cfg.Timeout)}

	ctx := context.Background()

	ids, csrf, err := client.GetAbstractIDsAndCSRFFromList(ctx)
	if err != nil {
		log.Fatalf("GetAbstractIDsAndCSRFFromList failed: %v", err)
	}
	log.Printf("found %d abstract ids, csrf=%s", len(ids), csrf)

	if len(ids) == 0 {
		log.Fatalf("no abstract ids found")
	}

	data, err := client.FetchAbstractsData(ctx, ids, csrf)
	if err != nil {
		log.Fatalf("FetchAbstractsData failed: %v", err)
	}

	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Fatalf("json marshal failed: %v", err)
	}

	if err := os.WriteFile(*out, b, 0644); err != nil {
		log.Fatalf("write file failed: %v", err)
	}

	fmt.Printf("Wrote %s (%d bytes)\n", *out, len(b))
}
