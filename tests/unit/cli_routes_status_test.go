package unit

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"sonora-cli/internal/cli/routes"
)

// The hub's status enum is upper-case, but `--status active` is what the
// docs show and what users type, so the CLI upper-cases before sending.
func TestRoutesRunList_StatusIsUpperCasedOnTheWire(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.URL.Query().Get("status")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
	}))
	defer srv.Close()

	for _, typed := range []string{"active", "Active", "ACTIVE"} {
		var stdout, stderr bytes.Buffer
		code := routes.RunList([]string{"--status", typed, "--hub-url", srv.URL}, &stdout, &stderr)

		if code != 0 {
			t.Fatalf("--status %s: exit code = %d, want 0; stderr: %s", typed, code, stderr.String())
		}
		if got != "ACTIVE" {
			t.Errorf("--status %s: hub received status=%q, want ACTIVE", typed, got)
		}
	}
}

func TestRoutesRunList_UnknownStatusRejectedWithoutCallingHub(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("hub was called for an invalid status: %s", r.URL)
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := routes.RunList([]string{"--status", "playing", "--hub-url", srv.URL}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), `unknown status "playing"`) {
		t.Errorf("expected the rejected value to be named, got stderr:\n%s", stderr.String())
	}
	for _, want := range []string{"STARTING", "ACTIVE", "STOPPING", "STOPPED", "FAILED"} {
		if !strings.Contains(stderr.String(), want) {
			t.Errorf("expected %s among the listed valid statuses, got stderr:\n%s", want, stderr.String())
		}
	}
}

func TestRoutesRunList_NoStatusSendsNoFilter(t *testing.T) {
	var raw string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := routes.RunList([]string{"--hub-url", srv.URL}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", code, stderr.String())
	}
	if strings.Contains(raw, "status") {
		t.Errorf("expected no status parameter, got query %q", raw)
	}
}
