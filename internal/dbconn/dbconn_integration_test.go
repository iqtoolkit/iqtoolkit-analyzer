package dbconn

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"
)

// Integration tests run only when TEST_DATABASE_URL is set, e.g.:
//
//	TEST_DATABASE_URL=postgres://postgres:postgres@localhost:5432/postgres go test ./internal/dbconn/
func testConn(t *testing.T) *Conn {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping integration test")
	}
	ctx, cancel := context.WithTimeoutCause(t.Context(), 15*time.Second, errors.New("test: connection timed out"))
	t.Cleanup(cancel)
	conn, err := Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("connecting: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close(context.Background()) })
	return conn
}

func TestVersion(t *testing.T) {
	conn := testConn(t)
	v, err := conn.Version(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(v, "PostgreSQL") {
		t.Errorf("version = %q", v)
	}
}

func TestSettings(t *testing.T) {
	conn := testConn(t)
	settings, err := conn.Settings(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(settings) == 0 {
		t.Fatal("no settings returned")
	}
	var found bool
	for _, s := range settings {
		if s.Name == "shared_buffers" {
			found = true
			if s.Value == "" {
				t.Error("shared_buffers has empty value")
			}
		}
	}
	if !found {
		t.Error("shared_buffers not in settings")
	}
}

func TestExtensions(t *testing.T) {
	conn := testConn(t)
	exts, err := conn.Extensions(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(exts) == 0 {
		t.Fatal("no extensions returned")
	}
}

func TestCheckExtension(t *testing.T) {
	conn := testConn(t)
	// plpgsql is installed by default in every PostgreSQL database.
	status, err := conn.CheckExtension(t.Context(), "plpgsql")
	if err != nil {
		t.Fatal(err)
	}
	if !status.Installed {
		t.Error("plpgsql should be installed")
	}

	status, err = conn.CheckExtension(t.Context(), "definitely_not_a_real_extension")
	if err != nil {
		t.Fatal(err)
	}
	if status.Installed || status.Available {
		t.Error("nonexistent extension should be neither installed nor available")
	}
}

func TestStatUserTables(t *testing.T) {
	conn := testConn(t)
	// Should not error even with no user tables.
	if _, err := conn.StatUserTables(t.Context()); err != nil {
		t.Fatal(err)
	}
}

func TestStatUserIndexes(t *testing.T) {
	conn := testConn(t)
	if _, err := conn.StatUserIndexes(t.Context()); err != nil {
		t.Fatal(err)
	}
}

func TestQueryTimeout(t *testing.T) {
	conn := testConn(t)
	conn.QueryTimeout = 1 * time.Nanosecond
	if _, err := conn.Settings(t.Context()); err == nil {
		t.Error("expected timeout error with 1ns QueryTimeout")
	}
}
