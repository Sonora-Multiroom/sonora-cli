package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// mockRouteHub serves /api/v2/inputs/{id}, /api/v2/outputs/{id},
// /api/v2/groups/{id}, and /api/v2/routes for exercising the full
// `sonora route` CLI path end to end, including both pre-checks. inputIDs/
// outputIDs/groupIDs name the identifiers that exist as each type;
// routesStatus (0 = success) makes the routes endpoint return a fixed
// non-2xx status with routesBody as its error body, for exercising the
// hub-error paths; rawRoutesBody, if set, is written verbatim instead of a
// 201 success body, for exercising a malformed-response path.
type mockRouteHub struct {
	inputIDs  map[string]bool
	outputIDs map[string]bool
	groupIDs  map[string]bool
	routeIDs  map[string]bool

	routesStatus  int
	routesBody    map[string]any
	rawRoutesBody *string

	deleteStatus int
	deleteBody   map[string]any

	transferStatus  int
	transferBody    map[string]any
	rawTransferBody *string

	inputRequests    int32
	outputRequests   int32
	groupRequests    int32
	routesRequests   int32
	routesRequest    map[string]any
	deleteRequests   int32
	transferRequests int32
	transferRequest  map[string]any
}

func newMockRouteHub(t *testing.T, inputIDs, outputIDs, groupIDs []string) (*httptest.Server, *mockRouteHub) {
	t.Helper()
	m := &mockRouteHub{inputIDs: map[string]bool{}, outputIDs: map[string]bool{}, groupIDs: map[string]bool{}, routeIDs: map[string]bool{}}
	for _, id := range inputIDs {
		m.inputIDs[id] = true
	}
	for _, id := range outputIDs {
		m.outputIDs[id] = true
	}
	for _, id := range groupIDs {
		m.groupIDs[id] = true
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v2/routes" && r.Method == http.MethodPost:
			atomic.AddInt32(&m.routesRequests, 1)
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			m.routesRequest = body
			w.Header().Set("Content-Type", "application/json")
			if m.routesStatus != 0 {
				w.WriteHeader(m.routesStatus)
				if m.routesBody != nil {
					_ = json.NewEncoder(w).Encode(m.routesBody)
				}
				return
			}
			if m.rawRoutesBody != nil {
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(*m.rawRoutesBody))
				return
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"routeId": "route_abc123", "inputId": body["inputId"], "targetId": body["targetId"],
				"targetType": body["targetType"], "status": "STARTING", "createdAt": "2026-01-01T00:00:00Z",
				"startedAt": nil, "transferable": true, "pauseable": true, "paused": false,
			})
		case strings.HasPrefix(r.URL.Path, "/api/v2/routes/") && strings.HasSuffix(r.URL.Path, "/transfer") && r.Method == http.MethodPost:
			atomic.AddInt32(&m.transferRequests, 1)
			id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/v2/routes/"), "/transfer")
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			m.transferRequest = body
			if !m.routeIDs[id] {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]any{"title": "Not Found", "detail": "route not found"})
				return
			}
			w.Header().Set("Content-Type", "application/json")
			if m.transferStatus != 0 {
				w.WriteHeader(m.transferStatus)
				if m.transferBody != nil {
					_ = json.NewEncoder(w).Encode(m.transferBody)
				}
				return
			}
			if m.rawTransferBody != nil {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(*m.rawTransferBody))
				return
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"routeId": "route_new456", "inputId": "spotify-1", "targetId": body["targetId"],
				"targetType": body["targetType"], "status": "STARTING", "createdAt": "2026-01-01T00:00:00Z",
				"startedAt": nil, "transferable": true, "pauseable": true, "paused": false,
			})
		case strings.HasPrefix(r.URL.Path, "/api/v2/routes/") && r.Method == http.MethodDelete:
			atomic.AddInt32(&m.deleteRequests, 1)
			id := strings.TrimPrefix(r.URL.Path, "/api/v2/routes/")
			if !m.routeIDs[id] {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]any{"title": "Not Found", "detail": "route not found"})
				return
			}
			if m.deleteStatus != 0 {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(m.deleteStatus)
				if m.deleteBody != nil {
					_ = json.NewEncoder(w).Encode(m.deleteBody)
				}
				return
			}
			w.WriteHeader(http.StatusNoContent)
		case strings.HasPrefix(r.URL.Path, "/api/v2/inputs/"):
			atomic.AddInt32(&m.inputRequests, 1)
			id := strings.TrimPrefix(r.URL.Path, "/api/v2/inputs/")
			if !m.inputIDs[id] {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]any{"title": "Not Found", "detail": "input not found"})
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"inputId": id, "displayName": "Spotify", "uri": "spotify://track/1",
				"enabled": true, "autoRemove": false, "source": "STATIC", "createdAt": nil, "pauseable": true,
			})
		case strings.HasPrefix(r.URL.Path, "/api/v2/outputs/"):
			atomic.AddInt32(&m.outputRequests, 1)
			id := strings.TrimPrefix(r.URL.Path, "/api/v2/outputs/")
			if !m.outputIDs[id] {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]any{"title": "Not Found", "detail": "output not found"})
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"outputId": id, "displayName": "Office Speaker", "volume": 50,
				"muted": false, "available": true, "enabled": true,
			})
		case strings.HasPrefix(r.URL.Path, "/api/v2/groups/"):
			atomic.AddInt32(&m.groupRequests, 1)
			id := strings.TrimPrefix(r.URL.Path, "/api/v2/groups/")
			if !m.groupIDs[id] {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]any{"title": "Not Found", "detail": "group not found"})
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"groupId": id, "displayName": "Whole House", "outputIds": []string{"a", "b"},
				"muted": false, "enabled": true,
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)
	return srv, m
}

func TestRoute_Success_SingleOutput_YAML(t *testing.T) {
	srv, m := newMockRouteHub(t, []string{"spotify-1"}, []string{"office-speaker"}, nil)

	start := time.Now()
	res := runCLI(t, "route", "inputs/spotify-1", "outputs/office-speaker", "--hub-url", srv.URL)
	elapsed := time.Since(start)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if elapsed > 5*time.Second {
		t.Errorf("expected completion under 5s (SC-002), took %v", elapsed)
	}
	for _, field := range []string{"routeId", "status", "message"} {
		if !strings.Contains(res.stdout, field) {
			t.Errorf("expected field %q in stdout, got:\n%s", field, res.stdout)
		}
	}
	if atomic.LoadInt32(&m.inputRequests) == 0 {
		t.Error("expected at least one GET request to the input endpoint")
	}
}

func TestRoute_Success_SingleOutput_JSON(t *testing.T) {
	srv, _ := newMockRouteHub(t, []string{"spotify-1"}, []string{"office-speaker"}, nil)

	res := runCLI(t, "route", "inputs/spotify-1", "outputs/office-speaker", "--hub-url", srv.URL, "--json")

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	var decoded struct {
		RouteID string `json:"routeId"`
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal([]byte(res.stdout), &decoded); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\ngot: %s", err, res.stdout)
	}
	if decoded.RouteID == "" || decoded.Status == "" || decoded.Message == "" {
		t.Errorf("unexpected decoded content: %+v", decoded)
	}
}

func TestRoute_InputNotFound(t *testing.T) {
	srv, m := newMockRouteHub(t, nil, []string{"office-speaker"}, nil)

	res := runCLI(t, "route", "inputs/nonexistent", "outputs/office-speaker", "--hub-url", srv.URL)

	if res.exitCode != 11 {
		t.Fatalf("exit code = %d, want 11; stderr: %s", res.exitCode, res.stderr)
	}
	if !strings.Contains(strings.ToLower(res.stderr), "input") {
		t.Errorf("expected an 'input not found' message, got:\n%s", res.stderr)
	}
	if got := atomic.LoadInt32(&m.routesRequests); got != 0 {
		t.Errorf("expected zero requests to /api/v2/routes, got %d", got)
	}
}

func TestRoute_TargetNotFound_Output(t *testing.T) {
	srv, m := newMockRouteHub(t, []string{"spotify-1"}, nil, nil)

	res := runCLI(t, "route", "inputs/spotify-1", "outputs/nonexistent", "--hub-url", srv.URL)

	if res.exitCode != 12 {
		t.Fatalf("exit code = %d, want 12; stderr: %s", res.exitCode, res.stderr)
	}
	if !strings.Contains(strings.ToLower(res.stderr), "output") {
		t.Errorf("expected an 'output not found' message, got:\n%s", res.stderr)
	}
	if got := atomic.LoadInt32(&m.routesRequests); got != 0 {
		t.Errorf("expected zero requests to /api/v2/routes, got %d", got)
	}
}

func TestRoute_HubErrorStatuses_MapToDistinctExitCodes(t *testing.T) {
	cases := []struct {
		name         string
		routesStatus int
		wantExit     int
	}{
		{"400", http.StatusBadRequest, 6},
		{"422", http.StatusUnprocessableEntity, 8},
		{"generic500", http.StatusInternalServerError, 3},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			srv, m := newMockRouteHub(t, []string{"spotify-1"}, []string{"office-speaker"}, nil)
			m.routesStatus = c.routesStatus
			m.routesBody = map[string]any{"title": "Error", "detail": "something went wrong"}

			res := runCLI(t, "route", "inputs/spotify-1", "outputs/office-speaker", "--hub-url", srv.URL)

			if res.exitCode != c.wantExit {
				t.Fatalf("exit code = %d, want %d; stderr: %s", res.exitCode, c.wantExit, res.stderr)
			}
			if res.stdout != "" {
				t.Errorf("expected empty stdout on failure, got:\n%s", res.stdout)
			}
		})
	}
}

func TestRoute_MalformedSuccessBody_ExitsHubError(t *testing.T) {
	srv, m := newMockRouteHub(t, []string{"spotify-1"}, []string{"office-speaker"}, nil)
	raw := `{"routeId":"","inputId":"spotify-1","targetId":"office-speaker","targetType":"SINGLE_OUTPUT","status":"STARTING"}`
	m.rawRoutesBody = &raw

	res := runCLI(t, "route", "inputs/spotify-1", "outputs/office-speaker", "--hub-url", srv.URL)

	if res.exitCode != 3 {
		t.Fatalf("exit code = %d, want 3; stderr: %s", res.exitCode, res.stderr)
	}
	if res.stdout != "" {
		t.Errorf("expected empty stdout on failure, got:\n%s", res.stdout)
	}
}

func TestRoute_GroupTarget_Success(t *testing.T) {
	srv, _ := newMockRouteHub(t, []string{"spotify-1"}, nil, []string{"whole-house"})

	res := runCLI(t, "route", "inputs/spotify-1", "groups/whole-house", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
}

// TestRoute_CollidingIdentifier_TargetsOnlyStatedType proves the target
// path's prefix — not any disambiguation logic — picks the resource type,
// even when the same identifier exists as both an output and a group
// (FR-003, acceptance scenarios 2-3).
func TestRoute_CollidingIdentifier_TargetsOnlyStatedType(t *testing.T) {
	srv, m := newMockRouteHub(t, []string{"spotify-1"}, []string{"shared-id"}, []string{"shared-id"})

	res := runCLI(t, "route", "inputs/spotify-1", "groups/shared-id", "--hub-url", srv.URL)
	if res.exitCode != 0 {
		t.Fatalf("groups/shared-id: exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if atomic.LoadInt32(&m.groupRequests) == 0 {
		t.Error("expected a request to /api/v2/groups/{id}")
	}
	if atomic.LoadInt32(&m.outputRequests) != 0 {
		t.Error("expected zero requests to /api/v2/outputs/{id} when targeting groups/shared-id")
	}

	srv2, m2 := newMockRouteHub(t, []string{"spotify-1"}, []string{"shared-id"}, []string{"shared-id"})
	res2 := runCLI(t, "route", "inputs/spotify-1", "outputs/shared-id", "--hub-url", srv2.URL)
	if res2.exitCode != 0 {
		t.Fatalf("outputs/shared-id: exit code = %d, want 0; stderr: %s", res2.exitCode, res2.stderr)
	}
	if atomic.LoadInt32(&m2.outputRequests) == 0 {
		t.Error("expected a request to /api/v2/outputs/{id}")
	}
	if atomic.LoadInt32(&m2.groupRequests) != 0 {
		t.Error("expected zero requests to /api/v2/groups/{id} when targeting outputs/shared-id")
	}
}

// TestRoute_ExactTypeMismatch_DoesNotFallBackToOtherType covers FR-003a: an
// identifier that exists as one type must not be found when the other type
// is requested.
func TestRoute_ExactTypeMismatch_DoesNotFallBackToOtherType(t *testing.T) {
	srv, _ := newMockRouteHub(t, []string{"spotify-1"}, []string{"output-only"}, nil)
	res := runCLI(t, "route", "inputs/spotify-1", "groups/output-only", "--hub-url", srv.URL)
	if res.exitCode != 12 {
		t.Fatalf("exit code = %d, want 12; stderr: %s", res.exitCode, res.stderr)
	}
	if !strings.Contains(strings.ToLower(res.stderr), "group") {
		t.Errorf("expected a 'group not found' message, got:\n%s", res.stderr)
	}

	srv2, _ := newMockRouteHub(t, []string{"spotify-1"}, nil, []string{"group-only"})
	res2 := runCLI(t, "route", "inputs/spotify-1", "outputs/group-only", "--hub-url", srv2.URL)
	if res2.exitCode != 12 {
		t.Fatalf("exit code = %d, want 12; stderr: %s", res2.exitCode, res2.stderr)
	}
	if !strings.Contains(strings.ToLower(res2.stderr), "output") {
		t.Errorf("expected an 'output not found' message, got:\n%s", res2.stderr)
	}
}

func TestRoute_TargetKindRestrictedToOutputsGroups(t *testing.T) {
	srv, m := newMockRouteHub(t, []string{"spotify-1"}, nil, nil)
	res := runCLI(t, "route", "inputs/spotify-1", "inputs/some-id", "--hub-url", srv.URL)
	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
	if got := atomic.LoadInt32(&m.routesRequests); got != 0 {
		t.Errorf("expected zero requests to /api/v2/routes, got %d", got)
	}
}

func TestRoute_InputKindRestrictedToInputs(t *testing.T) {
	srv, _ := newMockRouteHub(t, nil, []string{"office-speaker"}, nil)
	res := runCLI(t, "route", "outputs/office-speaker", "inputs/spotify-1", "--hub-url", srv.URL)
	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
}

func TestRoute_MissingArguments(t *testing.T) {
	res := runCLI(t, "route")
	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
}

func TestRouteDelete_Success_YAML(t *testing.T) {
	srv, m := newMockRouteHub(t, nil, nil, nil)
	m.routeIDs["route_abc123"] = true

	res := runCLI(t, "delete", "routes/route_abc123", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	for _, field := range []string{"routeId", "status", "message"} {
		if !strings.Contains(res.stdout, field) {
			t.Errorf("expected field %q in stdout, got:\n%s", field, res.stdout)
		}
	}
	if got := atomic.LoadInt32(&m.deleteRequests); got != 1 {
		t.Errorf("expected exactly 1 DELETE request, got %d", got)
	}
}

func TestRouteStop_IsIdenticalAliasOfDelete(t *testing.T) {
	srv, m := newMockRouteHub(t, nil, nil, nil)
	m.routeIDs["route_abc123"] = true

	deleteRes := runCLI(t, "delete", "routes/route_abc123", "--hub-url", srv.URL)
	if deleteRes.exitCode != 0 {
		t.Fatalf("delete: exit code = %d, want 0; stderr: %s", deleteRes.exitCode, deleteRes.stderr)
	}

	srv2, m2 := newMockRouteHub(t, nil, nil, nil)
	m2.routeIDs["route_abc123"] = true
	stopRes := runCLI(t, "stop", "routes/route_abc123", "--hub-url", srv2.URL)
	if stopRes.exitCode != deleteRes.exitCode {
		t.Fatalf("stop: exit code = %d, want %d; stderr: %s", stopRes.exitCode, deleteRes.exitCode, stopRes.stderr)
	}
	if stopRes.stdout != deleteRes.stdout {
		t.Errorf("expected stop and delete to render identical output; stop:\n%s\ndelete:\n%s", stopRes.stdout, deleteRes.stdout)
	}
}

func TestRouteDelete_NotFound(t *testing.T) {
	srv, _ := newMockRouteHub(t, nil, nil, nil)

	res := runCLI(t, "delete", "routes/nonexistent", "--hub-url", srv.URL)

	if res.exitCode != 5 {
		t.Fatalf("exit code = %d, want 5; stderr: %s", res.exitCode, res.stderr)
	}
	if res.stdout != "" {
		t.Errorf("expected empty stdout on failure, got:\n%s", res.stdout)
	}
}

func TestRouteDelete_422_StopFailed(t *testing.T) {
	srv, m := newMockRouteHub(t, nil, nil, nil)
	m.routeIDs["route_abc123"] = true
	m.deleteStatus = http.StatusUnprocessableEntity
	m.deleteBody = map[string]any{"title": "Route Stop Error", "detail": "route could not be stopped"}

	res := runCLI(t, "delete", "routes/route_abc123", "--hub-url", srv.URL)

	if res.exitCode != 8 {
		t.Fatalf("exit code = %d, want 8; stderr: %s", res.exitCode, res.stderr)
	}
	if !strings.Contains(res.stderr, "route could not be stopped") {
		t.Errorf("expected the hub's error detail in stderr, got:\n%s", res.stderr)
	}
}

func TestRouteDelete_NonRoutesResource_IsUsageError(t *testing.T) {
	res := runCLI(t, "delete", "inputs/spotify-1")
	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
}

func TestRoute_TargetPathAliases_IdenticalToFullNames(t *testing.T) {
	srv, _ := newMockRouteHub(t, []string{"spotify-1"}, []string{"office-speaker"}, []string{"whole-house"})

	outFull := runCLI(t, "route", "inputs/spotify-1", "outputs/office-speaker", "--hub-url", srv.URL)
	outAlias := runCLI(t, "route", "in/spotify-1", "out/office-speaker", "--hub-url", srv.URL)
	if outAlias.exitCode != outFull.exitCode {
		t.Fatalf("in//out/ alias exit code = %d, want %d; stderr: %s", outAlias.exitCode, outFull.exitCode, outAlias.stderr)
	}

	grFull := runCLI(t, "route", "inputs/spotify-1", "groups/whole-house", "--hub-url", srv.URL)
	grAlias := runCLI(t, "route", "in/spotify-1", "gr/whole-house", "--hub-url", srv.URL)
	if grAlias.exitCode != grFull.exitCode {
		t.Fatalf("gr/ alias exit code = %d, want %d; stderr: %s", grAlias.exitCode, grFull.exitCode, grAlias.stderr)
	}
}

func TestRouteTransfer_Success_Output_YAML(t *testing.T) {
	srv, m := newMockRouteHub(t, nil, []string{"bedroom-speaker"}, nil)
	m.routeIDs["route_abc123"] = true

	res := runCLI(t, "transfer", "routes/route_abc123", "outputs/bedroom-speaker", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	for _, field := range []string{"routeId", "status", "message"} {
		if !strings.Contains(res.stdout, field) {
			t.Errorf("expected field %q in stdout, got:\n%s", field, res.stdout)
		}
	}
	if got := atomic.LoadInt32(&m.transferRequests); got != 1 {
		t.Errorf("expected exactly 1 transfer request, got %d", got)
	}
	if m.transferRequest["targetId"] != "bedroom-speaker" || m.transferRequest["targetType"] != "SINGLE_OUTPUT" {
		t.Errorf("unexpected transfer request body: %+v", m.transferRequest)
	}
}

func TestRouteTransfer_Success_Group_JSON(t *testing.T) {
	srv, m := newMockRouteHub(t, nil, nil, []string{"whole-house"})
	m.routeIDs["route_abc123"] = true

	res := runCLI(t, "transfer", "routes/route_abc123", "groups/whole-house", "--hub-url", srv.URL, "--json")

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	var decoded struct {
		RouteID string `json:"routeId"`
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal([]byte(res.stdout), &decoded); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\ngot: %s", err, res.stdout)
	}
	if decoded.RouteID != "route_new456" || decoded.Status == "" || decoded.Message == "" {
		t.Errorf("unexpected decoded content: %+v", decoded)
	}
	if m.transferRequest["targetType"] != "OUTPUT_GROUP" {
		t.Errorf("unexpected transfer request body: %+v", m.transferRequest)
	}
}

func TestRouteTransfer_TargetNotFound(t *testing.T) {
	srv, m := newMockRouteHub(t, nil, nil, nil)
	m.routeIDs["route_abc123"] = true

	res := runCLI(t, "transfer", "routes/route_abc123", "outputs/nonexistent", "--hub-url", srv.URL)

	if res.exitCode != 12 {
		t.Fatalf("exit code = %d, want 12; stderr: %s", res.exitCode, res.stderr)
	}
	if !strings.Contains(strings.ToLower(res.stderr), "output") {
		t.Errorf("expected an 'output not found' message, got:\n%s", res.stderr)
	}
	if got := atomic.LoadInt32(&m.transferRequests); got != 0 {
		t.Errorf("expected zero requests to the transfer endpoint, got %d", got)
	}
}

func TestRouteTransfer_RouteNotFound(t *testing.T) {
	srv, _ := newMockRouteHub(t, nil, []string{"bedroom-speaker"}, nil)

	res := runCLI(t, "transfer", "routes/nonexistent", "outputs/bedroom-speaker", "--hub-url", srv.URL)

	if res.exitCode != 5 {
		t.Fatalf("exit code = %d, want 5; stderr: %s", res.exitCode, res.stderr)
	}
	if res.stdout != "" {
		t.Errorf("expected empty stdout on failure, got:\n%s", res.stdout)
	}
}

func TestRouteTransfer_400_NotTransferable(t *testing.T) {
	srv, m := newMockRouteHub(t, nil, []string{"bedroom-speaker"}, nil)
	m.routeIDs["route_abc123"] = true
	m.transferStatus = http.StatusBadRequest
	m.transferBody = map[string]any{"title": "Not Transferable", "detail": "route is not transferable"}

	res := runCLI(t, "transfer", "routes/route_abc123", "outputs/bedroom-speaker", "--hub-url", srv.URL)

	if res.exitCode != 6 {
		t.Fatalf("exit code = %d, want 6; stderr: %s", res.exitCode, res.stderr)
	}
	if !strings.Contains(res.stderr, "route is not transferable") {
		t.Errorf("expected the hub's error detail in stderr, got:\n%s", res.stderr)
	}
}

func TestRouteTransfer_422_TransferFailed(t *testing.T) {
	srv, m := newMockRouteHub(t, nil, []string{"bedroom-speaker"}, nil)
	m.routeIDs["route_abc123"] = true
	m.transferStatus = http.StatusUnprocessableEntity
	m.transferBody = map[string]any{"title": "Transfer Error", "detail": "route transfer failed"}

	res := runCLI(t, "transfer", "routes/route_abc123", "outputs/bedroom-speaker", "--hub-url", srv.URL)

	if res.exitCode != 8 {
		t.Fatalf("exit code = %d, want 8; stderr: %s", res.exitCode, res.stderr)
	}
	if !strings.Contains(res.stderr, "route transfer failed") {
		t.Errorf("expected the hub's error detail in stderr, got:\n%s", res.stderr)
	}
}

func TestRouteTransfer_MalformedSuccessBody_ExitsHubError(t *testing.T) {
	srv, m := newMockRouteHub(t, nil, []string{"bedroom-speaker"}, nil)
	m.routeIDs["route_abc123"] = true
	raw := `{"routeId":"","inputId":"spotify-1","targetId":"bedroom-speaker","targetType":"SINGLE_OUTPUT","status":"STARTING"}`
	m.rawTransferBody = &raw

	res := runCLI(t, "transfer", "routes/route_abc123", "outputs/bedroom-speaker", "--hub-url", srv.URL)

	if res.exitCode != 3 {
		t.Fatalf("exit code = %d, want 3; stderr: %s", res.exitCode, res.stderr)
	}
	if res.stdout != "" {
		t.Errorf("expected empty stdout on failure, got:\n%s", res.stdout)
	}
}

func TestRouteTransfer_MissingArguments(t *testing.T) {
	res := runCLI(t, "transfer")
	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
}

func TestRouteTransfer_TargetKindRestrictedToOutputsGroups(t *testing.T) {
	srv, m := newMockRouteHub(t, nil, nil, nil)
	m.routeIDs["route_abc123"] = true

	res := runCLI(t, "transfer", "routes/route_abc123", "inputs/spotify-1", "--hub-url", srv.URL)

	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
	if got := atomic.LoadInt32(&m.transferRequests); got != 0 {
		t.Errorf("expected zero requests to the transfer endpoint, got %d", got)
	}
}
