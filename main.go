package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/iqtoolkit/iqtoolkit-analyzer/internal/ai"
	"github.com/iqtoolkit/iqtoolkit-analyzer/internal/dbconn"
	"github.com/iqtoolkit/iqtoolkit-analyzer/internal/logparser"
	"github.com/iqtoolkit/iqtoolkit-analyzer/internal/metrics"
	"github.com/iqtoolkit/iqtoolkit-analyzer/internal/output"
	"github.com/iqtoolkit/iqtoolkit-analyzer/internal/recommendations"
	"github.com/iqtoolkit/iqtoolkit-analyzer/internal/report"
	"github.com/spf13/cobra"
)

// Set via -ldflags at build time, or auto-detected from go install:
//
//	go build -ldflags "-X main.version=v1.0.0 -X main.commit=abc1234 -X main.date=2026-06-04"
var (
	version = ""
	commit  = ""
	date    = ""
)

func formatVersion() string {
	if version == "" {
		// Auto-detect from go install metadata (uses git tag)
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
			version = info.Main.Version
			for _, s := range info.Settings {
				switch s.Key {
				case "vcs.revision":
					if len(s.Value) > 7 {
						commit = s.Value[:7]
					} else {
						commit = s.Value
					}
				case "vcs.time":
					if t, err := time.Parse(time.RFC3339, s.Value); err == nil {
						date = t.Format("2006-01-02")
					}
				}
			}
		}
	}
	if version == "" {
		return "dev"
	}
	if commit != "" {
		return fmt.Sprintf("%s (%s, %s)", version, commit, date)
	}
	return version
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	root := &cobra.Command{Use: "iqtoolkit-analyzer", Short: "PostgreSQL health checking and performance tuning recommendations"}
	root.Version = formatVersion()
	root.SetVersionTemplate("{{.Version}}\n")

	root.AddCommand(newAnalyzeCmd())
	root.AddCommand(newReportCmd())

	if err := root.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}

func newAnalyzeCmd() *cobra.Command {
	var dsn, logFile string
	var slowThreshold int
	var aiProvider, aiModel string
	var outputFile, outputFormat string
	var logFormat string

	analyze := &cobra.Command{
		Use:   "analyze",
		Short: "Analyze PostgreSQL logs and configuration",
		Long:  "Analyze PostgreSQL logs and configuration.\n\nIf --dsn is omitted, runs in log-only mode: settings and runtime statistics are skipped.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// Parse log file
			f, err := os.Open(logFile)
			if err != nil {
				return fmt.Errorf("opening log file: %w", err)
			}
			defer f.Close()

			entries, err := logparser.ParseFormat(f, logparser.Format(logFormat))
			if err != nil {
				return fmt.Errorf("parsing log file: %w", err)
			}

			// Connect to database and fetch settings (optional: log-only mode without --dsn)
			var settings []dbconn.Setting
			var conn *dbconn.Conn
			if dsn != "" {
				conn, err = dbconn.Connect(ctx, dsn)
				if err != nil {
					return fmt.Errorf("connecting to database: %w", err)
				}
				defer conn.Close(ctx)

				settings, err = conn.Settings(ctx)
				if err != nil {
					return fmt.Errorf("fetching settings: %w", err)
				}

				// Check required extensions
				for _, ext := range dbconn.RequiredExtensions {
					status, err := conn.CheckExtension(ctx, ext)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Warning: could not check extension %s: %v\n", ext, err)
						continue
					}
					if !status.Installed {
						if status.Available {
							fmt.Fprintf(os.Stderr, "Extension %q is available but not installed. Run: CREATE EXTENSION %s;\n", ext, ext)
						} else {
							fmt.Fprintf(os.Stderr, "Extension %q is not available on this server.\n", ext)
						}
					}
				}
			} else {
				fmt.Fprintln(os.Stderr, "No --dsn provided: running in log-only mode (settings and runtime stats skipped)")
			}

			// Analyze metrics
			rep := metrics.Analyze(entries, settings, time.Duration(slowThreshold)*time.Millisecond)

			// Collect extended stats (best-effort)
			if conn != nil {
				if stmts, err := conn.StatStatements(ctx, 20); err == nil {
					rep.Statements = stmts
				}
				if tables, err := conn.StatUserTables(ctx); err == nil {
					rep.Tables = tables
				}
				if indexes, err := conn.StatUserIndexes(ctx); err == nil {
					rep.Indexes = indexes
				}
				if bc, err := conn.StatBufferCache(ctx, 20); err == nil {
					rep.BufferCache = bc
				}
			}

			// Generate recommendations
			recs := recommendations.Generate(rep)

			// AI-enhanced analysis
			var aiContent string
			if aiProvider != "" {
				client, err := ai.ClientFromConfig(ai.Provider(aiProvider))
				if err != nil {
					return fmt.Errorf("configuring AI provider: %w", err)
				}
				model := aiModel
				if model == "" {
					model = ai.DefaultModel(ai.Provider(aiProvider))
				}
				prompt := ai.BuildPrompt(rep)
				resp, err := client.Complete(ctx, ai.Request{
					Model:    model,
					System:   "You are a PostgreSQL performance tuning expert. Analyze the provided metrics and settings, then give concise, prioritized recommendations.",
					Messages: []ai.Message{{Role: "user", Content: prompt}},
				})
				if err != nil {
					return fmt.Errorf("AI analysis failed: %w", err)
				}
				aiContent = resp.Content
			}

			// Determine output writer
			var w io.Writer = os.Stdout
			if outputFile != "" {
				of, err := os.Create(outputFile)
				if err != nil {
					return fmt.Errorf("creating output file: %w", err)
				}
				defer func() {
					if cerr := of.Close(); cerr != nil && err == nil {
						fmt.Fprintf(os.Stderr, "Warning: closing output file: %v\n", cerr)
					}
				}()
				w = of
			}

			if err := output.Write(w, output.Format(outputFormat), rep, recs, aiContent); err != nil {
				return fmt.Errorf("writing output: %w", err)
			}

			if outputFile != "" {
				fmt.Fprintf(os.Stderr, "Report written to %s\n", outputFile)
			}

			return nil
		},
	}

	analyze.Flags().StringVar(&dsn, "dsn", "", "PostgreSQL connection string (optional; omit for log-only mode)")
	analyze.Flags().StringVar(&logFile, "log-file", "", "Path to PostgreSQL log file")
	analyze.Flags().StringVar(&logFormat, "log-format", "", "Log format: stderr, csvlog, jsonlog (default: auto-detect)")
	analyze.Flags().IntVar(&slowThreshold, "slow-threshold", 1000, "Slow query threshold in milliseconds")
	analyze.Flags().StringVar(&aiProvider, "ai-provider", "", "AI provider for enhanced analysis (openai, anthropic, gemini, kiro)")
	analyze.Flags().StringVar(&aiModel, "ai-model", "", "AI model override (default: provider-specific)")
	analyze.Flags().StringVar(&outputFile, "output", "", "Write output to file instead of stdout")
	analyze.Flags().StringVar(&outputFormat, "format", "text", "Output format: text, json, or markdown")
	cobra.CheckErr(analyze.MarkFlagRequired("log-file"))

	return analyze
}

func newReportCmd() *cobra.Command {
	var reportDSN, reportOutput string
	reportCmd := &cobra.Command{
		Use:   "report",
		Short: "Generate an HTML report with settings, extensions, and version",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			conn, err := dbconn.Connect(ctx, reportDSN)
			if err != nil {
				return fmt.Errorf("connecting to database: %w", err)
			}
			defer conn.Close(ctx)

			version, err := conn.Version(ctx)
			if err != nil {
				return fmt.Errorf("fetching version: %w", err)
			}
			settings, err := conn.Settings(ctx)
			if err != nil {
				return fmt.Errorf("fetching settings: %w", err)
			}
			extensions, err := conn.Extensions(ctx)
			if err != nil {
				return fmt.Errorf("fetching extensions: %w", err)
			}

			f, err := os.Create(reportOutput)
			if err != nil {
				return fmt.Errorf("creating output file: %w", err)
			}

			if err := report.Generate(f, report.Data{
				Version:     version,
				Settings:    settings,
				Extensions:  extensions,
				GeneratedAt: time.Now(),
			}); err != nil {
				f.Close()
				return fmt.Errorf("generating report: %w", err)
			}
			if err := f.Close(); err != nil {
				return fmt.Errorf("closing output file: %w", err)
			}
			fmt.Printf("Report written to %s\n", reportOutput)
			return nil
		},
	}
	reportCmd.Flags().StringVar(&reportDSN, "dsn", "", "PostgreSQL connection string")
	reportCmd.Flags().StringVar(&reportOutput, "output", "report.html", "Output HTML file path")
	cobra.CheckErr(reportCmd.MarkFlagRequired("dsn"))
	return reportCmd
}
