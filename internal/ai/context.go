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

	if len(report.Settings) > 0 {
		b.WriteString("\n## PostgreSQL Settings\n\n")
		for _, s := range report.Settings {
			fmt.Fprintf(&b, "- %s = %s (source: %s)\n", s.Name, s.Value, s.Source)
		}
	}

	return b.String()
}
