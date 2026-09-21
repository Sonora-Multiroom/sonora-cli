package unit

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"sonora-cli/internal/cli/tts"
)

func TestRunGetCache_Help(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := tts.RunGetCache([]string{"--help"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Flags:") {
		t.Errorf("expected a Flags: section, got stdout:\n%s", stdout.String())
	}
}

func TestRunGetCache_Success_YAML(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"totalEntries": 3, "totalSizeBytes": 100, "maxSizeBytes": 1000,
			"entriesByProvider": map[string]any{"openai": 3},
		})
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := tts.RunGetCache([]string{"--hub-url", srv.URL}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", code, stderr.String())
	}
	for _, want := range []string{"totalEntries: 3", "totalSizeBytes: 100", "maxSizeBytes: 1000", "openai"} {
		if !strings.Contains(stdout.String(), want) {
			t.Errorf("expected stdout to contain %q, got:\n%s", want, stdout.String())
		}
	}
}

func TestRunGetCache_Success_JSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"totalEntries": 0, "totalSizeBytes": 0, "maxSizeBytes": 1000})
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := tts.RunGetCache([]string{"--hub-url", srv.URL, "--json"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", code, stderr.String())
	}
	var decoded map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &decoded); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\ngot: %s", err, stdout.String())
	}
}

func TestRunGetCache_ExtraPositional(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := tts.RunGetCache([]string{"extra"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
}

func TestRunClearCache_Help(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := tts.RunClearCache([]string{"--help"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "--provider") {
		t.Errorf("expected help to list --provider, got:\n%s", stdout.String())
	}
}

func TestRunClearCache_NoProvider(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := tts.RunClearCache([]string{"--hub-url", srv.URL}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", code, stderr.String())
	}
	if gotQuery != "" {
		t.Errorf("expected no query string, got: %q", gotQuery)
	}
	if stdout.String() != "cleared: all\n" {
		t.Errorf("stdout = %q, want %q", stdout.String(), "cleared: all\n")
	}
}

func TestRunClearCache_WithProvider(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query().Get("providerName")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := tts.RunClearCache([]string{"--provider", "openai", "--hub-url", srv.URL}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", code, stderr.String())
	}
	if gotQuery != "openai" {
		t.Errorf("got providerName=%q, want openai", gotQuery)
	}
	want := "cleared: provider\nprovider: \"openai\"\n"
	if stdout.String() != want {
		t.Errorf("stdout = %q, want %q", stdout.String(), want)
	}
}

func TestRunClearCache_EmptyProvider(t *testing.T) {
	for _, val := range []string{"", "  "} {
		t.Run(val, func(t *testing.T) {
			srv, count := countingSpeakHub(t)
			var stdout, stderr bytes.Buffer
			code := tts.RunClearCache([]string{"--provider", val, "--hub-url", srv.URL}, &stdout, &stderr)

			if code != 2 {
				t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
			}
			if !strings.Contains(stderr.String(), "--provider must not be empty") {
				t.Errorf("expected the empty-provider message, got: %s", stderr.String())
			}
			if got := atomic.LoadInt32(count); got != 0 {
				t.Errorf("expected zero requests, got %d", got)
			}
		})
	}
}

func TestRunClearCache_ExtraPositional(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := tts.RunClearCache([]string{"extra"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
}

func TestRunClearCache_JSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := tts.RunClearCache([]string{"--hub-url", srv.URL, "--json"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", code, stderr.String())
	}
	want := `{"cleared":"all"}` + "\n"
	if stdout.String() != want {
		t.Errorf("stdout = %q, want %q", stdout.String(), want)
	}
}

func TestRunGetCache_404_UsesReportError(t *testing.T) {
	var extCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/tts/cache/stats":
			w.WriteHeader(http.StatusNotFound)
		case "/api/v2/extensions":
			atomic.AddInt32(&extCount, 1)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"extensions": []map[string]any{{"id": "tts", "status": "REJECTED", "rejectionReason": "bad jar"}}})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	var stdout, stderr bytes.Buffer
	code := tts.RunGetCache([]string{"--hub-url", srv.URL}, &stdout, &stderr)

	if code != 13 {
		t.Fatalf("exit code = %d, want 13; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "bad jar") {
		t.Errorf("expected stderr to contain the rejection reason, got: %s", stderr.String())
	}
	if atomic.LoadInt32(&extCount) != 1 {
		t.Errorf("expected exactly 1 inventory request, got %d", extCount)
	}
}
