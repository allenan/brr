package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/allenan/brr/internal/history"
	"github.com/allenan/brr/internal/speedtest"
	"github.com/allenan/brr/internal/tui"
)

var (
	flagJSON       bool
	flagSimple     bool
	flagHistory    bool
	flagCompare    bool
	flagFullscreen bool
	flagTheme      string
	flagServer     string
)

var rootCmd = &cobra.Command{
	Use:   "brr",
	Short: "A delightful TUI speedtest",
	Long:  "brr — a terminal-based internet speed test with bufferbloat grading, sparkline visualizations, and spring-animated numbers.",
	RunE:  run,
}

func init() {
	rootCmd.Flags().BoolVar(&flagJSON, "json", false, "Output results as JSON")
	rootCmd.Flags().BoolVar(&flagSimple, "simple", false, "Output a single summary line")
	rootCmd.Flags().BoolVar(&flagHistory, "history", false, "Show history of past runs")
	rootCmd.Flags().BoolVar(&flagCompare, "compare", false, "Compare current run with previous")
	rootCmd.Flags().BoolVar(&flagFullscreen, "fullscreen", false, "Run in fullscreen (alt-screen) mode")
	rootCmd.Flags().StringVar(&flagTheme, "theme", "default", "Color theme: default, colorblind, mono")
	rootCmd.Flags().StringVar(&flagServer, "server", "", "Override test server")
}

func run(cmd *cobra.Command, args []string) error {
	if flagHistory {
		return showHistory()
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if flagJSON || flagSimple {
		return runHeadless(ctx)
	}

	return runTUI(ctx)
}

type cliCallback struct{}

func (c *cliCallback) OnPhase(phase speedtest.Phase) {
	switch phase {
	case speedtest.PhaseMeta:
		fmt.Fprintf(os.Stderr, "Connecting...\n")
	case speedtest.PhaseLatency:
		fmt.Fprintf(os.Stderr, "Measuring latency...\n")
	case speedtest.PhaseDownload:
		fmt.Fprintf(os.Stderr, "Testing download...\n")
	case speedtest.PhaseUpload:
		fmt.Fprintf(os.Stderr, "Testing upload...\n")
	case speedtest.PhaseDone:
		fmt.Fprintf(os.Stderr, "Done.\n")
	}
}
func (c *cliCallback) OnDownloadSample(s speedtest.Sample)              {}
func (c *cliCallback) OnUploadSample(s speedtest.Sample)                {}
func (c *cliCallback) OnIdleLatencySample(s speedtest.LatencySample)    {}
func (c *cliCallback) OnLoadedLatencySample(s speedtest.LatencySample)  {}

func runHeadless(ctx context.Context) error {
	engine := speedtest.NewEngine()
	result, err := engine.Run(ctx, &cliCallback{})
	if err != nil {
		return err
	}

	// Save to history (unless --json, to not pollute programmatic usage)
	if !flagJSON {
		store := history.NewStore()
		store.Save(result)
	}

	if flagJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	// Simple one-line output
	origin := result.Server.Location
	if result.Server.ClientCity != "" {
		origin = result.Server.ClientCity
	}
	fmt.Printf("↓ %.1f Mbps  ↑ %.1f Mbps  ⏱ %.1fms  Bloat: %s  %s → %s\n",
		result.Download.Mbps,
		result.Upload.Mbps,
		result.IdleLatency.Avg,
		result.BufferbloatDL,
		origin,
		result.Server.ColoCity,
	)
	return nil
}

func runTUI(ctx context.Context) error {
	store := history.NewStore()
	m := tui.NewModel(flagTheme, store)

	opts := []tea.ProgramOption{
		tea.WithMouseCellMotion(),
	}
	if flagFullscreen {
		opts = append(opts, tea.WithAltScreen())
	}

	p := tea.NewProgram(m, opts...)
	m.SetProgram(p)

	_, err := p.Run()
	return err
}

func showHistory() error {
	store := history.NewStore()
	entries, err := store.Last(20)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		fmt.Println("No history yet. Run brr to create your first entry.")
		return nil
	}

	fmt.Printf("%-20s  %-12s  %10s  %10s  %8s  %5s\n",
		"Date", "Server", "Download", "Upload", "Latency", "Grade")
	fmt.Printf("%-20s  %-12s  %10s  %10s  %8s  %5s\n",
		"────────────────────", "────────────", "──────────", "──────────", "────────", "─────")

	for _, e := range entries {
		date := e.Timestamp.Format("2006-01-02 15:04")
		server := e.Server.Colo
		if len(server) == 0 {
			server = "—"
		}
		fmt.Printf("%-20s  %-12s  %8.1f Mbps  %8.1f Mbps  %6.0fms  %5s\n",
			date, server,
			e.Download.Mbps, e.Upload.Mbps,
			e.IdleLatency.Avg, e.BufferbloatDL)
	}
	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
