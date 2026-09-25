package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"statusline/internal/adapter"
	"statusline/internal/collect"
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
	if flag.NArg() > 0 && flag.Arg(0) == "collect" {
		provider := ""
		for _, a := range flag.Args()[1:] {
			if strings.HasPrefix(a, "--provider=") {
				provider = strings.TrimPrefix(a, "--provider=")
			}
		}
		if provider == "zai" || provider == "zhipu" {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			cachePath, err := collect.ZaiCachePath()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			// 플래그 값은 플로우 선택용 — 실제 provider는 env URL 기준
			zaiProvider := adapter.ProviderFromBaseURL(os.Getenv("ANTHROPIC_BASE_URL"))
			result := collect.RunZaiCollect(ctx, &http.Client{Timeout: 5 * time.Second}, os.Getenv("ANTHROPIC_BASE_URL"), os.Getenv("ANTHROPIC_AUTH_TOKEN"), zaiProvider, cachePath, collect.ZaiRefreshPath(cachePath))
			if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		result := struct {
			Codex    map[string]json.RawMessage `json:"codex,omitempty"`
			Platform *collect.PlatformSnapshot  `json:"platform,omitempty"`
			Errors   map[string]string          `json:"errors,omitempty"`
		}{Errors: make(map[string]string)}
		var err error
		result.Codex, err = collect.Codex(ctx, "codex")
		if err != nil {
			result.Errors["codex"] = err.Error()
		}
		home, _ := os.UserHomeDir()
		key, keyErr := collect.AdminKey(filepath.Join(home, ".config", "statusline", "credentials.json"))
		if keyErr != nil {
			result.Errors["credentials"] = keyErr.Error()
		}
		if key != "" {
			platform, platformErr := collect.Platform(ctx, &http.Client{Timeout: 5 * time.Second}, "https://api.openai.com/v1", key, time.Now().UTC().AddDate(0, 0, -1))
			if platformErr != nil {
				result.Errors["platform"] = platformErr.Error()
			} else {
				result.Platform = &platform
			}
		}
		if len(result.Errors) == 0 {
			result.Errors = nil
		}
		if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if flag.NArg() > 0 && flag.Arg(0) == "watch" {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()
		if err := collect.CodexWatch(ctx, "codex", os.Stdout); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
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
	if status.TerminalWidth <= 0 {
		if cols, err := strconv.Atoi(os.Getenv("COLUMNS")); err == nil && cols > 0 {
			status.TerminalWidth = cols
		}
	}
	vcs.EnrichGit(status, status.Cwd, 10)

	output := render.Render(status, cfg)
	if output != "" {
		fmt.Println(output)
	}
}
