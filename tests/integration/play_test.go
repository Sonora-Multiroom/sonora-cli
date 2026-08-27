package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// mockPlayHub serves /api/v2/play, /api/v2/outputs/{id}, and
// /api/v2/groups/{id} for exercising the full `sonora play` CLI path,
// including target resolution, end to end. outputIDs/groupIDs name the
// identifiers that exist as each type; playStatus (0 = success) makes the
// play endpoint return a fixed non-2xx status with playBody as its error
// body, for exercising the hub-error paths.
type mockPlayHub struct {
	outputIDs   map[string]bool
	groupIDs    map[string]bool
	playStatus  int
	playBody    map[string]any
	playRequest map[string]any
}

func newMockPlayHub(t *testing.T, outputIDs, groupIDs []string) (*httptest.Server, *mockPlayHub) {
	t.Helper()
	m := &mockPlayHub{outputIDs: map[string]bool{}, groupIDs: map[string]bool{}}
	for _, id := range outputIDs {
		m.outputIDs[id] = true
	}
	for _, id := range groupIDs {
		m.groupIDs[id] = true
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v2/play":
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			m.playRequest = body
			w.Header().Set("Content-Type", "application/json")
			if m.playStatus != 0 {
				w.WriteHeader(m.playStatus)
				if m.playBody != nil {
					_ = json.NewEncoder(w).Encode(m.playBody)
				}
				return
			}
			name := ""
			if dn, ok := body["displayName"].(string); ok {
				name = dn
			}
			msg := "Playback started: Radio Stream → office-speaker"
			if name != "" {
				msg = "Playback started: " + name
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"inputId": "playback_1782345678",
				"route": map[string]any{
					"routeId": "route_abc123", "inputId": "playback_1782345678",
					"targetId": body["targetId"], "targetType": body["targetType"],
					"status": "STARTING", "createdAt": "2026-01-01T00:00:00Z",
					"startedAt": nil, "transferable": true, "pauseable": true, "paused": false,
				},
				"message": msg,
			})
		case strings.HasPrefix(r.URL.Path, "/api/v2/outputs/"):
			id := strings.TrimPrefix(r.URL.Path, "/api/v2/outputs/")
			if !m.outputIDs[id] {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]any{"error": "not found"})
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"outputId": id, "displayName": "Office Speaker", "volume": 50,
				"muted": false, "available": true, "enabled": true,
			})
		case strings.HasPrefix(r.URL.Path, "/api/v2/groups/"):
			id := strings.TrimPrefix(r.URL.Path, "/api/v2/groups/")
			if !m.groupIDs[id] {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]any{"error": "not found"})
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

func TestPlay_Success_SingleOutput_YAML(t *testing.T) {
	srv, _ := newMockPlayHub(t, []string{"office-speaker"}, nil)

	start := time.Now()
	res := runCLI(t, "play", "https://stream.example.com/live.mp3", "outputs/office-speaker", "--hub-url", srv.URL)
	elapsed := time.Since(start)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if elapsed > 5*time.Second {
		t.Errorf("expected completion under 5s (SC-003), took %v", elapsed)
	}
	for _, field := range []string{"inputId", "routeId", "status", "message"} {
		if !strings.Contains(res.stdout, field) {
			t.Errorf("expected field %q in stdout, got:\n%s", field, res.stdout)
		}
	}
}

func TestPlay_Success_SingleOutput_JSON(t *testing.T) {
	srv, _ := newMockPlayHub(t, []string{"office-speaker"}, nil)

	res := runCLI(t, "play", "https://stream.example.com/live.mp3", "outputs/office-speaker", "--hub-url", srv.URL, "--json")

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	var decoded struct {
		InputID string `json:"inputId"`
		RouteID string `json:"routeId"`
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal([]byte(res.stdout), &decoded); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\ngot: %s", err, res.stdout)
	}
	if decoded.InputID == "" || decoded.RouteID == "" || decoded.Status == "" || decoded.Message == "" {
		t.Errorf("unexpected decoded content: %+v", decoded)
	}
}

func TestPlay_HubErrorStatuses_MapToDistinctExitCodes(t *testing.T) {
	cases := []struct {
		name       string
		playStatus int
		wantExit   int
	}{
		{"400", http.StatusBadRequest, 6},
		{"422", http.StatusUnprocessableEntity, 8},
		{"502", http.StatusBadGateway, 9},
		{"503", http.StatusServiceUnavailable, 10},
		{"generic500", http.StatusInternalServerError, 3},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			srv, m := newMockPlayHub(t, []string{"office-speaker"}, nil)
			m.playStatus = c.playStatus
			m.playBody = map[string]any{"title": "Error", "detail": "something went wrong"}

			res := runCLI(t, "play", "https://stream.example.com/live.mp3", "outputs/office-speaker", "--hub-url", srv.URL)

			if res.exitCode != c.wantExit {
				t.Fatalf("exit code = %d, want %d; stderr: %s", res.exitCode, c.wantExit, res.stderr)
			}
			if res.stderr == "" {
				t.Error("expected a non-empty stderr message")
			}
			if res.stdout != "" {
				t.Errorf("expected empty stdout on failure, got:\n%s", res.stdout)
			}
		})
	}
}

func TestPlay_MalformedSuccessBody_ExitsHubError(t *testing.T) {
	// mockPlayHub always emits a well-formed play body on the success path,
	// so a malformed 200 (missing inputId) needs a dedicated fake server.
	malformedSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v2/play":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"route": map[string]any{"routeId": "route_1", "status": "STARTING"}, "message": "ok"})
		case strings.HasPrefix(r.URL.Path, "/api/v2/outputs/"):
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"outputId": "office-speaker", "displayName": "Office Speaker", "volume": 50,
				"muted": false, "available": true, "enabled": true,
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(malformedSrv.Close)

	res := runCLI(t, "play", "https://stream.example.com/live.mp3", "outputs/office-speaker", "--hub-url", malformedSrv.URL)

	if res.exitCode != 3 {
		t.Fatalf("exit code = %d, want 3; stderr: %s", res.exitCode, res.stderr)
	}
	if res.stdout != "" {
		t.Errorf("expected empty stdout on failure, got:\n%s", res.stdout)
	}
}

func TestPlay_GroupTarget_Success(t *testing.T) {
	srv, _ := newMockPlayHub(t, nil, []string{"whole-house"})

	res := runCLI(t, "play", "https://stream.example.com/live.mp3", "groups/whole-house", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
}

// TestPlay_OutputsPath_PicksOutputEvenWhenGroupAlsoExists proves the path
// prefix — not collision detection — picks the target when an id names both
// an output and a group. This keeps the value of the old
// --group/--output-forcing tests now that path-style addressing has removed
// the ambiguous-target case (exit 7) entirely (data-model.md).
func TestPlay_OutputsPath_PicksOutputEvenWhenGroupAlsoExists(t *testing.T) {
	srv, _ := newMockPlayHub(t, []string{"shared-id"}, []string{"shared-id"})

	res := runCLI(t, "play", "https://stream.example.com/live.mp3", "outputs/shared-id", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
}

func TestPlay_GroupsPath_PicksGroupEvenWhenOutputAlsoExists(t *testing.T) {
	srv, _ := newMockPlayHub(t, []string{"shared-id"}, []string{"shared-id"})

	res := runCLI(t, "play", "https://stream.example.com/live.mp3", "groups/shared-id", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
}

func TestPlay_GroupsPath_NotFoundWhenOnlyOutputExists(t *testing.T) {
	srv, _ := newMockPlayHub(t, []string{"output-only"}, nil)

	res := runCLI(t, "play", "https://stream.example.com/live.mp3", "groups/output-only", "--hub-url", srv.URL)

	if res.exitCode != 5 {
		t.Fatalf("exit code = %d, want 5; stderr: %s", res.exitCode, res.stderr)
	}
	if !strings.Contains(strings.ToLower(res.stderr), "group") {
		t.Errorf("expected a 'group not found' message, got:\n%s", res.stderr)
	}
}

func TestPlay_OutputsPath_NotFoundWhenOnlyGroupExists(t *testing.T) {
	srv, _ := newMockPlayHub(t, nil, []string{"group-only"})

	res := runCLI(t, "play", "https://stream.example.com/live.mp3", "outputs/group-only", "--hub-url", srv.URL)

	if res.exitCode != 5 {
		t.Fatalf("exit code = %d, want 5; stderr: %s", res.exitCode, res.stderr)
	}
	if !strings.Contains(strings.ToLower(res.stderr), "output") {
		t.Errorf("expected an 'output not found' message, got:\n%s", res.stderr)
	}
}

func TestPlay_TargetNotFound(t *testing.T) {
	srv, _ := newMockPlayHub(t, nil, nil)

	res := runCLI(t, "play", "https://stream.example.com/live.mp3", "outputs/nonexistent-id", "--hub-url", srv.URL)

	if res.exitCode != 5 {
		t.Fatalf("exit code = %d, want 5; stderr: %s", res.exitCode, res.stderr)
	}
}

func TestPlay_OldGroupOutputFlags_UsageError(t *testing.T) {
	srv, _ := newMockPlayHub(t, []string{"office-speaker"}, nil)

	res := runCLI(t, "play", "https://stream.example.com/live.mp3", "office-speaker", "--output", "--hub-url", srv.URL)

	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
}

func TestPlay_Volume_SetsRequestVolumeField(t *testing.T) {
	srv, m := newMockPlayHub(t, []string{"office-speaker"}, nil)

	res := runCLI(t, "play", "https://stream.example.com/live.mp3", "outputs/office-speaker", "--volume", "50", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if m.playRequest["volume"] != float64(50) {
		t.Errorf("expected volume=50 in request body, got: %+v", m.playRequest)
	}
}

func TestPlay_NoVolume_OmitsRequestVolumeField(t *testing.T) {
	srv, m := newMockPlayHub(t, []string{"office-speaker"}, nil)

	res := runCLI(t, "play", "https://stream.example.com/live.mp3", "outputs/office-speaker", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if _, ok := m.playRequest["volume"]; ok {
		t.Errorf("expected volume omitted from request body, got: %+v", m.playRequest)
	}
}

func TestPlay_Name_SetsDisplayNameField(t *testing.T) {
	srv, m := newMockPlayHub(t, []string{"office-speaker"}, nil)

	res := runCLI(t, "play", "https://stream.example.com/live.mp3", "outputs/office-speaker", "--display-name", "Kitchen Radio", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if m.playRequest["displayName"] != "Kitchen Radio" {
		t.Errorf("expected displayName=Kitchen Radio in request body, got: %+v", m.playRequest)
	}
	if !strings.Contains(res.stdout, "Kitchen Radio") {
		t.Errorf("expected rendered message to reflect the supplied name, got:\n%s", res.stdout)
	}
}

func TestPlay_NoName_OmitsDisplayNameField(t *testing.T) {
	srv, m := newMockPlayHub(t, []string{"office-speaker"}, nil)

	res := runCLI(t, "play", "https://stream.example.com/live.mp3", "outputs/office-speaker", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if _, ok := m.playRequest["displayName"]; ok {
		t.Errorf("expected displayName omitted from request body, got: %+v", m.playRequest)
	}
}

// TestPlay_TargetPathAliases_IdenticalToFullNames closes FR-004's
// "everywhere a resource path is accepted" claim for play: play calls the
// same respath.Parse the in/out/gr/rt aliases were added to (T014), so this
// exercises contracts/cli-play.md's example #3 invocation shape, which
// nothing else in this file covers.
func TestPlay_TargetPathAliases_IdenticalToFullNames(t *testing.T) {
	srv, _ := newMockPlayHub(t, []string{"office-speaker"}, []string{"whole-house"})

	outFull := runCLI(t, "play", "https://stream.example.com/live.mp3", "outputs/office-speaker", "--hub-url", srv.URL)
	outAlias := runCLI(t, "play", "https://stream.example.com/live.mp3", "out/office-speaker", "--hub-url", srv.URL)
	if outAlias.exitCode != outFull.exitCode {
		t.Fatalf("out/ alias exit code = %d, want %d; stderr: %s", outAlias.exitCode, outFull.exitCode, outAlias.stderr)
	}

	grFull := runCLI(t, "play", "https://stream.example.com/live.mp3", "groups/whole-house", "--hub-url", srv.URL)
	grAlias := runCLI(t, "play", "https://stream.example.com/live.mp3", "gr/whole-house", "--hub-url", srv.URL)
	if grAlias.exitCode != grFull.exitCode {
		t.Fatalf("gr/ alias exit code = %d, want %d; stderr: %s", grAlias.exitCode, grFull.exitCode, grAlias.stderr)
	}
}

func TestPlay_OldNameFlag_UsageError(t *testing.T) {
	srv, _ := newMockPlayHub(t, []string{"office-speaker"}, nil)

	res := runCLI(t, "play", "https://stream.example.com/live.mp3", "outputs/office-speaker", "--name", "Radio", "--hub-url", srv.URL)

	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
}
