package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"statusline/internal/adapter"
	"statusline/internal/config"
	"statusline/internal/render"
	"statusline/internal/vcs"
)

func main() {
	var cliFlag string
	var configPath string
	var autoFlag bool

	flag.StringVar(&cliFlag, "cli", "auto", "Target CLI adapter (auto, claude, codex, antigravity)")
	flag.StringVar(&configPath, "config", "", "Path to config file")
	flag.BoolVar(&autoFlag, "auto", false, "Alias for --cli=auto")
	flag.Parse()

	if autoFlag {
		cliFlag = "auto"
	}

	if configPath == "" {
		home, _ := os.UserHomeDir()
		configPath = filepath.Join(home, ".config", "statusline", "config.json")
	}

	// Handle init subcommand
	if flag.NArg() > 0 && flag.Arg(0) == "init" {
		if err := config.SaveDefaultConfig(configPath); err != nil {
			fmt.Fprintf(os.Stderr, "failed to initialize config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Initialized default config at %s\n", configPath)
		return
	}

	// Fail-soft stdin read
	var rawInput []byte
	fi, err := os.Stdin.Stat()
	if err == nil && (fi.Mode()&os.ModeCharDevice) == 0 {
		rawInput, _ = io.ReadAll(os.Stdin)
	}

	// Environment mapping
	env := map[string]string{
		"CLAUDE_CONFIG_DIR":   os.Getenv("CLAUDE_CONFIG_DIR"),
		"ANTIGRAVITY_APP_DIR": os.Getenv("ANTIGRAVITY_APP_DIR"),
		"CODEX_ENV":           os.Getenv("CODEX_ENV"),
	}

	cfg := config.LoadConfig(configPath)
	status, _ := adapter.ParseInput(cliFlag, rawInput, env)

	// Enrich Git status asynchronously with 10ms budget
	cwd, _ := os.Getwd()
	if status.Cwd == "" {
		status.Cwd = cwd
	}
	if status.Hostname == "" {
		status.Hostname, _ = os.Hostname()
	}
	vcs.EnrichGit(status, status.Cwd, 10)

	output := render.Render(status, cfg)
	if output != "" {
		fmt.Println(output)
	}
}
