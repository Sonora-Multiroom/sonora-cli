package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func mockRoutesServer(t *testing.T, body string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv
}

// mockRoutesServerFiltering mirrors the hub's documented listRoutes behavior:
// status/inputId/targetId query parameters are ANDed together, and any
// unrecognized status value 400s.
func mockRoutesServerFiltering(t *testing.T, allRoutes []map[string]any) *httptest.Server {
	t.Helper()
	validStatuses := map[string]bool{"STARTING": true, "ACTIVE": true, "STOPPING": true, "STOPPED": true, "FAILED": true}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		status := r.URL.Query().Get("status")
		inputID := r.URL.Query().Get("inputId")
		targetID := r.URL.Query().Get("targetId")
		if status != "" && !validStatuses[status] {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "invalid status"})
			return
		}
		var filtered []map[string]any
		for _, rt := range allRoutes {
			if status != "" && rt["status"] != status {
				continue
			}
			if inputID != "" && rt["inputId"] != inputID {
				continue
			}
			if targetID != "" && rt["targetId"] != targetID {
				continue
			}
			filtered = append(filtered, rt)
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

var mixedStatusRoutes = []map[string]any{
	{"routeId": "route-abc-123", "inputId": "spotify-1", "targetId": "kitchen-speaker", "targetType": "SINGLE_OUTPUT", "status": "ACTIVE", "createdAt": "2026-06-22T14:30:00Z", "startedAt": "2026-06-22T14:30:01Z", "transferable": true, "pauseable": true, "paused": false},
	{"routeId": "route-def-456", "inputId": "spotify-1", "targetId": "whole-house", "targetType": "OUTPUT_GROUP", "status": "FAILED", "createdAt": "2026-06-22T14:35:00Z", "startedAt": nil, "transferable": false, "pauseable": true, "paused": false},
	{"routeId": "route-ghi-789", "inputId": "line-in-1", "targetId": "kitchen-speaker", "targetType": "SINGLE_OUTPUT", "status": "STARTING", "createdAt": "2026-06-22T14:40:00Z", "startedAt": nil, "transferable": false, "pauseable": false, "paused": false},
}

func TestRoutesList_DefaultShowsAllStatusesYAML(t *testing.T) {
	srv := mockRoutesServerFiltering(t, mixedStatusRoutes)

	// Warm up the freshly built binary once, untimed (see
	// TestOutputsGet_SuccessYAML's comment for why).
	runCLI(t, "get", "routes", "--hub-url", srv.URL)

	start := time.Now()
	res := runCLI(t, "get", "routes", "--hub-url", srv.URL)
	elapsed := time.Since(start)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if elapsed > time.Second {
		t.Errorf("expected completion under 1s (SC-001), took %v", elapsed)
	}
	for _, id := range []string{"route-abc-123", "route-def-456", "route-ghi-789"} {
		if !strings.Contains(res.stdout, id) {
			t.Errorf("expected route %q in stdout, got:\n%s", id, res.stdout)
		}
	}
	for _, field := range []string{"routeId", "inputId", "targetId", "targetType", "status"} {
		if !strings.Contains(res.stdout, field) {
			t.Errorf("expected field %q in stdout, got:\n%s", field, res.stdout)
		}
	}
}

func TestRoutesList_ZeroRoutesIsUnambiguousSuccess(t *testing.T) {
	srv := mockRoutesServer(t, `[]`)

	res := runCLI(t, "get", "routes", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	lower := strings.ToLower(res.stdout)
	if !strings.Contains(lower, "no routes") {
		t.Errorf("expected an unambiguous 'no routes' indication, got:\n%s", res.stdout)
	}
}

func TestRoutesList_StatusFilterNarrowsResults(t *testing.T) {
	srv := mockRoutesServerFiltering(t, mixedStatusRoutes)

	res := runCLI(t, "get", "routes", "--hub-url", srv.URL, "--status", "FAILED")

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if !strings.Contains(res.stdout, "route-def-456") {
		t.Errorf("expected the FAILED route in stdout, got:\n%s", res.stdout)
	}
	if strings.Contains(res.stdout, "route-abc-123") || strings.Contains(res.stdout, "route-ghi-789") {
		t.Errorf("expected only the FAILED route, got:\n%s", res.stdout)
	}
}

func TestRoutesList_InputIDFilterNarrowsResults(t *testing.T) {
	srv := mockRoutesServerFiltering(t, mixedStatusRoutes)

	res := runCLI(t, "get", "routes", "--hub-url", srv.URL, "--input-id", "line-in-1")

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if !strings.Contains(res.stdout, "route-ghi-789") {
		t.Errorf("expected the line-in-1 route in stdout, got:\n%s", res.stdout)
	}
	if strings.Contains(res.stdout, "route-abc-123") || strings.Contains(res.stdout, "route-def-456") {
		t.Errorf("expected only the line-in-1 route, got:\n%s", res.stdout)
	}
}

func TestRoutesList_TargetIDFilterNarrowsResults(t *testing.T) {
	srv := mockRoutesServerFiltering(t, mixedStatusRoutes)

	res := runCLI(t, "get", "routes", "--hub-url", srv.URL, "--target-id", "whole-house")

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if !strings.Contains(res.stdout, "route-def-456") {
		t.Errorf("expected the whole-house route in stdout, got:\n%s", res.stdout)
	}
	if strings.Contains(res.stdout, "route-abc-123") || strings.Contains(res.stdout, "route-ghi-789") {
		t.Errorf("expected only the whole-house route, got:\n%s", res.stdout)
	}
}

func TestRoutesList_CombinedFiltersUseANDLogic(t *testing.T) {
	srv := mockRoutesServerFiltering(t, mixedStatusRoutes)

	res := runCLI(t, "get", "routes", "--hub-url", srv.URL, "--status", "ACTIVE", "--target-id", "kitchen-speaker")

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if !strings.Contains(res.stdout, "route-abc-123") {
		t.Errorf("expected route-abc-123 (ACTIVE + kitchen-speaker), got:\n%s", res.stdout)
	}
	if strings.Contains(res.stdout, "route-def-456") || strings.Contains(res.stdout, "route-ghi-789") {
		t.Errorf("expected only the route matching both filters, got:\n%s", res.stdout)
	}
}

func TestRoutesList_InvalidStatusFilterExitsHubError(t *testing.T) {
	srv := mockRoutesServerFiltering(t, mixedStatusRoutes)

	res := runCLI(t, "get", "routes", "--hub-url", srv.URL, "--status", "NOT_A_REAL_STATUS")

	if res.exitCode != 3 {
		t.Fatalf("exit code = %d, want 3; stderr: %s", res.exitCode, res.stderr)
	}
	if res.stdout != "" {
		t.Errorf("expected empty stdout on failure, got:\n%s", res.stdout)
	}
}

func TestRoutesList_JSONOutput(t *testing.T) {
	srv := mockRoutesServer(t, `[{"routeId":"route-abc-123","inputId":"spotify-1","targetId":"kitchen-speaker","targetType":"SINGLE_OUTPUT","status":"ACTIVE","createdAt":"2026-06-22T14:30:00Z","startedAt":"2026-06-22T14:30:01Z","transferable":true,"pauseable":true,"paused":false}]`)

	res := runCLI(t, "get", "routes", "--hub-url", srv.URL, "--json")

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	var decoded struct {
		Routes []struct {
			RouteID    string `json:"routeId"`
			InputID    string `json:"inputId"`
			TargetID   string `json:"targetId"`
			TargetType string `json:"targetType"`
			Status     string `json:"status"`
		} `json:"routes"`
	}
	if err := json.Unmarshal([]byte(res.stdout), &decoded); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\ngot: %s", err, res.stdout)
	}
	if len(decoded.Routes) != 1 || decoded.Routes[0].RouteID != "route-abc-123" {
		t.Errorf("unexpected decoded content: %+v", decoded)
	}
}

func TestRoutesList_UnreachableHub(t *testing.T) {
	res := runCLI(t, "get", "routes", "--hub-url", "http://127.0.0.1:1")

	if res.exitCode != 4 {
		t.Fatalf("exit code = %d, want 4; stderr: %s", res.exitCode, res.stderr)
	}
	if res.stdout != "" {
		t.Errorf("expected empty stdout on failure, got:\n%s", res.stdout)
	}
}

func TestRoutesList_HubNon2xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	res := runCLI(t, "get", "routes", "--hub-url", srv.URL)

	if res.exitCode != 3 {
		t.Fatalf("exit code = %d, want 3; stderr: %s", res.exitCode, res.stderr)
	}
}

func TestRoutesList_UnknownFlag(t *testing.T) {
	res := runCLI(t, "get", "routes", "--unknown-flag")

	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
}
