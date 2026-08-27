package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func mockOutputServer(t *testing.T, outputs map[string]map[string]any) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/api/v2/outputs/")
		o, ok := outputs[id]
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "not found"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(o)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestOutputsGet_JSONOutput(t *testing.T) {
	srv := mockOutputServer(t, map[string]map[string]any{
		"office-speaker": {"outputId": "office-speaker", "displayName": "Office Speaker", "volume": 75, "muted": false, "available": true, "enabled": true},
	})

	res := runCLI(t, "get", "outputs/office-speaker", "--hub-url", srv.URL, "--json")

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	var decoded struct {
		OutputID    string `json:"outputId"`
		DisplayName string `json:"displayName"`
		Volume      int    `json:"volume"`
		Muted       bool   `json:"muted"`
		Available   bool   `json:"available"`
		Enabled     bool   `json:"enabled"`
	}
	if err := json.Unmarshal([]byte(res.stdout), &decoded); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\ngot: %s", err, res.stdout)
	}
	if decoded.OutputID != "office-speaker" || decoded.DisplayName != "Office Speaker" ||
		decoded.Volume != 75 || decoded.Muted != false || decoded.Available != true || decoded.Enabled != true {
		t.Errorf("unexpected decoded content: %+v", decoded)
	}

	yamlRes := runCLI(t, "get", "outputs/office-speaker", "--hub-url", srv.URL)
	if yamlRes.exitCode != 0 {
		t.Fatalf("yaml exit code = %d, want 0; stderr: %s", yamlRes.exitCode, yamlRes.stderr)
	}
	if !strings.Contains(yamlRes.stdout, "office-speaker") || !strings.Contains(yamlRes.stdout, "Office Speaker") {
		t.Errorf("expected yaml view to match same values, got:\n%s", yamlRes.stdout)
	}
}

func TestOutputsGet_AliasIdenticalToFullName(t *testing.T) {
	srv := mockOutputServer(t, map[string]map[string]any{
		"office-speaker": {"outputId": "office-speaker", "displayName": "Office Speaker", "volume": 75, "muted": false, "available": true, "enabled": true},
	})

	full := runCLI(t, "get", "outputs/office-speaker", "--hub-url", srv.URL)
	alias := runCLI(t, "get", "out/office-speaker", "--hub-url", srv.URL)

	if alias.exitCode != full.exitCode {
		t.Fatalf("alias exit code = %d, want %d (same as full name); stderr: %s", alias.exitCode, full.exitCode, alias.stderr)
	}
	if alias.stdout != full.stdout {
		t.Errorf("alias stdout = %q, want identical to full-name stdout %q", alias.stdout, full.stdout)
	}
}

func TestOutputsGet_NotFound(t *testing.T) {
	srv := mockOutputServer(t, map[string]map[string]any{})

	res := runCLI(t, "get", "outputs/missing-speaker", "--hub-url", srv.URL)

	if res.exitCode != 5 {
		t.Fatalf("exit code = %d, want 5; stderr: %s", res.exitCode, res.stderr)
	}
	if res.exitCode == 3 || res.exitCode == 4 {
		t.Errorf("not-found exit code must differ from hub-error(3)/network-error(4)")
	}
	lower := strings.ToLower(res.stderr)
	if !strings.Contains(lower, "not found") || !strings.Contains(res.stderr, "missing-speaker") {
		t.Errorf("expected a clear 'output not found' message naming the identifier, got:\n%s", res.stderr)
	}
	if res.stdout != "" {
		t.Errorf("expected empty stdout on failure, got:\n%s", res.stdout)
	}
}

func TestOutputsGet_SuccessYAML(t *testing.T) {
	srv := mockOutputServer(t, map[string]map[string]any{
		"office-speaker": {"outputId": "office-speaker", "displayName": "Office Speaker", "volume": 75, "muted": false, "available": true, "enabled": true},
		"garage-speaker": {"outputId": "garage-speaker", "displayName": "Garage Speaker", "volume": 0, "muted": false, "available": false, "enabled": false},
	})

	// Warm up the freshly built binary once, untimed: the first exec of a
	// newly written .exe on Windows can carry a one-time AV-scan/cold-start
	// cost unrelated to the command's actual latency (SC-001 measures the
	// command, not first-process-launch overhead).
	runCLI(t, "get", "outputs/office-speaker", "--hub-url", srv.URL)

	for _, id := range []string{"office-speaker", "garage-speaker"} {
		start := time.Now()
		res := runCLI(t, "get", "outputs/"+id, "--hub-url", srv.URL)
		elapsed := time.Since(start)

		if res.exitCode != 0 {
			t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
		}
		if elapsed > time.Second {
			t.Errorf("expected completion under 1s (SC-001), took %v", elapsed)
		}
		for _, field := range []string{"outputId", "displayName", "volume", "muted", "available", "enabled"} {
			if !strings.Contains(res.stdout, field) {
				t.Errorf("id=%s: expected field %q in stdout, got:\n%s", id, field, res.stdout)
			}
		}
	}
}
