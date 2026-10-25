package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// mockRouteVolumeServer serves PUT /api/v2/routes/{id}/volume for the routes
// in routesByID, answering with the route and its new routeVolume, and 404
// for any other route. It records the last request body.
func mockRouteVolumeServer(t *testing.T, routesByID map[string]map[string]any) (*httptest.Server, *string) {
	t.Helper()
	var lastBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		id, ok := strings.CutSuffix(strings.TrimPrefix(r.URL.Path, "/api/v2/routes/"), "/volume")
		route, known := routesByID[id]
		if r.Method != http.MethodPut || !ok || !known {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{"title": "Not Found", "detail": "Route not found: " + id})
			return
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		b, _ := json.Marshal(body)
		lastBody = string(b)
		route["routeVolume"] = body["volume"]
		_ = json.NewEncoder(w).Encode(route)
	}))
	t.Cleanup(srv.Close)
	return srv, &lastBody
}

func playingRoute() map[string]any {
	return map[string]any{
		"routeId": "route-3f9a", "inputId": "radio", "targetId": "kitchen", "targetType": "SINGLE_OUTPUT",
		"status": "ACTIVE", "createdAt": "2026-10-25T08:00:00Z", "startedAt": "2026-10-25T08:00:01Z",
		"transferable": true, "pauseable": true, "paused": false, "joinMode": "MIX",
		"outputs": []string{"kitchen"}, "routeVolume": nil, "duckingLevelDb": nil,
	}
}

func TestRoutesSetVolume_Success(t *testing.T) {
	srv, lastBody := mockRouteVolumeServer(t, map[string]map[string]any{"route-3f9a": playingRoute()})

	res := runCLI(t, "set", "routes/route-3f9a", "volume", "40", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if *lastBody != `{"volume":40}` {
		t.Errorf("got request body %s, want volume=40", *lastBody)
	}
	for _, want := range []string{`routeId: "route-3f9a"`, "routeVolume: 40\n"} {
		if !strings.Contains(res.stdout, want) {
			t.Errorf("expected stdout to contain %q, got:\n%s", want, res.stdout)
		}
	}
}

func TestRoutesSetVolume_JSONOutput(t *testing.T) {
	srv, _ := mockRouteVolumeServer(t, map[string]map[string]any{"route-3f9a": playingRoute()})

	res := runCLI(t, "set", "routes/route-3f9a", "volume", "0", "--hub-url", srv.URL, "--json")

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	var decoded map[string]any
	if err := json.Unmarshal([]byte(res.stdout), &decoded); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\ngot: %s", err, res.stdout)
	}
	if decoded["routeId"] != "route-3f9a" || decoded["routeVolume"] != float64(0) {
		t.Errorf("unexpected decoded content: %+v", decoded)
	}
}

func TestRoutesSetVolume_NotFound(t *testing.T) {
	srv, _ := mockRouteVolumeServer(t, map[string]map[string]any{})

	res := runCLI(t, "set", "routes/missing-route", "volume", "40", "--hub-url", srv.URL)

	if res.exitCode != 5 {
		t.Fatalf("exit code = %d, want 5; stderr: %s", res.exitCode, res.stderr)
	}
	if !strings.Contains(strings.ToLower(res.stderr), "not found") || !strings.Contains(res.stderr, "missing-route") {
		t.Errorf("expected a 'route not found' message naming the identifier, got:\n%s", res.stderr)
	}
	if res.stdout != "" {
		t.Errorf("expected empty stdout on failure, got:\n%s", res.stdout)
	}
}

func TestRoutesSetVolume_OutOfRangeIsUsageError(t *testing.T) {
	res := runCLI(t, "set", "routes/route-3f9a", "volume", "150")

	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
}

func TestRoutesSetVolume_MissingIdentifier(t *testing.T) {
	res := runCLI(t, "set", "routes/")

	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
	if res.stdout != "" {
		t.Errorf("expected empty stdout, got:\n%s", res.stdout)
	}
}

func TestSet_UsageNamesRoutes(t *testing.T) {
	res := runCLI(t, "set", "inputs/spotify-1", "volume", "50")

	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
	for _, want := range []string{
		"usage: sonora set <outputs|groups|routes>/<id> volume <0-100> [flags]",
		"routes/<id> volume <0-100>",
	} {
		if !strings.Contains(res.stderr, want) {
			t.Errorf("expected stderr to contain %q, got:\n%s", want, res.stderr)
		}
	}
}
