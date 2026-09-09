package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func mockBulkStopServer(t *testing.T, knownIDs map[string]bool, response map[string]any) (*httptest.Server, *string) {
	t.Helper()
	var lastPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastPath = r.URL.Path
		if knownIDs != nil {
			parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
			// api/v2/{outputs|groups}/{id}/routes
			if len(parts) == 5 && !knownIDs[parts[3]] {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]any{"error": "not found"})
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	t.Cleanup(srv.Close)
	return srv, &lastPath
}

func TestStopRoutes_Success(t *testing.T) {
	srv, lastPath := mockBulkStopServer(t, nil, map[string]any{
		"stoppedCount": 2,
		"stoppedRoutes": []map[string]any{
			{"routeId": "route-1", "targetType": "SINGLE_OUTPUT", "targetId": "kitchen-speaker", "stopReason": nil},
			{"routeId": "route-2", "targetType": "OUTPUT_GROUP", "targetId": "living-room", "stopReason": nil},
		},
	})

	res := runCLI(t, "stop", "routes", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if *lastPath != "/api/v2/routes" {
		t.Errorf("got path %q, want /api/v2/routes", *lastPath)
	}
	if !strings.Contains(res.stdout, "stoppedCount: 2") {
		t.Errorf("expected stoppedCount: 2 in stdout, got:\n%s", res.stdout)
	}
}

func TestStopRoutes_ZeroStopped(t *testing.T) {
	srv, _ := mockBulkStopServer(t, nil, map[string]any{"stoppedCount": 0, "stoppedRoutes": []any{}})

	res := runCLI(t, "stop", "routes", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if !strings.Contains(res.stdout, "stoppedCount: 0") {
		t.Errorf("expected stoppedCount: 0 in stdout, got:\n%s", res.stdout)
	}
}

func TestStopOutputs_JSONOutput(t *testing.T) {
	srv, lastPath := mockBulkStopServer(t, map[string]bool{"office-speaker": true}, map[string]any{
		"stoppedCount": 1,
		"stoppedRoutes": []map[string]any{
			{"routeId": "route-1", "targetType": "SINGLE_OUTPUT", "targetId": "office-speaker", "stopReason": "DIRECT"},
		},
	})

	res := runCLI(t, "stop", "outputs/office-speaker", "--hub-url", srv.URL, "--json")

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if *lastPath != "/api/v2/outputs/office-speaker/routes" {
		t.Errorf("got path %q, want /api/v2/outputs/office-speaker/routes", *lastPath)
	}
	var decoded struct {
		StoppedCount int `json:"stoppedCount"`
	}
	if err := json.Unmarshal([]byte(res.stdout), &decoded); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\ngot: %s", err, res.stdout)
	}
	if decoded.StoppedCount != 1 {
		t.Errorf("unexpected decoded content: %+v", decoded)
	}
}

func TestStopOutputs_NotFound(t *testing.T) {
	srv, _ := mockBulkStopServer(t, map[string]bool{}, nil)

	res := runCLI(t, "stop", "outputs/missing-output", "--hub-url", srv.URL)

	if res.exitCode != 5 {
		t.Fatalf("exit code = %d, want 5; stderr: %s", res.exitCode, res.stderr)
	}
	lower := strings.ToLower(res.stderr)
	if !strings.Contains(lower, "not found") || !strings.Contains(res.stderr, "missing-output") {
		t.Errorf("expected a clear 'output not found' message naming the identifier, got:\n%s", res.stderr)
	}
	if res.stdout != "" {
		t.Errorf("expected empty stdout on failure, got:\n%s", res.stdout)
	}
}

func TestStopGroups_Success(t *testing.T) {
	srv, lastPath := mockBulkStopServer(t, map[string]bool{"living-room": true}, map[string]any{
		"stoppedCount": 1,
		"stoppedRoutes": []map[string]any{
			{"routeId": "route-1", "targetType": "OUTPUT_GROUP", "targetId": "living-room", "stopReason": "DIRECT"},
		},
	})

	res := runCLI(t, "stop", "groups/living-room", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if *lastPath != "/api/v2/groups/living-room/routes" {
		t.Errorf("got path %q, want /api/v2/groups/living-room/routes", *lastPath)
	}
	if !strings.Contains(res.stdout, "living-room") {
		t.Errorf("expected targetId in stdout, got:\n%s", res.stdout)
	}
}

func TestStopGroups_NotFound(t *testing.T) {
	srv, _ := mockBulkStopServer(t, map[string]bool{}, nil)

	res := runCLI(t, "stop", "groups/missing-group", "--hub-url", srv.URL)

	if res.exitCode != 5 {
		t.Fatalf("exit code = %d, want 5; stderr: %s", res.exitCode, res.stderr)
	}
	if !strings.Contains(res.stderr, "missing-group") {
		t.Errorf("expected the identifier in stderr, got:\n%s", res.stderr)
	}
}

func TestStopRouteByID_StillWorks(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/routes/route-abc-123" {
			t.Errorf("got path %q, want /api/v2/routes/route-abc-123", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	res := runCLI(t, "stop", "routes/route-abc-123", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if !strings.Contains(res.stdout, "route-abc-123") {
		t.Errorf("expected routeId in stdout, got:\n%s", res.stdout)
	}
}

func TestStop_BareOutputsAndGroups_Rejected(t *testing.T) {
	for _, resource := range []string{"outputs", "groups"} {
		res := runCLI(t, "stop", resource)
		if res.exitCode != 2 {
			t.Errorf("stop %s: exit code = %d, want 2; stderr: %s", resource, res.exitCode, res.stderr)
		}
	}
}

func TestStop_Inputs_Rejected(t *testing.T) {
	res := runCLI(t, "stop", "inputs/spotify-1")

	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
}
