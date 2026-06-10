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

func TestGenerateSettingsRules(t *testing.T) {
	cases := []struct {
		name     string
		setting  dbconn.Setting
		severity string
	}{
		{"autovacuum off", dbconn.Setting{Name: "autovacuum", Value: "off"}, "critical"},
		{"effective_cache_size default", dbconn.Setting{Name: "effective_cache_size", Value: "524288"}, "info"},
		{"max_connections high", dbconn.Setting{Name: "max_connections", Value: "1000"}, "warning"},
		{"checkpoint_completion_target low", dbconn.Setting{Name: "checkpoint_completion_target", Value: "0.5"}, "info"},
	}
	for _, c := range cases {
		recs := Generate(&metrics.Report{Settings: []dbconn.Setting{c.setting}})
		if len(recs) != 1 {
			t.Errorf("%s: got %d recs, want 1", c.name, len(recs))
			continue
		}
		if recs[0].Severity != c.severity {
			t.Errorf("%s: severity = %q, want %q", c.name, recs[0].Severity, c.severity)
		}
	}

	// Healthy values produce no recommendations.
	healthy := []dbconn.Setting{
		{Name: "autovacuum", Value: "on"},
		{Name: "effective_cache_size", Value: "1048576"},
		{Name: "max_connections", Value: "100"},
		{Name: "checkpoint_completion_target", Value: "0.9"},
	}
	if recs := Generate(&metrics.Report{Settings: healthy}); len(recs) != 0 {
		t.Errorf("healthy settings: got %d recs, want 0", len(recs))
	}
}

func TestGenerateTableRules(t *testing.T) {
	report := &metrics.Report{
		Tables: []dbconn.TableStat{
			{Schema: "public", Table: "bloated", NDeadTup: 50000, NLiveTup: 100000},
			{Schema: "public", Table: "unindexed", SeqScan: 5000, IdxScan: 0, NLiveTup: 50000},
			{Schema: "public", Table: "healthy", SeqScan: 10, IdxScan: 9000, NDeadTup: 100, NLiveTup: 100000},
		},
	}
	recs := Generate(report)
	if len(recs) != 2 {
		t.Fatalf("got %d recs, want 2: %+v", len(recs), recs)
	}
	var hasVacuum, hasIndex bool
	for _, r := range recs {
		if r.Category == "maintenance" {
			hasVacuum = true
		}
		if r.Category == "performance" {
			hasIndex = true
		}
	}
	if !hasVacuum {
		t.Error("expected dead-tuple vacuum recommendation")
	}
	if !hasIndex {
		t.Error("expected missing-index recommendation")
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
