package ai

import (
	"fmt"
	"strings"

	"github.com/iqtoolkit/iqtoolkit-analyzer/internal/metrics"
)

// BuildPrompt formats a metrics Report into a structured prompt for AI analysis.
func BuildPrompt(report *metrics.Report) string {
	var b strings.Builder

	b.WriteString("## PostgreSQL Health Metrics\n\n")
	fmt.Fprintf(&b, "- Total log entries: %d\n", report.TotalEntries)
	fmt.Fprintf(&b, "- Error count: %d\n", report.ErrorCount)
	fmt.Fprintf(&b, "- Slow query count: %d\n", len(report.SlowQueries))
	fmt.Fprintf(&b, "- Average query duration: %s\n", report.AvgDuration)
	if !report.PeakErrorTime.IsZero() {
		fmt.Fprintf(&b, "- Peak error hour: %s\n", report.PeakErrorTime.Format("2006-01-02 15:00"))
	}

	if len(report.SlowQueries) > 0 {
		b.WriteString("\n## Slow Queries (top 10)\n\n")
		limit := len(report.SlowQueries)
		if limit > 10 {
			limit = 10
		}
		for _, q := range report.SlowQueries[:limit] {
			fmt.Fprintf(&b, "- [%s] %s (duration: %s)\n", q.Timestamp.Format("15:04:05"), q.Message, q.Duration)
		}
	}

	if len(report.Statements) > 0 {
		b.WriteString("\n## Top Queries by Total Time (pg_stat_statements)\n\n")
		for _, s := range report.Statements {
			fmt.Fprintf(&b, "- calls=%d total=%.1fms mean=%.1fms rows=%d blks_hit=%d blks_read=%d | %s\n",
				s.Calls, s.TotalExecTime, s.MeanExecTime, s.Rows, s.SharedBlksHit, s.SharedBlksRead, truncate(s.Query, 120))
		}
	}

	if len(report.Tables) > 0 {
		b.WriteString("\n## Table Statistics (pg_stat_user_tables)\n\n")
		for _, t := range report.Tables {
			fmt.Fprintf(&b, "- %s.%s: seq_scan=%d idx_scan=%d dead_tup=%d live_tup=%d\n",
				t.Schema, t.Table, t.SeqScan, t.IdxScan, t.NDeadTup, t.NLiveTup)
		}
	}

	if len(report.Indexes) > 0 {
		b.WriteString("\n## Unused/Underused Indexes (pg_stat_user_indexes, scan < 10)\n\n")
		for _, i := range report.Indexes {
			if i.IdxScan < 10 {
				fmt.Fprintf(&b, "- %s.%s.%s: scans=%d\n", i.Schema, i.Table, i.IndexName, i.IdxScan)
			}
		}
	}

	if len(report.BufferCache) > 0 {
		b.WriteString("\n## Buffer Cache Usage (pg_buffercache)\n\n")
		for _, bc := range report.BufferCache {
			fmt.Fprintf(&b, "- %s.%s: %d buffers (%.1f MB)\n", bc.Schema, bc.Table, bc.Buffers, bc.SizeMB)
		}
	}

	if len(report.Settings) > 0 {
		b.WriteString("\n## PostgreSQL Settings\n\n")
		for _, s := range report.Settings {
			fmt.Fprintf(&b, "- %s = %s (source: %s)\n", s.Name, s.Value, s.Source)
		}
	}

	return b.String()
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
