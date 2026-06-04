package logparser

import (
	"strings"
	"testing"
	"time"
)

func TestParseLine(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		wantOK  bool
		level   string
		dur     time.Duration
		message string
	}{
		{
			name:    "standard error",
			line:    "2024-03-15 10:30:45.123 UTC [12345] ERROR: connection refused",
			wantOK:  true,
			level:   "ERROR",
			message: "connection refused",
		},
		{
			name:    "log with duration",
			line:    "2024-03-15 10:30:45.123 UTC [99] LOG: duration: 250.5 ms  statement: SELECT 1",
			wantOK:  true,
			level:   "LOG",
			dur:     250500 * time.Microsecond,
			message: "duration: 250.5 ms  statement: SELECT 1",
		},
		{
			name:   "no fractional seconds",
			line:   "2024-03-15 10:30:45 UTC [1] LOG: started",
			wantOK: true,
			level:  "LOG",
		},
		{
			name:   "invalid line",
			line:   "some random text",
			wantOK: false,
		},
		{
			name:   "empty line",
			line:   "",
			wantOK: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, ok := parseLine(tt.line)
			if ok != tt.wantOK {
				t.Fatalf("parseLine() ok = %v, want %v", ok, tt.wantOK)
			}
			if !ok {
				return
			}
			if e.Level != tt.level {
				t.Errorf("Level = %q, want %q", e.Level, tt.level)
			}
			if tt.dur != 0 && e.Duration != tt.dur {
				t.Errorf("Duration = %v, want %v", e.Duration, tt.dur)
			}
			if tt.message != "" && e.Message != tt.message {
				t.Errorf("Message = %q, want %q", e.Message, tt.message)
			}
		})
	}
}

func TestParse(t *testing.T) {
	input := `2024-03-15 10:30:45.123 UTC [1] ERROR: something failed
2024-03-15 10:30:46.000 UTC [2] LOG: duration: 100.0 ms  statement: SELECT 1
not a log line
2024-03-15 10:30:47.000 UTC [3] FATAL: shutdown`

	entries, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 {
		t.Fatalf("got %d entries, want 3", len(entries))
	}
	if entries[0].Level != "ERROR" {
		t.Errorf("entries[0].Level = %q, want ERROR", entries[0].Level)
	}
	if entries[1].Duration != 100*time.Millisecond {
		t.Errorf("entries[1].Duration = %v, want 100ms", entries[1].Duration)
	}
	if entries[2].Level != "FATAL" {
		t.Errorf("entries[2].Level = %q, want FATAL", entries[2].Level)
	}
}
