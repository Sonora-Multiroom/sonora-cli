package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func mockGroupServer(t *testing.T, groupsByID map[string]map[string]any) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/api/v2/groups/")
		g, ok := groupsByID[id]
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "not found"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(g)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestGroupsGet_SuccessYAML(t *testing.T) {
	srv := mockGroupServer(t, map[string]map[string]any{
		"living-room":  {"groupId": "living-room", "displayName": "Living Room Speakers", "outputIds": []string{"office-speaker", "bedroom-speaker"}, "muted": false, "enabled": true},
		"unused-group": {"groupId": "unused-group", "displayName": "Unused Group", "outputIds": []string{}, "muted": true, "enabled": false},
	})

	// Warm up the freshly built binary once, untimed (see
	// TestOutputsGet_SuccessYAML's comment for why).
	runCLI(t, "groups", "get", "living-room", "--hub-url", srv.URL)

	for _, id := range []string{"living-room", "unused-group"} {
		start := time.Now()
		res := runCLI(t, "groups", "get", id, "--hub-url", srv.URL)
		elapsed := time.Since(start)

		if res.exitCode != 0 {
			t.Fatalf("id=%s: exit code = %d, want 0; stderr: %s", id, res.exitCode, res.stderr)
		}
		if elapsed > time.Second {
			t.Errorf("id=%s: expected completion under 1s (SC-002), took %v", id, elapsed)
		}
		for _, field := range []string{"groupId", "displayName", "outputIds", "muted", "enabled"} {
			if !strings.Contains(res.stdout, field) {
				t.Errorf("id=%s: expected field %q in stdout, got:\n%s", id, field, res.stdout)
			}
		}
	}
}

func TestGroupsGet_JSONOutput(t *testing.T) {
	srv := mockGroupServer(t, map[string]map[string]any{
		"living-room": {"groupId": "living-room", "displayName": "Living Room Speakers", "outputIds": []string{"office-speaker", "bedroom-speaker"}, "muted": false, "enabled": true},
	})

	res := runCLI(t, "groups", "get", "living-room", "--hub-url", srv.URL, "--json")

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	var decoded struct {
		GroupID     string   `json:"groupId"`
		DisplayName string   `json:"displayName"`
		OutputIDs   []string `json:"outputIds"`
		Muted       bool     `json:"muted"`
		Enabled     bool     `json:"enabled"`
	}
	if err := json.Unmarshal([]byte(res.stdout), &decoded); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\ngot: %s", err, res.stdout)
	}
	if decoded.GroupID != "living-room" || decoded.DisplayName != "Living Room Speakers" || len(decoded.OutputIDs) != 2 {
		t.Errorf("unexpected decoded content: %+v", decoded)
	}
}

func TestGroupsGet_NotFound(t *testing.T) {
	srv := mockGroupServer(t, map[string]map[string]any{})

	res := runCLI(t, "groups", "get", "missing-group", "--hub-url", srv.URL)

	if res.exitCode != 5 {
		t.Fatalf("exit code = %d, want 5; stderr: %s", res.exitCode, res.stderr)
	}
	if res.exitCode == 3 || res.exitCode == 4 {
		t.Errorf("not-found exit code must differ from hub-error(3)/network-error(4)")
	}
	lower := strings.ToLower(res.stderr)
	if !strings.Contains(lower, "not found") || !strings.Contains(res.stderr, "missing-group") {
		t.Errorf("expected a clear 'group not found' message naming the identifier, got:\n%s", res.stderr)
	}
	if res.stdout != "" {
		t.Errorf("expected empty stdout on failure, got:\n%s", res.stdout)
	}
}

func TestGroupsGet_MissingIdentifier(t *testing.T) {
	res := runCLI(t, "groups", "get")

	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
}
