package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func mockInputServer(t *testing.T, inputsByID map[string]map[string]any) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/api/v2/inputs/")
		i, ok := inputsByID[id]
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "not found"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(i)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestInputsGet_SuccessYAML(t *testing.T) {
	srv := mockInputServer(t, map[string]map[string]any{
		"spotify-1": {"inputId": "spotify-1", "displayName": "Spotify Stream", "uri": "u1", "enabled": true, "autoRemove": false, "source": "STATIC", "createdAt": nil, "pauseable": true},
		"aux-1":     {"inputId": "aux-1", "displayName": "Aux In", "uri": "u2", "enabled": false, "autoRemove": true, "source": "EPHEMERAL", "createdAt": "2026-06-22T14:30:00Z", "pauseable": false},
	})

	// Warm up the freshly built binary once, untimed (see
	// TestOutputsGet_SuccessYAML's comment for why).
	runCLI(t, "inputs", "get", "spotify-1", "--hub-url", srv.URL)

	for _, id := range []string{"spotify-1", "aux-1"} {
		start := time.Now()
		res := runCLI(t, "inputs", "get", id, "--hub-url", srv.URL)
		elapsed := time.Since(start)

		if res.exitCode != 0 {
			t.Fatalf("id=%s: exit code = %d, want 0; stderr: %s", id, res.exitCode, res.stderr)
		}
		if elapsed > time.Second {
			t.Errorf("id=%s: expected completion under 1s (SC-002), took %v", id, elapsed)
		}
		for _, field := range []string{"inputId", "displayName", "uri", "source", "enabled", "autoRemove", "pauseable", "createdAt"} {
			if !strings.Contains(res.stdout, field) {
				t.Errorf("id=%s: expected field %q in stdout, got:\n%s", id, field, res.stdout)
			}
		}
	}
}

func TestInputsGet_JSONOutput(t *testing.T) {
	srv := mockInputServer(t, map[string]map[string]any{
		"spotify-1": {"inputId": "spotify-1", "displayName": "Spotify Stream", "uri": "u1", "enabled": true, "autoRemove": false, "source": "STATIC", "createdAt": nil, "pauseable": true},
	})

	res := runCLI(t, "inputs", "get", "spotify-1", "--hub-url", srv.URL, "--json")

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	var decoded struct {
		InputID     string  `json:"inputId"`
		DisplayName string  `json:"displayName"`
		URI         string  `json:"uri"`
		Enabled     bool    `json:"enabled"`
		AutoRemove  bool    `json:"autoRemove"`
		Source      string  `json:"source"`
		CreatedAt   *string `json:"createdAt"`
		Pauseable   bool    `json:"pauseable"`
	}
	if err := json.Unmarshal([]byte(res.stdout), &decoded); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\ngot: %s", err, res.stdout)
	}
	if decoded.InputID != "spotify-1" || decoded.CreatedAt != nil {
		t.Errorf("unexpected decoded content: %+v", decoded)
	}
}

func TestInputsGet_NotFound(t *testing.T) {
	srv := mockInputServer(t, map[string]map[string]any{})

	res := runCLI(t, "inputs", "get", "missing-input", "--hub-url", srv.URL)

	if res.exitCode != 5 {
		t.Fatalf("exit code = %d, want 5; stderr: %s", res.exitCode, res.stderr)
	}
	if res.exitCode == 3 || res.exitCode == 4 {
		t.Errorf("not-found exit code must differ from hub-error(3)/network-error(4)")
	}
	lower := strings.ToLower(res.stderr)
	if !strings.Contains(lower, "not found") || !strings.Contains(res.stderr, "missing-input") {
		t.Errorf("expected a clear 'input not found' message naming the identifier, got:\n%s", res.stderr)
	}
	if res.stdout != "" {
		t.Errorf("expected empty stdout on failure, got:\n%s", res.stdout)
	}
}

func TestInputsGet_MissingIdentifier(t *testing.T) {
	res := runCLI(t, "inputs", "get")

	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
}
