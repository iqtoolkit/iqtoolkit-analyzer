package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/iqtoolkit/iqtoolkit-analyzer/internal/ai"
	"github.com/iqtoolkit/iqtoolkit-analyzer/internal/dbconn"
	"github.com/iqtoolkit/iqtoolkit-analyzer/internal/logparser"
	"github.com/iqtoolkit/iqtoolkit-analyzer/internal/metrics"
	"github.com/iqtoolkit/iqtoolkit-analyzer/internal/recommendations"
	"github.com/iqtoolkit/iqtoolkit-analyzer/internal/report"
	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{Use: "iqtoolkit-analyzer", Short: "PostgreSQL health checking and performance tuning recommendations"}

	var dsn, logFile string
	var slowThreshold int
	var aiProvider, aiModel string

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

			entries, err := logparser.Parse(f)
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

			// Analyze metrics
			report := metrics.Analyze(entries, settings, time.Duration(slowThreshold)*time.Millisecond)

			// Print summary
			fmt.Printf("=== Summary ===\n")
			fmt.Printf("Total entries:    %d\n", report.TotalEntries)
			fmt.Printf("Error count:      %d\n", report.ErrorCount)
			fmt.Printf("Slow queries:     %d\n", len(report.SlowQueries))
			fmt.Printf("Avg duration:     %v\n", report.AvgDuration)
			if !report.PeakErrorTime.IsZero() {
				fmt.Printf("Peak error time:  %v\n", report.PeakErrorTime)
			}

			// Generate and print recommendations
			recs := recommendations.Generate(report)
			if len(recs) > 0 {
				fmt.Printf("\n=== Recommendations ===\n")
				for _, r := range recs {
					fmt.Printf("[%s][%s] %s\n", r.Severity, r.Category, r.Message)
				}
			} else {
				fmt.Println("\nNo recommendations — configuration looks good!")
			}

			// AI-enhanced analysis
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
				fmt.Printf("\n=== AI-Enhanced Recommendations ===\n%s\n", resp.Content)
			}

			return nil
		},
	}

	analyze.Flags().StringVar(&dsn, "dsn", "", "PostgreSQL connection string")
	analyze.Flags().StringVar(&logFile, "log-file", "", "Path to PostgreSQL log file")
	analyze.Flags().IntVar(&slowThreshold, "slow-threshold", 1000, "Slow query threshold in milliseconds")
	analyze.Flags().StringVar(&aiProvider, "ai-provider", "", "AI provider for enhanced analysis (openai, anthropic, gemini, kiro)")
	analyze.Flags().StringVar(&aiModel, "ai-model", "", "AI model override (default: provider-specific)")
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
