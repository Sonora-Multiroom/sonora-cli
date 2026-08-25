package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func mockInputsServer(t *testing.T, body string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv
}

// mockInputsServerFilteringDisabled mirrors the real hub's documented
// behavior (api/openapi.json: "By default only enabled inputs are
// returned"): it drops disabled inputs unless includeDisabled=true is
// present on the request.
func mockInputsServerFilteringDisabled(t *testing.T, allInputs []map[string]any) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		includeDisabled := r.URL.Query().Get("includeDisabled") == "true"
		var filtered []map[string]any
		for _, i := range allInputs {
			if includeDisabled || i["enabled"] == true {
				filtered = append(filtered, i)
			}
		}
		if filtered == nil {
			filtered = []map[string]any{}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(filtered)
	}))
	t.Cleanup(srv.Close)
	return srv
}

var twoEnabledOneDisabledInputs = []map[string]any{
	{"inputId": "spotify-1", "displayName": "Spotify Stream", "uri": "https://stream.example.com/live.mp3", "enabled": true, "autoRemove": false, "source": "STATIC", "createdAt": nil, "pauseable": true},
	{"inputId": "line-in-1", "displayName": "Line In", "uri": "line://1", "enabled": true, "autoRemove": true, "source": "EPHEMERAL", "createdAt": "2026-06-22T14:30:00Z", "pauseable": false},
	{"inputId": "aux-1", "displayName": "Aux In", "uri": "aux://1", "enabled": false, "autoRemove": false, "source": "STATIC", "createdAt": nil, "pauseable": true},
}

func TestInputsList_DefaultEnabledOnlyYAML(t *testing.T) {
	srv := mockInputsServerFilteringDisabled(t, twoEnabledOneDisabledInputs)

	// Warm up the freshly built binary once, untimed: the first exec of a
	// newly written .exe on Windows can carry a one-time AV-scan/cold-start
	// cost unrelated to the command's actual latency (SC-001 measures the
	// command, not first-process-launch overhead) — mirrors
	// TestOutputsGet_SuccessYAML's warm-up.
	runCLI(t, "inputs", "list", "--hub-url", srv.URL)

	start := time.Now()
	res := runCLI(t, "inputs", "list", "--hub-url", srv.URL)
	elapsed := time.Since(start)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if elapsed > time.Second {
		t.Errorf("expected completion under 1s (SC-001), took %v", elapsed)
	}
	if !strings.Contains(res.stdout, "spotify-1") || !strings.Contains(res.stdout, "line-in-1") {
		t.Errorf("expected enabled inputs in stdout, got:\n%s", res.stdout)
	}
	if strings.Contains(res.stdout, "aux-1") {
		t.Errorf("disabled input should not appear by default, got:\n%s", res.stdout)
	}
	for _, field := range []string{"inputId", "displayName", "uri", "source", "enabled", "autoRemove", "pauseable", "createdAt"} {
		if !strings.Contains(res.stdout, field) {
			t.Errorf("expected field %q in stdout, got:\n%s", field, res.stdout)
		}
	}
}

func TestInputsList_ZeroInputsIsUnambiguousSuccess(t *testing.T) {
	srv := mockInputsServer(t, `[]`)

	res := runCLI(t, "inputs", "list", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	lower := strings.ToLower(res.stdout)
	if !strings.Contains(lower, "no inputs") {
		t.Errorf("expected an unambiguous 'no inputs' indication, got:\n%s", res.stdout)
	}
}

func TestInputsList_IncludeDisabledShowsBothStates(t *testing.T) {
	srv := mockInputsServerFilteringDisabled(t, twoEnabledOneDisabledInputs)

	res := runCLI(t, "inputs", "list", "--hub-url", srv.URL, "--include-disabled")

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if !strings.Contains(res.stdout, "spotify-1") || !strings.Contains(res.stdout, "aux-1") {
		t.Errorf("expected both enabled and disabled inputs in stdout, got:\n%s", res.stdout)
	}
	if !strings.Contains(res.stdout, "enabled: true") || !strings.Contains(res.stdout, "enabled: false") {
		t.Errorf("expected both enabled states visible, got:\n%s", res.stdout)
	}
}

func TestInputsList_JSONOutput(t *testing.T) {
	srv := mockInputsServer(t, `[{"inputId":"spotify-1","displayName":"Spotify Stream","uri":"u1","enabled":true,"autoRemove":false,"source":"STATIC","createdAt":null,"pauseable":true}]`)

	res := runCLI(t, "inputs", "list", "--hub-url", srv.URL, "--json")

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	var decoded struct {
		Inputs []struct {
			InputID     string  `json:"inputId"`
			DisplayName string  `json:"displayName"`
			URI         string  `json:"uri"`
			Enabled     bool    `json:"enabled"`
			AutoRemove  bool    `json:"autoRemove"`
			Source      string  `json:"source"`
			CreatedAt   *string `json:"createdAt"`
			Pauseable   bool    `json:"pauseable"`
		} `json:"inputs"`
	}
	if err := json.Unmarshal([]byte(res.stdout), &decoded); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\ngot: %s", err, res.stdout)
	}
	if len(decoded.Inputs) != 1 || decoded.Inputs[0].InputID != "spotify-1" || decoded.Inputs[0].CreatedAt != nil {
		t.Errorf("unexpected decoded content: %+v", decoded)
	}
}

func TestInputsList_UnreachableHub(t *testing.T) {
	res := runCLI(t, "inputs", "list", "--hub-url", "http://127.0.0.1:1")

	if res.exitCode != 4 {
		t.Fatalf("exit code = %d, want 4; stderr: %s", res.exitCode, res.stderr)
	}
	if res.stdout != "" {
		t.Errorf("expected empty stdout on failure, got:\n%s", res.stdout)
	}
}

func TestInputsList_HubNon2xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	res := runCLI(t, "inputs", "list", "--hub-url", srv.URL)

	if res.exitCode != 3 {
		t.Fatalf("exit code = %d, want 3; stderr: %s", res.exitCode, res.stderr)
	}
}

func TestInputsList_UnknownFlag(t *testing.T) {
	res := runCLI(t, "inputs", "list", "--unknown-flag")

	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
}
