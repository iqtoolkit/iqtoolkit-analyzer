// Package output renders analysis results in text, JSON, or markdown format.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/iqtoolkit/iqtoolkit-analyzer/internal/logparser"
	"github.com/iqtoolkit/iqtoolkit-analyzer/internal/metrics"
	"github.com/iqtoolkit/iqtoolkit-analyzer/internal/recommendations"
)

// Format identifies an output format.
type Format string

const (
	Text     Format = "text"
	JSON     Format = "json"
	Markdown Format = "markdown"
)

type jsonSummary struct {
	TotalEntries  int    `json:"total_entries"`
	ErrorCount    int    `json:"error_count"`
	SlowQueries   int    `json:"slow_queries"`
	AvgDuration   string `json:"avg_duration"`
	PeakErrorTime string `json:"peak_error_time,omitempty"`
}

type jsonRecommendation struct {
	Severity string `json:"severity"`
	Category string `json:"category"`
	Message  string `json:"message"`
}

type jsonSlowQuery struct {
	Timestamp string `json:"timestamp"`
	Duration  string `json:"duration"`
	Query     string `json:"query"`
}

type jsonReport struct {
	Summary           jsonSummary          `json:"summary"`
	SlowQueries       []jsonSlowQuery      `json:"slow_queries,omitempty"`
	Recommendations   []jsonRecommendation `json:"recommendations"`
	AIRecommendations string               `json:"ai_recommendations,omitempty"`
}

// maxSlowQueries limits how many slow queries are rendered in reports.
const maxSlowQueries = 10

func topSlowQueries(report *metrics.Report) []logparser.Entry {
	qs := make([]logparser.Entry, len(report.SlowQueries))
	copy(qs, report.SlowQueries)
	sort.Slice(qs, func(i, j int) bool { return qs[i].Duration > qs[j].Duration })
	if len(qs) > maxSlowQueries {
		qs = qs[:maxSlowQueries]
	}
	return qs
}

// Write renders the report, recommendations, and optional AI content to w in
// the given format. Unknown formats fall back to text.
func Write(w io.Writer, format Format, report *metrics.Report, recs []recommendations.Recommendation, aiContent string) error {
	switch format {
	case JSON:
		return writeJSON(w, report, recs, aiContent)
	case Markdown:
		return writeMarkdown(w, report, recs, aiContent)
	default:
		return writeText(w, report, recs, aiContent)
	}
}

func writeJSON(w io.Writer, report *metrics.Report, recs []recommendations.Recommendation, aiContent string) error {
	out := jsonReport{
		Summary: jsonSummary{
			TotalEntries: report.TotalEntries,
			ErrorCount:   report.ErrorCount,
			SlowQueries:  len(report.SlowQueries),
			AvgDuration:  report.AvgDuration.String(),
		},
		AIRecommendations: aiContent,
	}
	if !report.PeakErrorTime.IsZero() {
		out.Summary.PeakErrorTime = report.PeakErrorTime.Format(time.RFC3339)
	}
	for _, q := range topSlowQueries(report) {
		out.SlowQueries = append(out.SlowQueries, jsonSlowQuery{
			Timestamp: q.Timestamp.Format(time.RFC3339),
			Duration:  q.Duration.String(),
			Query:     q.Message,
		})
	}
	for _, r := range recs {
		out.Recommendations = append(out.Recommendations, jsonRecommendation{r.Severity, r.Category, r.Message})
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func writeMarkdown(w io.Writer, report *metrics.Report, recs []recommendations.Recommendation, aiContent string) error {
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
	if qs := topSlowQueries(report); len(qs) > 0 {
		fmt.Fprintf(w, "\n## Slowest Queries (top %d)\n\n", len(qs))
		fmt.Fprintf(w, "| Time | Duration | Query |\n|------|----------|-------|\n")
		for _, q := range qs {
			fmt.Fprintf(w, "| %s | %v | `%s` |\n", q.Timestamp.Format("2006-01-02 15:04:05"), q.Duration, q.Message)
		}
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
	return nil
}

func writeText(w io.Writer, report *metrics.Report, recs []recommendations.Recommendation, aiContent string) error {
	fmt.Fprintf(w, "=== Summary ===\n")
	fmt.Fprintf(w, "Total entries:    %d\n", report.TotalEntries)
	fmt.Fprintf(w, "Error count:      %d\n", report.ErrorCount)
	fmt.Fprintf(w, "Slow queries:     %d\n", len(report.SlowQueries))
	fmt.Fprintf(w, "Avg duration:     %v\n", report.AvgDuration)
	if !report.PeakErrorTime.IsZero() {
		fmt.Fprintf(w, "Peak error time:  %v\n", report.PeakErrorTime)
	}
	if qs := topSlowQueries(report); len(qs) > 0 {
		fmt.Fprintf(w, "\n=== Slowest Queries (top %d) ===\n", len(qs))
		for _, q := range qs {
			fmt.Fprintf(w, "[%s] %v  %s\n", q.Timestamp.Format("2006-01-02 15:04:05"), q.Duration, q.Message)
		}
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
	return nil
}
