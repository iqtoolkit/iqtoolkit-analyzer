package metrics

import (
	"testing"
	"time"

	"github.com/iqtoolkit/iqtoolkit-analyzer/internal/dbconn"
	"github.com/iqtoolkit/iqtoolkit-analyzer/internal/logparser"
)

func TestAnalyze(t *testing.T) {
	base := time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC)
	entries := []logparser.Entry{
		{Timestamp: base, Level: "ERROR", Message: "fail"},
		{Timestamp: base.Add(time.Minute), Level: "LOG", Duration: 200 * time.Millisecond},
		{Timestamp: base.Add(2 * time.Minute), Level: "LOG", Duration: 800 * time.Millisecond},
		{Timestamp: base.Add(3 * time.Minute), Level: "FATAL", Message: "crash"},
	}
	settings := []dbconn.Setting{{Name: "shared_buffers", Value: "16384", Source: "configuration file"}}

	report := Analyze(entries, settings, 500*time.Millisecond)

	if report.TotalEntries != 4 {
		t.Errorf("TotalEntries = %d, want 4", report.TotalEntries)
	}
	if report.ErrorCount != 2 {
		t.Errorf("ErrorCount = %d, want 2", report.ErrorCount)
	}
	if len(report.SlowQueries) != 1 {
		t.Errorf("SlowQueries = %d, want 1", len(report.SlowQueries))
	}
	if report.AvgDuration != 500*time.Millisecond {
		t.Errorf("AvgDuration = %v, want 500ms", report.AvgDuration)
	}
}

func TestAnalyzeEmpty(t *testing.T) {
	report := Analyze(nil, nil, time.Second)
	if report.TotalEntries != 0 {
		t.Errorf("TotalEntries = %d, want 0", report.TotalEntries)
	}
	if report.AvgDuration != 0 {
		t.Errorf("AvgDuration = %v, want 0", report.AvgDuration)
	}
}
