package recommendations

import (
	"fmt"
	"strconv"
	"time"

	"github.com/iqtoolkit/iqtoolkit-analyzer/internal/metrics"
)

type Recommendation struct {
	Severity string // critical, warning, info
	Category string // performance, configuration, reliability
	Message  string
}

func Generate(report *metrics.Report) []Recommendation {
	var recs []Recommendation

	if report.AvgDuration > 500*time.Millisecond {
		recs = append(recs, Recommendation{
			Severity: "warning",
			Category: "performance",
			Message:  fmt.Sprintf("Average query duration is %v. Consider reviewing slow queries and adding indexes.", report.AvgDuration),
		})
	}

	if len(report.SlowQueries) > 10 {
		recs = append(recs, Recommendation{
			Severity: "critical",
			Category: "performance",
			Message:  fmt.Sprintf("%d slow queries detected. Review query plans and consider connection pooling.", len(report.SlowQueries)),
		})
	}

	if report.ErrorCount > 100 {
		recs = append(recs, Recommendation{
			Severity: "critical",
			Category: "reliability",
			Message:  fmt.Sprintf("%d errors found. Peak error time: %v. Investigate application errors.", report.ErrorCount, report.PeakErrorTime),
		})
	}

	// Table-level checks from pg_stat_user_tables.
	for _, t := range report.Tables {
		if t.NDeadTup > 10000 && t.NLiveTup > 0 && float64(t.NDeadTup)/float64(t.NLiveTup) > 0.2 {
			recs = append(recs, Recommendation{
				Severity: "warning",
				Category: "maintenance",
				Message:  fmt.Sprintf("Table %s.%s has %d dead tuples (%.0f%% of live). Run VACUUM ANALYZE or tune autovacuum.", t.Schema, t.Table, t.NDeadTup, 100*float64(t.NDeadTup)/float64(t.NLiveTup)),
			})
		}
		if t.IdxScan == 0 && t.SeqScan > 1000 && t.NLiveTup > 10000 {
			recs = append(recs, Recommendation{
				Severity: "warning",
				Category: "performance",
				Message:  fmt.Sprintf("Table %s.%s: %d sequential scans, no index scans, %d rows. Consider adding indexes for common queries.", t.Schema, t.Table, t.SeqScan, t.NLiveTup),
			})
		}
	}

	for _, s := range report.Settings {
		switch s.Name {
		case "autovacuum":
			if s.Value == "off" {
				recs = append(recs, Recommendation{
					Severity: "critical",
					Category: "maintenance",
					Message:  "autovacuum is DISABLED. This is never acceptable: dead tuples will accumulate, tables will bloat, and the database is at risk of transaction ID wraparound shutdown. Re-enable autovacuum immediately.",
				})
			}
		case "effective_cache_size":
			// pg_settings stores effective_cache_size in 8kB pages; 524288 = 4GB (default)
			if s.Value == "524288" {
				recs = append(recs, Recommendation{
					Severity: "info",
					Category: "configuration",
					Message:  "effective_cache_size is at default (4GB). Set to ~75% of available RAM so the planner prefers index scans.",
				})
			}
		case "max_connections":
			if n, err := strconv.Atoi(s.Value); err == nil && n > 500 {
				recs = append(recs, Recommendation{
					Severity: "warning",
					Category: "configuration",
					Message:  fmt.Sprintf("max_connections is %d. High connection counts waste memory; use a connection pooler (e.g. PgBouncer) instead.", n),
				})
			}
		case "checkpoint_completion_target":
			if v, err := strconv.ParseFloat(s.Value, 64); err == nil && v < 0.9 {
				recs = append(recs, Recommendation{
					Severity: "info",
					Category: "configuration",
					Message:  fmt.Sprintf("checkpoint_completion_target is %s. Set to 0.9 to spread checkpoint I/O and avoid write spikes.", s.Value),
				})
			}
		case "shared_buffers":
			// pg_settings stores shared_buffers in 8kB pages; 16384 pages = 128MB (default)
			if s.Value == "16384" || s.Value == "1024" {
				recs = append(recs, Recommendation{
					Severity: "warning",
					Category: "configuration",
					Message:  "shared_buffers is at default. Set to ~25% of available RAM.",
				})
			}
		case "work_mem":
			// pg_settings stores work_mem in kB; 4096 = 4MB (default)
			if s.Value == "4096" {
				recs = append(recs, Recommendation{
					Severity: "info",
					Category: "configuration",
					Message:  "work_mem is at default (4MB). Increase for complex sort/hash operations.",
				})
			}
		case "log_min_duration_statement":
			if s.Value == "-1" {
				recs = append(recs, Recommendation{
					Severity: "info",
					Category: "configuration",
					Message:  "log_min_duration_statement is disabled. Enable to capture slow queries.",
				})
			}
		}
	}

	return recs
}
