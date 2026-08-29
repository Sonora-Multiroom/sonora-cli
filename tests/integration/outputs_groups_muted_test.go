package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func mockOutputMutedServer(t *testing.T, outputsByID map[string]map[string]any) (*httptest.Server, *string) {
	t.Helper()
	var lastBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/v2/outputs/"), "/mute")
		o, ok := outputsByID[id]
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "not found"})
			return
		}
		var body struct {
			Muted bool `json:"muted"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		lastBody = strings.TrimSpace(func() string {
			b, _ := json.Marshal(body)
			return string(b)
		}())
		o["muted"] = body.Muted
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(o)
	}))
	t.Cleanup(srv.Close)
	return srv, &lastBody
}

func mockGroupMutedServer(t *testing.T, groupsByID map[string]map[string]any) (*httptest.Server, *string) {
	t.Helper()
	var lastBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/v2/groups/"), "/mute")
		g, ok := groupsByID[id]
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "not found"})
			return
		}
		var body struct {
			Muted bool `json:"muted"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		lastBody = strings.TrimSpace(func() string {
			b, _ := json.Marshal(body)
			return string(b)
		}())
		g["muted"] = body.Muted
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(g)
	}))
	t.Cleanup(srv.Close)
	return srv, &lastBody
}

func TestOutputsMute_Success(t *testing.T) {
	srv, lastBody := mockOutputMutedServer(t, map[string]map[string]any{
		"office-speaker": {"outputId": "office-speaker", "displayName": "Office Speaker", "volume": 40, "muted": false, "available": true, "enabled": true},
	})

	res := runCLI(t, "mute", "outputs/office-speaker", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if *lastBody != `{"muted":true}` {
		t.Errorf("got request body %s, want muted=true", *lastBody)
	}
	if !strings.Contains(res.stdout, "office-speaker") {
		t.Errorf("expected outputId in stdout, got:\n%s", res.stdout)
	}
}

func TestOutputsUnmute_Success(t *testing.T) {
	srv, lastBody := mockOutputMutedServer(t, map[string]map[string]any{
		"office-speaker": {"outputId": "office-speaker", "displayName": "Office Speaker", "volume": 40, "muted": true, "available": true, "enabled": true},
	})

	res := runCLI(t, "unmute", "outputs/office-speaker", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if *lastBody != `{"muted":false}` {
		t.Errorf("got request body %s, want muted=false", *lastBody)
	}
}

func TestOutputsMute_JSONOutput(t *testing.T) {
	srv, _ := mockOutputMutedServer(t, map[string]map[string]any{
		"office-speaker": {"outputId": "office-speaker", "displayName": "Office Speaker", "volume": 40, "muted": false, "available": true, "enabled": true},
	})

	res := runCLI(t, "mute", "outputs/office-speaker", "--hub-url", srv.URL, "--json")

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	var decoded struct {
		OutputID string `json:"outputId"`
		Muted    bool   `json:"muted"`
	}
	if err := json.Unmarshal([]byte(res.stdout), &decoded); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\ngot: %s", err, res.stdout)
	}
	if decoded.OutputID != "office-speaker" || !decoded.Muted {
		t.Errorf("unexpected decoded content: %+v", decoded)
	}
}

func TestOutputsMute_NotFound(t *testing.T) {
	srv, _ := mockOutputMutedServer(t, map[string]map[string]any{})

	res := runCLI(t, "mute", "outputs/missing-output", "--hub-url", srv.URL)

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

func TestOutputsUnmute_MissingIdentifier(t *testing.T) {
	res := runCLI(t, "unmute", "outputs/")

	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
}

func TestGroupsMute_Success(t *testing.T) {
	srv, lastBody := mockGroupMutedServer(t, map[string]map[string]any{
		"living-room": {"groupId": "living-room", "displayName": "Living Room", "outputIds": []string{"office-speaker"}, "muted": false, "enabled": true},
	})

	res := runCLI(t, "mute", "groups/living-room", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if *lastBody != `{"muted":true}` {
		t.Errorf("got request body %s, want muted=true", *lastBody)
	}
	if !strings.Contains(res.stdout, "living-room") {
		t.Errorf("expected groupId in stdout, got:\n%s", res.stdout)
	}
}

func TestGroupsUnmute_Success(t *testing.T) {
	srv, lastBody := mockGroupMutedServer(t, map[string]map[string]any{
		"living-room": {"groupId": "living-room", "displayName": "Living Room", "outputIds": []string{"office-speaker"}, "muted": true, "enabled": true},
	})

	res := runCLI(t, "unmute", "groups/living-room", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if *lastBody != `{"muted":false}` {
		t.Errorf("got request body %s, want muted=false", *lastBody)
	}
}

func TestGroupsMute_JSONOutput(t *testing.T) {
	srv, _ := mockGroupMutedServer(t, map[string]map[string]any{
		"living-room": {"groupId": "living-room", "displayName": "Living Room", "outputIds": []string{"office-speaker"}, "muted": false, "enabled": true},
	})

	res := runCLI(t, "mute", "groups/living-room", "--hub-url", srv.URL, "--json")

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	var decoded struct {
		GroupID string `json:"groupId"`
		Muted   bool   `json:"muted"`
	}
	if err := json.Unmarshal([]byte(res.stdout), &decoded); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\ngot: %s", err, res.stdout)
	}
	if decoded.GroupID != "living-room" || !decoded.Muted {
		t.Errorf("unexpected decoded content: %+v", decoded)
	}
}

func TestGroupsMute_NotFound(t *testing.T) {
	srv, _ := mockGroupMutedServer(t, map[string]map[string]any{})

	res := runCLI(t, "mute", "groups/missing-group", "--hub-url", srv.URL)

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

func TestGroupsUnmute_MissingIdentifier(t *testing.T) {
	res := runCLI(t, "unmute", "groups/")

	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
}

func TestMute_InputsRejected(t *testing.T) {
	res := runCLI(t, "mute", "inputs/spotify-1")

	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
}

func TestMute_RoutesRejected(t *testing.T) {
	res := runCLI(t, "mute", "routes/route-abc-123")

	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
}
