package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func mockInputDeleteServer(t *testing.T, existingIDs map[string]bool, staticIDs map[string]bool) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/api/v2/inputs/")
		if !existingIDs[id] {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{"title": "Not Found", "detail": "input not found"})
			return
		}
		if staticIDs[id] {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"title": "Validation Error", "detail": "static inputs cannot be deleted"})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestInputsDelete_Success_YAML(t *testing.T) {
	srv := mockInputDeleteServer(t, map[string]bool{"spotify-1": true}, nil)

	res := runCLI(t, "delete", "inputs/spotify-1", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	for _, field := range []string{"inputId", "status", "message"} {
		if !strings.Contains(res.stdout, field) {
			t.Errorf("expected field %q in stdout, got:\n%s", field, res.stdout)
		}
	}
}

func TestInputsDelete_JSONOutput(t *testing.T) {
	srv := mockInputDeleteServer(t, map[string]bool{"spotify-1": true}, nil)

	res := runCLI(t, "delete", "inputs/spotify-1", "--hub-url", srv.URL, "--json")

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	var decoded struct {
		InputID string `json:"inputId"`
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal([]byte(res.stdout), &decoded); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\ngot: %s", err, res.stdout)
	}
	if decoded.InputID != "spotify-1" || decoded.Status == "" || decoded.Message == "" {
		t.Errorf("unexpected decoded content: %+v", decoded)
	}
}

func TestInputsDelete_NotFound(t *testing.T) {
	srv := mockInputDeleteServer(t, nil, nil)

	res := runCLI(t, "delete", "inputs/missing-input", "--hub-url", srv.URL)

	if res.exitCode != 5 {
		t.Fatalf("exit code = %d, want 5; stderr: %s", res.exitCode, res.stderr)
	}
	if res.stdout != "" {
		t.Errorf("expected empty stdout on failure, got:\n%s", res.stdout)
	}
}

func TestInputsDelete_StaticInput_400(t *testing.T) {
	srv := mockInputDeleteServer(t, map[string]bool{"spotify-1": true}, map[string]bool{"spotify-1": true})

	res := runCLI(t, "delete", "inputs/spotify-1", "--hub-url", srv.URL)

	if res.exitCode != 6 {
		t.Fatalf("exit code = %d, want 6; stderr: %s", res.exitCode, res.stderr)
	}
	if !strings.Contains(res.stderr, "static inputs cannot be deleted") {
		t.Errorf("expected the hub's error detail in stderr, got:\n%s", res.stderr)
	}
}

func TestInputsDelete_MissingIdentifier(t *testing.T) {
	res := runCLI(t, "delete", "inputs/")

	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
}

// TestInputsDelete_RoutesUnaffected confirms extending dispatchDelete to
// support inputs left `delete routes/<id>` behaving exactly as before.
func TestInputsDelete_RoutesUnaffected(t *testing.T) {
	srv, m := newMockRouteHub(t, nil, nil, nil)
	m.routeIDs["route_abc123"] = true

	res := runCLI(t, "delete", "routes/route_abc123", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
}

// TestStopInputs_IsUsageError confirms `stop` was not extended to `inputs` —
// it remains a routes-only alias of `delete`.
func TestStopInputs_IsUsageError(t *testing.T) {
	res := runCLI(t, "stop", "inputs/spotify-1")

	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
}
