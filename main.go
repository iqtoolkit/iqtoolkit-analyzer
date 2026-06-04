package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"time"

	"github.com/iqtoolkit/iqtoolkit-analyzer/internal/ai"
	"github.com/iqtoolkit/iqtoolkit-analyzer/internal/dbconn"
	"github.com/iqtoolkit/iqtoolkit-analyzer/internal/logparser"
	"github.com/iqtoolkit/iqtoolkit-analyzer/internal/metrics"
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
	root := &cobra.Command{Use: "iqtoolkit-analyzer", Short: "PostgreSQL health checking and performance tuning recommendations"}
	root.Version = formatVersion()
	root.SetVersionTemplate("{{.Version}}\n")

	var dsn, logFile string
	var slowThreshold int
	var aiProvider, aiModel string
	var outputFile, outputFormat string
	var logFormat string

	analyze := &cobra.Command{
		Use:   "analyze",
		Short: "Analyze PostgreSQL logs and configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

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

			// Connect to database and fetch settings
			conn, err := dbconn.Connect(ctx, dsn)
			if err != nil {
				return fmt.Errorf("connecting to database: %w", err)
			}
			defer conn.Close(ctx)

			settings, err := conn.Settings(ctx)
			if err != nil {
				return fmt.Errorf("fetching settings: %w", err)
			}

			// Check required extensions and collect stats
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

			// Analyze metrics
			report := metrics.Analyze(entries, settings, time.Duration(slowThreshold)*time.Millisecond)

			// Collect extended stats (best-effort)
			if stmts, err := conn.StatStatements(ctx, 20); err == nil {
				report.Statements = stmts
			}
			if tables, err := conn.StatUserTables(ctx); err == nil {
				report.Tables = tables
			}
			if indexes, err := conn.StatUserIndexes(ctx); err == nil {
				report.Indexes = indexes
			}
			if bc, err := conn.StatBufferCache(ctx, 20); err == nil {
				report.BufferCache = bc
			}

			// Generate recommendations
			recs := recommendations.Generate(report)

			// AI-enhanced analysis
			var aiContent string
			if aiProvider != "" {
				client, err := ai.ClientFromConfig(ai.Provider(aiProvider))
				if err != nil {
					return fmt.Errorf("configuring AI provider: %w", err)
				}
				model := aiModel
				if model == "" {
					switch ai.Provider(aiProvider) {
					case ai.OpenAI:
						model = "gpt-4o"
					case ai.Anthropic:
						model = "claude-sonnet-4-20250514"
					case ai.Gemini:
						model = "gemini-2.5-pro"
					case ai.Kiro:
						model = "anthropic.claude-sonnet-4-20250514-v1:0"
					}
				}
				prompt := ai.BuildPrompt(report)
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
				defer of.Close()
				w = of
			}

			switch outputFormat {
			case "json":
				out := struct {
					Summary struct {
						TotalEntries  int    `json:"total_entries"`
						ErrorCount    int    `json:"error_count"`
						SlowQueries   int    `json:"slow_queries"`
						AvgDuration   string `json:"avg_duration"`
						PeakErrorTime string `json:"peak_error_time,omitempty"`
					} `json:"summary"`
					Recommendations []struct {
						Severity string `json:"severity"`
						Category string `json:"category"`
						Message  string `json:"message"`
					} `json:"recommendations"`
					AIRecommendations string `json:"ai_recommendations,omitempty"`
				}{}
				out.Summary.TotalEntries = report.TotalEntries
				out.Summary.ErrorCount = report.ErrorCount
				out.Summary.SlowQueries = len(report.SlowQueries)
				out.Summary.AvgDuration = report.AvgDuration.String()
				if !report.PeakErrorTime.IsZero() {
					out.Summary.PeakErrorTime = report.PeakErrorTime.Format(time.RFC3339)
				}
				for _, r := range recs {
					out.Recommendations = append(out.Recommendations, struct {
						Severity string `json:"severity"`
						Category string `json:"category"`
						Message  string `json:"message"`
					}{r.Severity, r.Category, r.Message})
				}
				out.AIRecommendations = aiContent
				enc := json.NewEncoder(w)
				enc.SetIndent("", "  ")
				if err := enc.Encode(out); err != nil {
					return err
				}

			case "markdown":
				fmt.Fprintf(w, "# PostgreSQL Analysis Report\n\n")
				fmt.Fprintf(w, "## Summary\n\n")
				fmt.Fprintf(w, "| Metric | Value |\n|--------|-------|\n")
				fmt.Fprintf(w, "| Total entries | %d |\n", report.TotalEntries)
				fmt.Fprintf(w, "| Error count | %d |\n", report.ErrorCount)
				fmt.Fprintf(w, "| Slow queries | %d |\n", len(report.SlowQueries))
				fmt.Fprintf(w, "| Avg duration | %v |\n", report.AvgDuration)
				if !report.PeakErrorTime.IsZero() {
					fmt.Fprintf(w, "| Peak error time | %v |\n", report.PeakErrorTime)
				}
				if len(recs) > 0 {
					fmt.Fprintf(w, "\n## Recommendations\n\n")
					for _, r := range recs {
						fmt.Fprintf(w, "- **[%s]** (%s) %s\n", r.Severity, r.Category, r.Message)
					}
				}
				if aiContent != "" {
					fmt.Fprintf(w, "\n## AI-Enhanced Recommendations\n\n%s\n", aiContent)
				}

			default: // text
				fmt.Fprintf(w, "=== Summary ===\n")
				fmt.Fprintf(w, "Total entries:    %d\n", report.TotalEntries)
				fmt.Fprintf(w, "Error count:      %d\n", report.ErrorCount)
				fmt.Fprintf(w, "Slow queries:     %d\n", len(report.SlowQueries))
				fmt.Fprintf(w, "Avg duration:     %v\n", report.AvgDuration)
				if !report.PeakErrorTime.IsZero() {
					fmt.Fprintf(w, "Peak error time:  %v\n", report.PeakErrorTime)
				}
				if len(recs) > 0 {
					fmt.Fprintf(w, "\n=== Recommendations ===\n")
					for _, r := range recs {
						fmt.Fprintf(w, "[%s][%s] %s\n", r.Severity, r.Category, r.Message)
					}
				} else {
					fmt.Fprintln(w, "\nNo recommendations — configuration looks good!")
				}
				if aiContent != "" {
					fmt.Fprintf(w, "\n=== AI-Enhanced Recommendations ===\n%s\n", aiContent)
				}
			}

			if outputFile != "" {
				fmt.Fprintf(os.Stderr, "Report written to %s\n", outputFile)
			}

			return nil
		},
	}

	analyze.Flags().StringVar(&dsn, "dsn", "", "PostgreSQL connection string")
	analyze.Flags().StringVar(&logFile, "log-file", "", "Path to PostgreSQL log file")
	analyze.Flags().StringVar(&logFormat, "log-format", "", "Log format: stderr, csvlog, jsonlog (default: auto-detect)")
	analyze.Flags().IntVar(&slowThreshold, "slow-threshold", 1000, "Slow query threshold in milliseconds")
	analyze.Flags().StringVar(&aiProvider, "ai-provider", "", "AI provider for enhanced analysis (openai, anthropic, gemini, kiro)")
	analyze.Flags().StringVar(&aiModel, "ai-model", "", "AI model override (default: provider-specific)")
	analyze.Flags().StringVar(&outputFile, "output", "", "Write output to file instead of stdout")
	analyze.Flags().StringVar(&outputFormat, "format", "text", "Output format: text, json, or markdown")
	analyze.MarkFlagRequired("dsn")
	analyze.MarkFlagRequired("log-file")

	root.AddCommand(analyze)

	var reportDSN, reportOutput string
	reportCmd := &cobra.Command{
		Use:   "report",
		Short: "Generate an HTML report with settings, extensions, and version",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
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
			defer f.Close()

			err = report.Generate(f, report.Data{
				Version:     version,
				Settings:    settings,
				Extensions:  extensions,
				GeneratedAt: time.Now(),
			})
			if err != nil {
				return fmt.Errorf("generating report: %w", err)
			}
			fmt.Printf("Report written to %s\n", reportOutput)
			return nil
		},
	}
	reportCmd.Flags().StringVar(&reportDSN, "dsn", "", "PostgreSQL connection string")
	reportCmd.Flags().StringVar(&reportOutput, "output", "report.html", "Output HTML file path")
	reportCmd.MarkFlagRequired("dsn")
	root.AddCommand(reportCmd)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
