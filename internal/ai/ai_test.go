package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/iqtoolkit/iqtoolkit-analyzer/internal/logparser"
	"github.com/iqtoolkit/iqtoolkit-analyzer/internal/metrics"
)

func TestDefaultModel(t *testing.T) {
	for _, p := range []Provider{OpenAI, Anthropic, Gemini, Kiro} {
		if DefaultModel(p) == "" {
			t.Errorf("DefaultModel(%s) is empty", p)
		}
	}
	if DefaultModel(Provider("nope")) != "" {
		t.Error("unknown provider should return empty model")
	}
}

func TestDefaultModelEnvOverride(t *testing.T) {
	t.Setenv("IQTOOLKIT_AI_MODEL", "my-model")
	if got := DefaultModel(OpenAI); got != "my-model" {
		t.Errorf("env override ignored, got %q", got)
	}
}

func TestIsRetryable(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"429", &HTTPError{Provider: OpenAI, StatusCode: 429}, true},
		{"503", &HTTPError{Provider: Anthropic, StatusCode: 503}, true},
		{"500", &HTTPError{Provider: Gemini, StatusCode: 500}, true},
		{"401", &HTTPError{Provider: OpenAI, StatusCode: 401}, false},
		{"404", &HTTPError{Provider: OpenAI, StatusCode: 404}, false},
		{"connection refused", syscall.ECONNREFUSED, true},
		{"connection reset", syscall.ECONNRESET, true},
		{"unexpected EOF", io.ErrUnexpectedEOF, true},
		{"deadline exceeded", context.DeadlineExceeded, true},
		{"wrapped 429", fmt.Errorf("call failed: %w", &HTTPError{Provider: OpenAI, StatusCode: 429}), true},
		{"plain error mentioning 429", errors.New("query returned 429 rows"), false},
		{"unsupported provider", errors.New("ai: unsupported provider"), false},
	}
	for _, c := range cases {
		if got := isRetryable(c.err); got != c.want {
			t.Errorf("%s: isRetryable = %v, want %v", c.name, got, c.want)
		}
	}
}

// roundTripperFunc lets us redirect provider URLs to a test server.
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func redirectClient(ts *httptest.Server) *http.Client {
	return &http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			r.URL.Scheme = "http"
			r.URL.Host = strings.TrimPrefix(ts.URL, "http://")
			return http.DefaultTransport.RoundTrip(r)
		}),
	}
}

func TestCompleteOpenAI(t *testing.T) {
	var gotAuth string
	var gotBody map[string]any
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"hello"}}]}`))
	}))
	defer ts.Close()

	c := NewClient(OpenAI, "sk-test")
	c.HTTPClient = redirectClient(ts)
	resp, err := c.Complete(t.Context(), Request{
		Model:    "gpt-4o",
		System:   "be brief",
		Messages: []Message{{Role: "user", Content: "hi"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Content != "hello" {
		t.Errorf("content = %q", resp.Content)
	}
	if gotAuth != "Bearer sk-test" {
		t.Errorf("auth header = %q", gotAuth)
	}
	msgs := gotBody["messages"].([]any)
	if len(msgs) != 2 {
		t.Fatalf("expected system+user messages, got %d", len(msgs))
	}
	if msgs[0].(map[string]any)["role"] != "system" {
		t.Error("system prompt should be prepended as first message")
	}
}

func TestCompleteAnthropic(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "ak-test" {
			t.Errorf("missing x-api-key header")
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["system"] != "sys" {
			t.Errorf("system = %v", body["system"])
		}
		_, _ = w.Write([]byte(`{"content":[{"text":"claude says"}]}`))
	}))
	defer ts.Close()

	c := NewClient(Anthropic, "ak-test")
	c.HTTPClient = redirectClient(ts)
	resp, err := c.Complete(t.Context(), Request{
		Model:    "claude-sonnet-4-5",
		System:   "sys",
		Messages: []Message{{Role: "user", Content: "hi"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Content != "claude says" {
		t.Errorf("content = %q", resp.Content)
	}
}

func TestCompleteGemini(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"gemini says"}]}}]}`))
	}))
	defer ts.Close()

	c := NewClient(Gemini, "gk-test")
	c.HTTPClient = redirectClient(ts)
	resp, err := c.Complete(t.Context(), Request{
		Model:    "gemini-2.5-pro",
		Messages: []Message{{Role: "user", Content: "hi"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Content != "gemini says" {
		t.Errorf("content = %q", resp.Content)
	}
}

func TestCompleteRetriesOn429(t *testing.T) {
	var calls int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"ok"}}]}`))
	}))
	defer ts.Close()

	c := NewClient(OpenAI, "k")
	c.HTTPClient = redirectClient(ts)
	c.Timeout = 5 * time.Second
	resp, err := c.Complete(t.Context(), Request{Model: "m", Messages: []Message{{Role: "user", Content: "x"}}})
	if err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Errorf("expected 2 calls (1 retry), got %d", calls)
	}
	if resp.Content != "ok" {
		t.Errorf("content = %q", resp.Content)
	}
}

func TestCompleteDoesNotRetryOn401(t *testing.T) {
	var calls int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()

	c := NewClient(OpenAI, "bad")
	c.HTTPClient = redirectClient(ts)
	_, err := c.Complete(t.Context(), Request{Model: "m", Messages: []Message{{Role: "user", Content: "x"}}})
	if err == nil {
		t.Fatal("expected error")
	}
	if calls != 1 {
		t.Errorf("401 should not retry, got %d calls", calls)
	}
}

func TestCompleteUnsupportedProvider(t *testing.T) {
	c := NewClient(Provider("bogus"), "")
	_, err := c.Complete(t.Context(), Request{Model: "m"})
	if err == nil || !strings.Contains(err.Error(), "unsupported provider") {
		t.Errorf("err = %v", err)
	}
}

func TestBuildPrompt(t *testing.T) {
	rep := &metrics.Report{
		TotalEntries: 50,
		ErrorCount:   3,
		AvgDuration:  120 * time.Millisecond,
		SlowQueries: []logparser.Entry{
			{Timestamp: time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC), Message: "SELECT 1", Duration: 2 * time.Second},
		},
	}
	p := BuildPrompt(rep)
	for _, want := range []string{
		"## PostgreSQL Health Metrics",
		"Total log entries: 50",
		"Error count: 3",
		"## Slow Queries",
		"SELECT 1",
	} {
		if !strings.Contains(p, want) {
			t.Errorf("prompt missing %q", want)
		}
	}
}

func TestTruncate(t *testing.T) {
	if got := truncate("hello", 10); got != "hello" {
		t.Errorf("got %q", got)
	}
	if got := truncate("hello world", 5); got != "hello..." {
		t.Errorf("got %q", got)
	}
}
