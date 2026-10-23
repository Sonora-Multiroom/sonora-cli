package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

func mockOutputServer(t *testing.T, outputs map[string]map[string]any) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/api/v2/outputs/")
		o, ok := outputs[id]
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "not found"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(o)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestOutputsGet_JSONOutput(t *testing.T) {
	srv := mockOutputServer(t, map[string]map[string]any{
		"office-speaker": {"outputId": "office-speaker", "displayName": "Office Speaker", "volume": 75, "muted": false, "available": true, "enabled": true},
	})

	res := runCLI(t, "get", "outputs/office-speaker", "--hub-url", srv.URL, "--json")

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	var decoded struct {
		OutputID    string `json:"outputId"`
		DisplayName string `json:"displayName"`
		Volume      int    `json:"volume"`
		Muted       bool   `json:"muted"`
		Available   bool   `json:"available"`
		Enabled     bool   `json:"enabled"`
	}
	if err := json.Unmarshal([]byte(res.stdout), &decoded); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\ngot: %s", err, res.stdout)
	}
	if decoded.OutputID != "office-speaker" || decoded.DisplayName != "Office Speaker" ||
		decoded.Volume != 75 || decoded.Muted != false || decoded.Available != true || decoded.Enabled != true {
		t.Errorf("unexpected decoded content: %+v", decoded)
	}

	yamlRes := runCLI(t, "get", "outputs/office-speaker", "--hub-url", srv.URL)
	if yamlRes.exitCode != 0 {
		t.Fatalf("yaml exit code = %d, want 0; stderr: %s", yamlRes.exitCode, yamlRes.stderr)
	}
	if !strings.Contains(yamlRes.stdout, "office-speaker") || !strings.Contains(yamlRes.stdout, "Office Speaker") {
		t.Errorf("expected yaml view to match same values, got:\n%s", yamlRes.stdout)
	}
}

func TestOutputsGet_AliasIdenticalToFullName(t *testing.T) {
	srv := mockOutputServer(t, map[string]map[string]any{
		"office-speaker": {"outputId": "office-speaker", "displayName": "Office Speaker", "volume": 75, "muted": false, "available": true, "enabled": true},
	})

	full := runCLI(t, "get", "outputs/office-speaker", "--hub-url", srv.URL)
	alias := runCLI(t, "get", "out/office-speaker", "--hub-url", srv.URL)

	if alias.exitCode != full.exitCode {
		t.Fatalf("alias exit code = %d, want %d (same as full name); stderr: %s", alias.exitCode, full.exitCode, alias.stderr)
	}
	if alias.stdout != full.stdout {
		t.Errorf("alias stdout = %q, want identical to full-name stdout %q", alias.stdout, full.stdout)
	}
}

func TestOutputsGet_NotFound(t *testing.T) {
	srv := mockOutputServer(t, map[string]map[string]any{})

	res := runCLI(t, "get", "outputs/missing-speaker", "--hub-url", srv.URL)

	if res.exitCode != 5 {
		t.Fatalf("exit code = %d, want 5; stderr: %s", res.exitCode, res.stderr)
	}
	if res.exitCode == 3 || res.exitCode == 4 {
		t.Errorf("not-found exit code must differ from hub-error(3)/network-error(4)")
	}
	lower := strings.ToLower(res.stderr)
	if !strings.Contains(lower, "not found") || !strings.Contains(res.stderr, "missing-speaker") {
		t.Errorf("expected a clear 'output not found' message naming the identifier, got:\n%s", res.stderr)
	}
	if res.stdout != "" {
		t.Errorf("expected empty stdout on failure, got:\n%s", res.stdout)
	}
}

func TestOutputsGet_SuccessYAML(t *testing.T) {
	srv := mockOutputServer(t, map[string]map[string]any{
		"office-speaker": {"outputId": "office-speaker", "displayName": "Office Speaker", "volume": 75, "muted": false, "available": true, "enabled": true},
		"garage-speaker": {"outputId": "garage-speaker", "displayName": "Garage Speaker", "volume": 0, "muted": false, "available": false, "enabled": false},
	})

	// Warm up the freshly built binary once, untimed: the first exec of a
	// newly written .exe on Windows can carry a one-time AV-scan/cold-start
	// cost unrelated to the command's actual latency (SC-001 measures the
	// command, not first-process-launch overhead).
	runCLI(t, "get", "outputs/office-speaker", "--hub-url", srv.URL)

	for _, id := range []string{"office-speaker", "garage-speaker"} {
		start := time.Now()
		res := runCLI(t, "get", "outputs/"+id, "--hub-url", srv.URL)
		elapsed := time.Since(start)

		if res.exitCode != 0 {
			t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
		}
		if elapsed > time.Second {
			t.Errorf("expected completion under 1s (SC-001), took %v", elapsed)
		}
		for _, field := range []string{"outputId", "displayName", "volume", "muted", "available", "enabled"} {
			if !strings.Contains(res.stdout, field) {
				t.Errorf("id=%s: expected field %q in stdout, got:\n%s", id, field, res.stdout)
			}
		}
	}
}

// outputRoutesHub counts the requests mockOutputRoutesServer receives, per
// path.
type outputRoutesHub struct {
	mu       sync.Mutex
	requests map[string]int
}

func (m *outputRoutesHub) count(path string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.requests[path]
}

func (m *outputRoutesHub) total() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	n := 0
	for _, c := range m.requests {
		n += c
	}
	return n
}

// mockOutputRoutesServer serves GET /api/v2/outputs/<id> from output (a 404
// for any other ID) and GET /api/v2/outputs/<id>/routes with routesStatus and
// routesBody.
func mockOutputRoutesServer(t *testing.T, output map[string]any, routesStatus int, routesBody string) (*httptest.Server, *outputRoutesHub) {
	t.Helper()
	m := &outputRoutesHub{requests: map[string]int{}}
	id, _ := output["outputId"].(string)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.mu.Lock()
		m.requests[r.URL.Path]++
		m.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v2/outputs/" + id:
			_ = json.NewEncoder(w).Encode(output)
		case "/api/v2/outputs/" + id + "/routes":
			w.WriteHeader(routesStatus)
			_, _ = w.Write([]byte(routesBody))
		default:
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{"title": "Not Found", "detail": "Output not found"})
		}
	}))
	t.Cleanup(srv.Close)
	return srv, m
}

var kitchenOutput = map[string]any{"outputId": "kitchen", "displayName": "Kitchen", "volume": 30, "muted": false, "available": true, "enabled": true}

const kitchenRoutesBody = `[
	{"routeId": "route-1", "inputId": "radio", "targetId": "downstairs", "targetType": "OUTPUT_GROUP",
	 "status": "ACTIVE", "joinMode": "REPLACE", "createdAt": "2026-09-30T08:00:00Z",
	 "startedAt": "2026-09-30T08:00:01Z", "transferable": true, "pauseable": true, "paused": false,
	 "duckState": "LOWERED", "main": true},
	{"routeId": "route-2", "inputId": "doorbell", "targetId": "kitchen", "targetType": "SINGLE_OUTPUT",
	 "status": "STARTING", "joinMode": "DUCK_OTHERS", "createdAt": "2026-09-30T08:05:00Z",
	 "startedAt": null, "transferable": false, "pauseable": false, "paused": false,
	 "duckState": "FULL", "main": false}
]`

const kitchenRecordYAML = `outputId: "kitchen"
displayName: "Kitchen"
volume: 30
muted: false
available: true
enabled: true
`

const kitchenRoutesYAML = `routes:
  - routeId: "route-1"
    inputId: "radio"
    targetId: "downstairs"
    targetType: "OUTPUT_GROUP"
    status: "ACTIVE"
    joinMode: "REPLACE"
    duckState: "LOWERED"
    main: true
    paused: false
  - routeId: "route-2"
    inputId: "doorbell"
    targetId: "kitchen"
    targetType: "SINGLE_OUTPUT"
    status: "STARTING"
    joinMode: "DUCK_OTHERS"
    duckState: "FULL"
    main: false
    paused: false
`

func TestOutputsGet_IncludeRoutes_YAML(t *testing.T) {
	srv, m := mockOutputRoutesServer(t, kitchenOutput, http.StatusOK, kitchenRoutesBody)

	res := runCLI(t, "get", "outputs/kitchen", "--include-routes", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if want := kitchenRecordYAML + kitchenRoutesYAML; res.stdout != want {
		t.Errorf("got:\n%s\nwant:\n%s", res.stdout, want)
	}
	if m.count("/api/v2/outputs/kitchen") != 1 || m.count("/api/v2/outputs/kitchen/routes") != 1 {
		t.Errorf("expected one request to each path, got %v", m.requests)
	}
}

func TestOutputsGet_IncludeRoutes_JSON(t *testing.T) {
	srv, _ := mockOutputRoutesServer(t, kitchenOutput, http.StatusOK, kitchenRoutesBody)

	res := runCLI(t, "get", "outputs/kitchen", "--include-routes", "--json", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	var decoded map[string]any
	if err := json.Unmarshal([]byte(res.stdout), &decoded); err != nil {
		t.Fatalf("stdout is not one JSON object: %v\ngot: %s", err, res.stdout)
	}
	if decoded["outputId"] != "kitchen" || decoded["volume"] != float64(30) {
		t.Errorf("expected the output's fields at the top level, got: %s", res.stdout)
	}
	routes, ok := decoded["routes"].([]any)
	if !ok || len(routes) != 2 {
		t.Fatalf("expected a two-entry routes array, got: %s", res.stdout)
	}
	first := routes[0].(map[string]any)
	if first["routeId"] != "route-1" || first["duckState"] != "LOWERED" || first["main"] != true || first["createdAt"] != "2026-09-30T08:00:00Z" {
		t.Errorf("unexpected first route: %v", first)
	}
}

func TestOutputsGet_IncludeRoutes_Empty(t *testing.T) {
	srv, _ := mockOutputRoutesServer(t, kitchenOutput, http.StatusOK, `[]`)

	res := runCLI(t, "get", "outputs/kitchen", "--include-routes", "--hub-url", srv.URL)
	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if want := kitchenRecordYAML + "routes: []\n"; res.stdout != want {
		t.Errorf("got:\n%s\nwant:\n%s", res.stdout, want)
	}

	jsonRes := runCLI(t, "get", "outputs/kitchen", "--include-routes", "--json", "--hub-url", srv.URL)
	if jsonRes.exitCode != 0 {
		t.Fatalf("json exit code = %d, want 0; stderr: %s", jsonRes.exitCode, jsonRes.stderr)
	}
	if !strings.Contains(jsonRes.stdout, `"routes":[]`) {
		t.Errorf("expected \"routes\":[], got: %s", jsonRes.stdout)
	}
}

func TestOutputsGet_IncludeRoutes_OutputNotFound(t *testing.T) {
	srv, m := mockOutputRoutesServer(t, kitchenOutput, http.StatusOK, kitchenRoutesBody)

	res := runCLI(t, "get", "outputs/attic", "--include-routes", "--hub-url", srv.URL)

	if res.exitCode != 5 {
		t.Fatalf("exit code = %d, want 5; stderr: %s", res.exitCode, res.stderr)
	}
	if !strings.Contains(res.stderr, "output not found: attic") {
		t.Errorf("expected the output-not-found message, got: %s", res.stderr)
	}
	if res.stdout != "" {
		t.Errorf("expected empty stdout, got:\n%s", res.stdout)
	}
	if got := m.count("/api/v2/outputs/attic/routes"); got != 0 {
		t.Errorf("expected no routes request, got %d", got)
	}
}

func TestOutputsGet_IncludeRoutes_OlderHub(t *testing.T) {
	for _, status := range []int{http.StatusNotFound, http.StatusMethodNotAllowed} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			srv, _ := mockOutputRoutesServer(t, kitchenOutput, status, ``)

			res := runCLI(t, "get", "outputs/kitchen", "--include-routes", "--hub-url", srv.URL)

			if res.exitCode != 3 {
				t.Fatalf("exit code = %d, want 3; stderr: %s", res.exitCode, res.stderr)
			}
			if !strings.Contains(res.stderr, "this hub cannot list an output's routes; it needs hub 0.1.21 or later") {
				t.Errorf("expected the older-hub message, got: %s", res.stderr)
			}
			if strings.Contains(res.stderr, "not found") {
				t.Errorf("expected no not-found message, got: %s", res.stderr)
			}
			if strings.Contains(res.stderr, "detail:") {
				t.Errorf("expected no detail line without --verbose, got: %s", res.stderr)
			}
			if res.stdout != "" {
				t.Errorf("expected empty stdout, got:\n%s", res.stdout)
			}
		})
	}
}

func TestOutputsGet_IncludeRoutes_OlderHub_Verbose(t *testing.T) {
	srv, _ := mockOutputRoutesServer(t, kitchenOutput, http.StatusMethodNotAllowed, ``)

	res := runCLI(t, "get", "outputs/kitchen", "--include-routes", "--verbose", "--hub-url", srv.URL)

	if res.exitCode != 3 {
		t.Fatalf("exit code = %d, want 3; stderr: %s", res.exitCode, res.stderr)
	}
	if !strings.Contains(res.stderr, "detail:") {
		t.Errorf("expected a detail: line with --verbose, got: %s", res.stderr)
	}
}

func TestOutputsGet_IncludeRoutes_RoutesFailOtherwise(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
	}{
		{"server error", http.StatusInternalServerError, ``},
		{"not an array", http.StatusOK, `{"not":"an array"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, _ := mockOutputRoutesServer(t, kitchenOutput, tt.status, tt.body)

			res := runCLI(t, "get", "outputs/kitchen", "--include-routes", "--hub-url", srv.URL)

			if res.exitCode != 3 {
				t.Fatalf("exit code = %d, want 3; stderr: %s", res.exitCode, res.stderr)
			}
			if strings.Contains(res.stderr, "0.1.21") {
				t.Errorf("expected no older-hub message, got: %s", res.stderr)
			}
			if res.stdout != "" {
				t.Errorf("expected empty stdout, got:\n%s", res.stdout)
			}
		})
	}
}

func TestOutputsGet_WithoutIncludeRoutes_OneRequest(t *testing.T) {
	srv, m := mockOutputRoutesServer(t, kitchenOutput, http.StatusOK, kitchenRoutesBody)

	res := runCLI(t, "get", "outputs/kitchen", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if res.stdout != kitchenRecordYAML {
		t.Errorf("got:\n%s\nwant:\n%s", res.stdout, kitchenRecordYAML)
	}
	if m.total() != 1 || m.count("/api/v2/outputs/kitchen") != 1 {
		t.Errorf("expected exactly one request, to /api/v2/outputs/kitchen, got %v", m.requests)
	}
}

func TestOutputsGet_IncludeRoutes_FlagAfterID(t *testing.T) {
	srv, _ := mockOutputRoutesServer(t, kitchenOutput, http.StatusOK, kitchenRoutesBody)

	res := runCLI(t, "get", "outputs/kitchen", "--hub-url", srv.URL, "--include-routes")

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if want := kitchenRecordYAML + kitchenRoutesYAML; res.stdout != want {
		t.Errorf("got:\n%s\nwant:\n%s", res.stdout, want)
	}
}

// TestOutputsGet_IncludeRoutes_FlagBeforeID puts the flag straight after the
// path, ahead of every other flag. The dispatcher hands RunGet the ID first,
// so this is the nearest a command line gets to the flag preceding the ID.
func TestOutputsGet_IncludeRoutes_FlagBeforeID(t *testing.T) {
	srv, _ := mockOutputRoutesServer(t, kitchenOutput, http.StatusOK, kitchenRoutesBody)

	res := runCLI(t, "get", "outputs/kitchen", "--include-routes", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if want := kitchenRecordYAML + kitchenRoutesYAML; res.stdout != want {
		t.Errorf("got:\n%s\nwant:\n%s", res.stdout, want)
	}
}
