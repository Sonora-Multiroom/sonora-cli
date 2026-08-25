package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func mockRouteServer(t *testing.T, routesByID map[string]map[string]any) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/api/v2/routes/")
		rt, ok := routesByID[id]
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "not found"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(rt)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestRoutesGet_SuccessYAML(t *testing.T) {
	srv := mockRouteServer(t, map[string]map[string]any{
		"route-abc-123": {"routeId": "route-abc-123", "inputId": "spotify-1", "targetId": "kitchen-speaker", "targetType": "SINGLE_OUTPUT", "status": "ACTIVE", "createdAt": "2026-06-22T14:30:00Z", "startedAt": "2026-06-22T14:30:01Z", "transferable": true, "pauseable": true, "paused": false},
		"route-def-456": {"routeId": "route-def-456", "inputId": "spotify-1", "targetId": "whole-house", "targetType": "OUTPUT_GROUP", "status": "STARTING", "createdAt": "2026-06-22T14:35:00Z", "startedAt": nil, "transferable": false, "pauseable": true, "paused": false},
	})

	// Warm up the freshly built binary once, untimed (see
	// TestOutputsGet_SuccessYAML's comment for why).
	runCLI(t, "routes", "get", "route-abc-123", "--hub-url", srv.URL)

	for _, id := range []string{"route-abc-123", "route-def-456"} {
		start := time.Now()
		res := runCLI(t, "routes", "get", id, "--hub-url", srv.URL)
		elapsed := time.Since(start)

		if res.exitCode != 0 {
			t.Fatalf("id=%s: exit code = %d, want 0; stderr: %s", id, res.exitCode, res.stderr)
		}
		if elapsed > time.Second {
			t.Errorf("id=%s: expected completion under 1s (SC-002), took %v", id, elapsed)
		}
		for _, field := range []string{"routeId", "inputId", "targetId", "targetType", "status", "createdAt", "startedAt", "transferable", "pauseable", "paused"} {
			if !strings.Contains(res.stdout, field) {
				t.Errorf("id=%s: expected field %q in stdout, got:\n%s", id, field, res.stdout)
			}
		}
	}
}

func TestRoutesGet_JSONOutput(t *testing.T) {
	srv := mockRouteServer(t, map[string]map[string]any{
		"route-abc-123": {"routeId": "route-abc-123", "inputId": "spotify-1", "targetId": "kitchen-speaker", "targetType": "SINGLE_OUTPUT", "status": "ACTIVE", "createdAt": "2026-06-22T14:30:00Z", "startedAt": "2026-06-22T14:30:01Z", "transferable": true, "pauseable": true, "paused": false},
	})

	res := runCLI(t, "routes", "get", "route-abc-123", "--hub-url", srv.URL, "--json")

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	var decoded struct {
		RouteID      string  `json:"routeId"`
		InputID      string  `json:"inputId"`
		TargetID     string  `json:"targetId"`
		TargetType   string  `json:"targetType"`
		Status       string  `json:"status"`
		CreatedAt    string  `json:"createdAt"`
		StartedAt    *string `json:"startedAt"`
		Transferable bool    `json:"transferable"`
		Pauseable    bool    `json:"pauseable"`
		Paused       bool    `json:"paused"`
	}
	if err := json.Unmarshal([]byte(res.stdout), &decoded); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\ngot: %s", err, res.stdout)
	}
	if decoded.RouteID != "route-abc-123" || decoded.StartedAt == nil || *decoded.StartedAt != "2026-06-22T14:30:01Z" {
		t.Errorf("unexpected decoded content: %+v", decoded)
	}
}

func TestRoutesGet_NotFound(t *testing.T) {
	srv := mockRouteServer(t, map[string]map[string]any{})

	res := runCLI(t, "routes", "get", "missing-route", "--hub-url", srv.URL)

	if res.exitCode != 5 {
		t.Fatalf("exit code = %d, want 5; stderr: %s", res.exitCode, res.stderr)
	}
	if res.exitCode == 3 || res.exitCode == 4 {
		t.Errorf("not-found exit code must differ from hub-error(3)/network-error(4)")
	}
	lower := strings.ToLower(res.stderr)
	if !strings.Contains(lower, "not found") || !strings.Contains(res.stderr, "missing-route") {
		t.Errorf("expected a clear 'route not found' message naming the identifier, got:\n%s", res.stderr)
	}
	if res.stdout != "" {
		t.Errorf("expected empty stdout on failure, got:\n%s", res.stdout)
	}
}

func TestRoutesGet_MissingIdentifier(t *testing.T) {
	res := runCLI(t, "routes", "get")

	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
}
