package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"html"
	"log"
	"net/http"
	"time"

	"IndicoDataFusion/backend"
)

func main() {
	cfgPath := flag.String("config", "backend/config.yaml", "path to config yaml")
	detail := flag.String("detail", "", "detail query value (optional)")
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
	ev, err := client.GetEventInfo(ctx, *detail)
	if err != nil {
		log.Fatalf("GetEventInfo failed: %v", err)
	}

	// Ensure description contains real HTML tags (backend may already unescape it)
	ev.Description = html.UnescapeString(ev.Description)

	// Encode without escaping HTML, then pretty-print the JSON
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(ev); err != nil {
		log.Fatalf("encode event: %v", err)
	}
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, buf.Bytes(), "", "  "); err != nil {
		// fallback: print compact unescaped JSON
		fmt.Print(buf.String())
		return
	}
	fmt.Print(pretty.String())
}
