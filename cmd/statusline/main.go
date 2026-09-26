package main

import (
	"context"
	"flag"
	"fmt"
	"io"
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
	// 정식 이름(claudecode, agy) → 내부 엔진 값 정규화. 기존 값은 그대로.
	cliFlag = normalizeCliAlias(cliFlag)

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
		rest := flag.Args()[1:]
		for _, a := range rest {
			if a == "--cli" || a == "-cli" || strings.HasPrefix(a, "--cli=") || strings.HasPrefix(a, "-cli=") {
				fmt.Fprintln(os.Stderr, "--cli is not supported for collect (ignored)")
				break
			}
		}
		provider, err := parseProviderArg(rest)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		providers, err := selectProviders(provider)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		cachePath, err := collect.ZaiCachePath()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(runCollect(os.Stdout, os.Stderr, providers, credentialsPath(), cachePath))
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
		"CLAUDE_CONFIG_DIR":    os.Getenv("CLAUDE_CONFIG_DIR"),
		"ANTIGRAVITY_APP_DIR":  os.Getenv("ANTIGRAVITY_APP_DIR"),
		"CODEX_ENV":            os.Getenv("CODEX_ENV"),
		"ANTHROPIC_BASE_URL":   os.Getenv("ANTHROPIC_BASE_URL"),
		"ANTHROPIC_AUTH_TOKEN": os.Getenv("ANTHROPIC_AUTH_TOKEN"),
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
	zaiToken := env["ANTHROPIC_AUTH_TOKEN"]
	if zaiToken == "" {
		if t, err := collect.ZaiAuthToken(credentialsPath()); err == nil {
			zaiToken = t // 폴백 실패(권한 등)는 quota 미표시로 흡수 (fail-soft)
		}
	}
	collect.AttachZaiQuota(status, 5*time.Minute, zaiToken)

	output := render.Render(status, cfg)
	if output != "" {
		fmt.Println(output)
	}
}

// credentialsPath — 로컬 자격 증명 저장소 (~/.config/statusline/credentials.json)
func credentialsPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "statusline", "credentials.json")
}
