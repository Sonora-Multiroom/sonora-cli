package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// mockInfoHub answers GET /api/v2/info with the given status and body.
func mockInfoHub(t *testing.T, status int, body string) (*httptest.Server, *[]string) {
	t.Helper()
	var mu sync.Mutex
	var seen []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		seen = append(seen, r.Method+" "+r.URL.Path)
		mu.Unlock()
		if r.URL.Path != "/api/v2/info" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv, &seen
}

const integrationInfoBody = `{"product":"multiroom","version":"0.1.33","apiVersion":"0.1.33","startedAt":"2026-11-06T10:06:59.512Z","uptimeSeconds":16981}`

func TestGetInfo_YAML(t *testing.T) {
	srv, seen := mockInfoHub(t, 200, integrationInfoBody)

	res := runCLI(t, "get", "info", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	want := "product: \"multiroom\"\nversion: \"0.1.33\"\napiVersion: \"0.1.33\"\n" +
		"startedAt: \"2026-11-06T10:06:59.512Z\"\nuptimeSeconds: 16981\n"
	if res.stdout != want {
		t.Errorf("stdout:\n%s\nwant:\n%s", res.stdout, want)
	}
	if len(*seen) != 1 || (*seen)[0] != "GET /api/v2/info" {
		t.Errorf("requests = %v, want one GET /api/v2/info", *seen)
	}
}

func TestGetInfo_JSON(t *testing.T) {
	srv, _ := mockInfoHub(t, 200, integrationInfoBody)

	res := runCLI(t, "get", "info", "--hub-url", srv.URL, "--json")

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(res.stdout), &m); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\n%s", err, res.stdout)
	}
	for _, k := range []string{"product", "version", "apiVersion", "startedAt", "uptimeSeconds"} {
		if _, ok := m[k]; !ok {
			t.Errorf("missing %q in %s", k, res.stdout)
		}
	}
}

func TestGetInfo_404_OldHub(t *testing.T) {
	srv, _ := mockInfoHub(t, 404, ``)

	res := runCLI(t, "get", "info", "--hub-url", srv.URL)

	if res.exitCode != 3 {
		t.Fatalf("exit code = %d, want 3; stderr: %s", res.exitCode, res.stderr)
	}
	want := "error: this hub does not offer hub info (needs hub 0.1.33 or later) (hub URL: " + srv.URL + ")\n"
	if res.stderr != want {
		t.Errorf("stderr = %q, want %q", res.stderr, want)
	}
	if res.stdout != "" {
		t.Errorf("expected empty stdout, got %q", res.stdout)
	}
}

func TestGetInfo_UnreachableHub(t *testing.T) {
	res := runCLI(t, "get", "info", "--hub-url", "http://127.0.0.1:1")
	if res.exitCode != 4 {
		t.Fatalf("exit code = %d, want 4; stderr: %s", res.exitCode, res.stderr)
	}
	if res.stdout != "" {
		t.Errorf("expected empty stdout, got %q", res.stdout)
	}
}

func TestGetInfo_MalformedBody(t *testing.T) {
	srv, _ := mockInfoHub(t, 200, `{"version":"1"}`)

	res := runCLI(t, "get", "info", "--hub-url", srv.URL)

	if res.exitCode != 3 {
		t.Fatalf("exit code = %d, want 3; stderr: %s", res.exitCode, res.stderr)
	}
	if !strings.Contains(res.stderr, "hub returned an unexpected or malformed response") {
		t.Errorf("stderr = %q", res.stderr)
	}
}

func TestGetInfo_Verbose_KeepsCause(t *testing.T) {
	srv, _ := mockInfoHub(t, 200, `not json`)

	res := runCLI(t, "get", "info", "--hub-url", srv.URL, "--verbose")

	if res.exitCode != 3 {
		t.Fatalf("exit code = %d, want 3; stderr: %s", res.exitCode, res.stderr)
	}
	if !strings.Contains(res.stderr, "detail: ") {
		t.Errorf("expected a detail line, got %q", res.stderr)
	}
}

func TestGetInfo_UsageErrors(t *testing.T) {
	srv, seen := mockInfoHub(t, 200, integrationInfoBody)

	res := runCLI(t, "get", "info", "extra", "--hub-url", srv.URL)
	if res.exitCode != 2 {
		t.Errorf("get info extra: exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
	res = runCLI(t, "list", "info", "--hub-url", srv.URL)
	if res.exitCode != 2 {
		t.Errorf("list info: exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
	if !strings.Contains(res.stderr, "sonora get info") {
		t.Errorf("list info should name 'sonora get info', got %q", res.stderr)
	}
	if len(*seen) != 0 {
		t.Errorf("usage errors must send no request, got %v", *seen)
	}
}

func TestGetInfo_Help(t *testing.T) {
	res := runCLI(t, "get", "info", "--help")
	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if !strings.Contains(res.stdout, "usage: sonora get info") {
		t.Errorf("stdout = %q", res.stdout)
	}
}
