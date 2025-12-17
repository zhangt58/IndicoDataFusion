package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"IndicoDataFusion/backend"
)

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		_ = out.Close()
	}()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

func main() {
	cfgPath := flag.String("config", "", "Path to config YAML file (required)")
	dryRun := flag.Bool("dry-run", true, "Dry run mode; show planned actions but don't modify keyring or config")
	overwrite := flag.Bool("overwrite", false, "If true, overwrite existing keyring secrets with the YAML token value")
	backup := flag.Bool("backup", true, "Create a backup of the config file before modifying it")
	flag.Parse()

	if *cfgPath == "" {
		// try env
		if env := os.Getenv("INDICO_DATA_FUSION_CONFIG_PATH"); env != "" {
			*cfgPath = env
		}
	}
	if *cfgPath == "" {
		fmt.Fprintln(os.Stderr, "Error: --config is required (or set INDICO_DATA_FUSION_CONFIG_PATH)")
		flag.Usage()
		os.Exit(2)
	}

	absPath, err := filepath.Abs(*cfgPath)
	if err != nil {
		log.Fatalf("Failed to resolve config path: %v", err)
	}

	cfg, err := backend.LoadConfig(absPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if cfg.APITokens == nil || len(cfg.APITokens) == 0 {
		log.Println("No api-tokens section found in config; nothing to migrate.")
		return
	}

	// Prepare summary
	total := 0
	tmigrated := 0
	tskipped := 0
	actions := []string{}

	for i := range cfg.APITokens {
		entry := &cfg.APITokens[i]
		if entry.Token == "" {
			actions = append(actions, fmt.Sprintf("[%s] no token to migrate", entry.Name))
			tskipped++
			continue
		}
		total++
		// Check existing keyring secret
		existing, gerr := backend.GetAPITokenSecret(entry.Name)
		if gerr == nil && existing != "" {
			if !*overwrite {
				actions = append(actions, fmt.Sprintf("[%s] keyring already has a secret, skipping (use --overwrite to replace)", entry.Name))
				tskipped++
				continue
			}
			// else we'll overwrite
		}
		if *dryRun {
			actions = append(actions, fmt.Sprintf("[%s] would store secret to keyring and clear YAML token", entry.Name))
			continue
		}
		// store in keyring
		if err := backend.SetAPITokenSecret(entry.Name, entry.Token); err != nil {
			actions = append(actions, fmt.Sprintf("[%s] FAILED to store secret: %v", entry.Name, err))
			continue
		}
		// clear YAML token
		entry.Token = ""
		actions = append(actions, fmt.Sprintf("[%s] migrated token into keyring and cleared YAML", entry.Name))
		tmigrated++
	}

	// If dry-run, print summary and exit
	if *dryRun {
		log.Println("Dry run mode — no changes made. Planned actions:")
		for _, a := range actions {
			log.Println(" -", a)
		}
		log.Printf("Summary: total with tokens=%d, migrated=0, skipped=%d\n", total, tskipped)
		return
	}

	// Create backup if requested
	if *backup {
		bak := absPath + ".migrated." + time.Now().Format("20060102T150405") + ".bak"
		if err := copyFile(absPath, bak); err != nil {
			log.Fatalf("Failed to create backup %s: %v", bak, err)
		}
		log.Printf("Backup created: %s\n", bak)
	}

	// Save updated config (with cleared tokens)
	if err := backend.SaveConfig(absPath, cfg); err != nil {
		log.Fatalf("Failed to save migrated config: %v", err)
	}

	log.Println("Migration completed. Actions:")
	for _, a := range actions {
		log.Println(" -", a)
	}
	log.Printf("Summary: total with tokens=%d, migrated=%d, skipped=%d\n", total, tmigrated, tskipped)
}
