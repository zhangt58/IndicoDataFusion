package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"IndicoDataFusion/backend"
	"gopkg.in/yaml.v3"
)

const ConfEnvName = "INDICO_DATA_FUSION_CONFIG_PATH"

func main() {
	cfgPathFlag := flag.String("config", "", "path to config YAML file")
	raw := flag.Bool("raw", false, "print raw file without parsing")
	jsonOut := flag.Bool("json", false, "print parsed config as JSON instead of YAML")
	flag.Parse()

	var cfgPath string
	if *cfgPathFlag != "" {
		cfgPath = *cfgPathFlag
	} else if env := os.Getenv(ConfEnvName); env != "" {
		cfgPath = env
	} else {
		cfgPath = "config.yaml"
	}

	if *raw {
		b, err := os.ReadFile(cfgPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to read %s: %v\n", cfgPath, err)
			os.Exit(2)
		}
		fmt.Print(string(b))
		return
	}

	cfg, err := backend.LoadConfig(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		// attempt to print raw file as fallback
		if b, rerr := os.ReadFile(cfgPath); rerr == nil {
			fmt.Print(string(b))
			os.Exit(1)
		}
		os.Exit(2)
	}

	// Validate and report issues but continue to print the parsed config
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "config validation error: %v\n", err)
	}

	if *jsonOut {
		out, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to marshal config to JSON: %v\n", err)
			os.Exit(3)
		}
		fmt.Print(string(out))
		return
	}

	out, err := yaml.Marshal(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to marshal config: %v\n", err)
		os.Exit(3)
	}
	fmt.Print(string(out))
}
