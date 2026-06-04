package recommendations

import (
	"testing"
	"time"

	"github.com/iqtoolkit/iqtoolkit-analyzer/internal/dbconn"
	"github.com/iqtoolkit/iqtoolkit-analyzer/internal/logparser"
	"github.com/iqtoolkit/iqtoolkit-analyzer/internal/metrics"
)

func TestGeneratePerformanceWarnings(t *testing.T) {
	report := &metrics.Report{
		AvgDuration: 600 * time.Millisecond,
		SlowQueries: make([]logparser.Entry, 15),
		ErrorCount:  5,
	}
	recs := Generate(report)

	var hasAvgDur, hasSlowQuery bool
	for _, r := range recs {
		if r.Category == "performance" && r.Severity == "warning" {
			hasAvgDur = true
		}
		if r.Category == "performance" && r.Severity == "critical" {
			hasSlowQuery = true
		}
	}
	if !hasAvgDur {
		t.Error("expected avg duration warning")
	}
	if !hasSlowQuery {
		t.Error("expected slow query critical recommendation")
	}
}

func TestGenerateConfigRecommendations(t *testing.T) {
	report := &metrics.Report{
		Settings: []dbconn.Setting{
			{Name: "shared_buffers", Value: "16384"},
			{Name: "work_mem", Value: "4096"},
			{Name: "log_min_duration_statement", Value: "-1"},
		},
	}
	recs := Generate(report)
	if len(recs) != 3 {
		t.Fatalf("got %d recommendations, want 3", len(recs))
	}
	for _, r := range recs {
		if r.Category != "configuration" {
			t.Errorf("expected configuration category, got %q", r.Category)
		}
	}
}

func TestGenerateNoRecommendations(t *testing.T) {
	report := &metrics.Report{
		AvgDuration: 50 * time.Millisecond,
		ErrorCount:  2,
		Settings:    []dbconn.Setting{{Name: "shared_buffers", Value: "262144"}},
	}
	recs := Generate(report)
	if len(recs) != 0 {
		t.Errorf("expected no recommendations, got %d", len(recs))
	}
}
