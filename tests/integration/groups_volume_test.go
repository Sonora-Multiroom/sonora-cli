package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func mockGroupVolumeServer(t *testing.T, groupsByID map[string]map[string]any) (*httptest.Server, *string) {
	t.Helper()
	var lastBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/v2/groups/"), "/volume")
		g, ok := groupsByID[id]
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "not found"})
			return
		}
		var body struct {
			Volume int `json:"volume"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		lastBody = strings.TrimSpace(func() string {
			b, _ := json.Marshal(body)
			return string(b)
		}())
		g["volume"] = body.Volume
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(g)
	}))
	t.Cleanup(srv.Close)
	return srv, &lastBody
}

func TestGroupsSetVolume_Success(t *testing.T) {
	srv, lastBody := mockGroupVolumeServer(t, map[string]map[string]any{
		"downstairs": {"groupId": "downstairs", "volume": 50, "updatedAt": "2026-06-22T14:30:00Z"},
	})

	res := runCLI(t, "set", "groups/downstairs", "volume", "75", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if *lastBody != `{"volume":75}` {
		t.Errorf("got request body %s, want volume=75", *lastBody)
	}
	if !strings.Contains(res.stdout, "downstairs") {
		t.Errorf("expected groupId in stdout, got:\n%s", res.stdout)
	}
}

func TestGroupsSetVolume_JSONOutput(t *testing.T) {
	srv, _ := mockGroupVolumeServer(t, map[string]map[string]any{
		"downstairs": {"groupId": "downstairs", "volume": 50, "updatedAt": "2026-06-22T14:30:00Z"},
	})

	res := runCLI(t, "set", "groups/downstairs", "volume", "30", "--hub-url", srv.URL, "--json")

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	var decoded struct {
		GroupID string `json:"groupId"`
		Volume  int    `json:"volume"`
	}
	if err := json.Unmarshal([]byte(res.stdout), &decoded); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\ngot: %s", err, res.stdout)
	}
	if decoded.GroupID != "downstairs" || decoded.Volume != 30 {
		t.Errorf("unexpected decoded content: %+v", decoded)
	}
}

func TestGroupsSetVolume_NotFound(t *testing.T) {
	srv, _ := mockGroupVolumeServer(t, map[string]map[string]any{})

	res := runCLI(t, "set", "groups/missing-group", "volume", "50", "--hub-url", srv.URL)

	if res.exitCode != 5 {
		t.Fatalf("exit code = %d, want 5; stderr: %s", res.exitCode, res.stderr)
	}
	lower := strings.ToLower(res.stderr)
	if !strings.Contains(lower, "not found") || !strings.Contains(res.stderr, "missing-group") {
		t.Errorf("expected a clear 'group not found' message naming the identifier, got:\n%s", res.stderr)
	}
	if res.stdout != "" {
		t.Errorf("expected empty stdout on failure, got:\n%s", res.stdout)
	}
}

func TestGroupsSetVolume_OutOfRangeIsUsageError(t *testing.T) {
	res := runCLI(t, "set", "groups/downstairs", "volume", "150")

	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
}

func TestSetVolume_NonOutputsGroupsResourceIsUsageError(t *testing.T) {
	for _, resource := range []string{"inputs/spotify-1", "routes/route-abc-123"} {
		res := runCLI(t, "set", resource, "volume", "50")

		if res.exitCode != 2 {
			t.Fatalf("resource=%s: exit code = %d, want 2; stderr: %s", resource, res.exitCode, res.stderr)
		}
	}
}

func TestGroupsSetVolume_MissingIdentifier(t *testing.T) {
	res := runCLI(t, "set", "groups/")

	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
}
