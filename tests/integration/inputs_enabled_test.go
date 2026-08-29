package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func mockInputEnabledServer(t *testing.T, inputsByID map[string]map[string]any) (*httptest.Server, *string) {
	t.Helper()
	var lastBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/v2/inputs/"), "/enabled")
		i, ok := inputsByID[id]
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "not found"})
			return
		}
		var body struct {
			Enabled bool `json:"enabled"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		lastBody = strings.TrimSpace(func() string {
			b, _ := json.Marshal(body)
			return string(b)
		}())
		i["enabled"] = body.Enabled
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(i)
	}))
	t.Cleanup(srv.Close)
	return srv, &lastBody
}

func TestInputsEnable_Success(t *testing.T) {
	srv, lastBody := mockInputEnabledServer(t, map[string]map[string]any{
		"spotify-1": {"inputId": "spotify-1", "displayName": "Spotify Stream", "uri": "u1", "enabled": false, "autoRemove": false, "source": "STATIC", "createdAt": nil, "pauseable": true},
	})

	res := runCLI(t, "enable", "inputs/spotify-1", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if *lastBody != `{"enabled":true}` {
		t.Errorf("got request body %s, want enabled=true", *lastBody)
	}
	if !strings.Contains(res.stdout, "spotify-1") {
		t.Errorf("expected inputId in stdout, got:\n%s", res.stdout)
	}
}

func TestInputsDisable_Success(t *testing.T) {
	srv, lastBody := mockInputEnabledServer(t, map[string]map[string]any{
		"spotify-1": {"inputId": "spotify-1", "displayName": "Spotify Stream", "uri": "u1", "enabled": true, "autoRemove": false, "source": "STATIC", "createdAt": nil, "pauseable": true},
	})

	res := runCLI(t, "disable", "inputs/spotify-1", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if *lastBody != `{"enabled":false}` {
		t.Errorf("got request body %s, want enabled=false", *lastBody)
	}
}

func TestInputsEnable_JSONOutput(t *testing.T) {
	srv, _ := mockInputEnabledServer(t, map[string]map[string]any{
		"spotify-1": {"inputId": "spotify-1", "displayName": "Spotify Stream", "uri": "u1", "enabled": false, "autoRemove": false, "source": "STATIC", "createdAt": nil, "pauseable": true},
	})

	res := runCLI(t, "enable", "inputs/spotify-1", "--hub-url", srv.URL, "--json")

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	var decoded struct {
		InputID string `json:"inputId"`
		Enabled bool   `json:"enabled"`
	}
	if err := json.Unmarshal([]byte(res.stdout), &decoded); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\ngot: %s", err, res.stdout)
	}
	if decoded.InputID != "spotify-1" || !decoded.Enabled {
		t.Errorf("unexpected decoded content: %+v", decoded)
	}
}

func TestInputsEnable_NotFound(t *testing.T) {
	srv, _ := mockInputEnabledServer(t, map[string]map[string]any{})

	res := runCLI(t, "enable", "inputs/missing-input", "--hub-url", srv.URL)

	if res.exitCode != 5 {
		t.Fatalf("exit code = %d, want 5; stderr: %s", res.exitCode, res.stderr)
	}
	lower := strings.ToLower(res.stderr)
	if !strings.Contains(lower, "not found") || !strings.Contains(res.stderr, "missing-input") {
		t.Errorf("expected a clear 'input not found' message naming the identifier, got:\n%s", res.stderr)
	}
	if res.stdout != "" {
		t.Errorf("expected empty stdout on failure, got:\n%s", res.stdout)
	}
}

func TestInputsEnable_UnsupportedResourceIsUsageError(t *testing.T) {
	res := runCLI(t, "enable", "routes/route-abc-123")

	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
}

func TestInputsDisable_MissingIdentifier(t *testing.T) {
	res := runCLI(t, "disable", "inputs/")

	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
}
