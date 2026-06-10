package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/iqtoolkit/iqtoolkit-analyzer/internal/metrics"
	"github.com/iqtoolkit/iqtoolkit-analyzer/internal/recommendations"
)

func sampleReport() *metrics.Report {
	return &metrics.Report{
		TotalEntries:  100,
		ErrorCount:    5,
		AvgDuration:   250 * time.Millisecond,
		PeakErrorTime: time.Date(2026, 6, 1, 14, 0, 0, 0, time.UTC),
	}
}

func sampleRecs() []recommendations.Recommendation {
	return []recommendations.Recommendation{
		{Severity: "warning", Category: "performance", Message: "slow stuff"},
		{Severity: "info", Category: "configuration", Message: "tune work_mem"},
	}
}

func TestWriteJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := Write(&buf, JSON, sampleReport(), sampleRecs(), "ai says hi"); err != nil {
		t.Fatal(err)
	}
	var out struct {
		Summary struct {
			TotalEntries  int    `json:"total_entries"`
			ErrorCount    int    `json:"error_count"`
			AvgDuration   string `json:"avg_duration"`
			PeakErrorTime string `json:"peak_error_time"`
		} `json:"summary"`
		Recommendations []struct {
			Severity string `json:"severity"`
		} `json:"recommendations"`
		AIRecommendations string `json:"ai_recommendations"`
	}
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if out.Summary.TotalEntries != 100 || out.Summary.ErrorCount != 5 {
		t.Errorf("summary mismatch: %+v", out.Summary)
	}
	if out.Summary.AvgDuration != "250ms" {
		t.Errorf("avg_duration = %q, want 250ms", out.Summary.AvgDuration)
	}
	if out.Summary.PeakErrorTime == "" {
		t.Error("peak_error_time missing")
	}
	if len(out.Recommendations) != 2 {
		t.Errorf("got %d recommendations, want 2", len(out.Recommendations))
	}
	if out.AIRecommendations != "ai says hi" {
		t.Errorf("ai_recommendations = %q", out.AIRecommendations)
	}
}

func TestWriteJSONOmitsEmptyFields(t *testing.T) {
	var buf bytes.Buffer
	rep := sampleReport()
	rep.PeakErrorTime = time.Time{}
	if err := Write(&buf, JSON, rep, nil, ""); err != nil {
		t.Fatal(err)
	}
	s := buf.String()
	if strings.Contains(s, "peak_error_time") {
		t.Error("peak_error_time should be omitted when zero")
	}
	if strings.Contains(s, "ai_recommendations") {
		t.Error("ai_recommendations should be omitted when empty")
	}
}

func TestWriteMarkdown(t *testing.T) {
	var buf bytes.Buffer
	if err := Write(&buf, Markdown, sampleReport(), sampleRecs(), "ai content"); err != nil {
		t.Fatal(err)
	}
	s := buf.String()
	for _, want := range []string{
		"# PostgreSQL Analysis Report",
		"| Total entries | 100 |",
		"| Error count | 5 |",
		"## Recommendations",
		"- **[warning]** (performance) slow stuff",
		"## AI-Enhanced Recommendations",
		"ai content",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("markdown missing %q", want)
		}
	}
}

func TestWriteText(t *testing.T) {
	var buf bytes.Buffer
	if err := Write(&buf, Text, sampleReport(), sampleRecs(), ""); err != nil {
		t.Fatal(err)
	}
	s := buf.String()
	for _, want := range []string{
		"=== Summary ===",
		"Total entries:    100",
		"[warning][performance] slow stuff",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("text missing %q", want)
		}
	}
	if strings.Contains(s, "AI-Enhanced") {
		t.Error("AI section should be absent when aiContent is empty")
	}
}

func TestWriteTextNoRecommendations(t *testing.T) {
	var buf bytes.Buffer
	if err := Write(&buf, Text, sampleReport(), nil, ""); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "No recommendations") {
		t.Error("expected 'No recommendations' message")
	}
}

func TestWriteUnknownFormatFallsBackToText(t *testing.T) {
	var buf bytes.Buffer
	if err := Write(&buf, Format("bogus"), sampleReport(), nil, ""); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "=== Summary ===") {
		t.Error("unknown format should fall back to text")
	}
}
