package recommendations

import (
	"fmt"
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

	for _, s := range report.Settings {
		switch s.Name {
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
