package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/iqtoolkit/iqtoolkit-analyzer/internal/dbconn"
	"github.com/iqtoolkit/iqtoolkit-analyzer/internal/logparser"
	"github.com/iqtoolkit/iqtoolkit-analyzer/internal/metrics"
	"github.com/iqtoolkit/iqtoolkit-analyzer/internal/recommendations"
	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{Use: "iqtoolkit-analyzer", Short: "PostgreSQL health checking and performance tuning recommendations"}

	var dsn, logFile string
	var slowThreshold int

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

			return nil
		},
	}

	analyze.Flags().StringVar(&dsn, "dsn", "", "PostgreSQL connection string")
	analyze.Flags().StringVar(&logFile, "log-file", "", "Path to PostgreSQL log file")
	analyze.Flags().IntVar(&slowThreshold, "slow-threshold", 1000, "Slow query threshold in milliseconds")
	analyze.MarkFlagRequired("dsn")
	analyze.MarkFlagRequired("log-file")

	root.AddCommand(analyze)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
