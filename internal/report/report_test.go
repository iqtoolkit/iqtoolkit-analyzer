package report

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/iqtoolkit/iqtoolkit-analyzer/internal/dbconn"
)

func TestGenerate(t *testing.T) {
	var buf bytes.Buffer
	err := Generate(&buf, Data{
		Version: "PostgreSQL 17.2",
		Settings: []dbconn.Setting{
			{Name: "shared_buffers", Value: "16384", Source: "default"},
		},
		Extensions: []dbconn.Extension{
			{Name: "pg_stat_statements", DefaultVersion: "1.11", InstalledVersion: "1.11"},
			{Name: "pg_buffercache", DefaultVersion: "1.5", InstalledVersion: ""},
		},
		GeneratedAt: time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	html := buf.String()
	for _, want := range []string{
		"<!DOCTYPE html>",
		"PostgreSQL 17.2",
		"shared_buffers",
		"16384",
		"pg_stat_statements",
		`<span class="installed">1.11</span>`,
		`<span class="not-installed">—</span>`,
		"2026-06-01 12:00:00",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("report missing %q", want)
		}
	}
}

func TestGenerateEscapesHTML(t *testing.T) {
	var buf bytes.Buffer
	err := Generate(&buf, Data{
		Version:     "<script>alert(1)</script>",
		GeneratedAt: time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(buf.String(), "<script>alert(1)</script>") {
		t.Error("HTML not escaped in version field")
	}
}
