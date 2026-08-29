package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func mockGroupEnabledServer(t *testing.T, groupsByID map[string]map[string]any) (*httptest.Server, *string) {
	t.Helper()
	var lastBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/v2/groups/"), "/enabled")
		g, ok := groupsByID[id]
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "not found"})
			return
		}
		var body struct {
			Enabled bool `json:"enabled"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		lastBody = strings.TrimSpace(func() string {
			b, _ := json.Marshal(body)
			return string(b)
		}())
		g["enabled"] = body.Enabled
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(g)
	}))
	t.Cleanup(srv.Close)
	return srv, &lastBody
}

func TestGroupsEnable_Success(t *testing.T) {
	srv, lastBody := mockGroupEnabledServer(t, map[string]map[string]any{
		"living-room": {"groupId": "living-room", "displayName": "Living Room", "outputIds": []string{"office-speaker"}, "muted": false, "enabled": false},
	})

	res := runCLI(t, "enable", "groups/living-room", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if *lastBody != `{"enabled":true}` {
		t.Errorf("got request body %s, want enabled=true", *lastBody)
	}
	if !strings.Contains(res.stdout, "living-room") {
		t.Errorf("expected groupId in stdout, got:\n%s", res.stdout)
	}
}

func TestGroupsDisable_Success(t *testing.T) {
	srv, lastBody := mockGroupEnabledServer(t, map[string]map[string]any{
		"living-room": {"groupId": "living-room", "displayName": "Living Room", "outputIds": []string{"office-speaker"}, "muted": false, "enabled": true},
	})

	res := runCLI(t, "disable", "groups/living-room", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if *lastBody != `{"enabled":false}` {
		t.Errorf("got request body %s, want enabled=false", *lastBody)
	}
}

func TestGroupsEnable_JSONOutput(t *testing.T) {
	srv, _ := mockGroupEnabledServer(t, map[string]map[string]any{
		"living-room": {"groupId": "living-room", "displayName": "Living Room", "outputIds": []string{"office-speaker"}, "muted": false, "enabled": false},
	})

	res := runCLI(t, "enable", "groups/living-room", "--hub-url", srv.URL, "--json")

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	var decoded struct {
		GroupID string `json:"groupId"`
		Enabled bool   `json:"enabled"`
	}
	if err := json.Unmarshal([]byte(res.stdout), &decoded); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\ngot: %s", err, res.stdout)
	}
	if decoded.GroupID != "living-room" || !decoded.Enabled {
		t.Errorf("unexpected decoded content: %+v", decoded)
	}
}

func TestGroupsEnable_NotFound(t *testing.T) {
	srv, _ := mockGroupEnabledServer(t, map[string]map[string]any{})

	res := runCLI(t, "enable", "groups/missing-group", "--hub-url", srv.URL)

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

func TestGroupsDisable_MissingIdentifier(t *testing.T) {
	res := runCLI(t, "disable", "groups/")

	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
}

func TestInputsEnable_StillWorksAlongsideGroups(t *testing.T) {
	srv, lastBody := mockInputEnabledServer(t, map[string]map[string]any{
		"spotify-1": {"inputId": "spotify-1", "displayName": "Spotify Stream", "uri": "u1", "enabled": false, "autoRemove": false, "source": "STATIC", "createdAt": nil, "pauseable": true},
	})

	res := runCLI(t, "enable", "inputs/spotify-1", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if *lastBody != `{"enabled":true}` {
		t.Errorf("got request body %s, want enabled=true", *lastBody)
	}
}

func TestOutputsEnable_StillWorksAlongsideGroups(t *testing.T) {
	srv, lastBody := mockOutputEnabledServer(t, map[string]map[string]any{
		"office-speaker": {"outputId": "office-speaker", "displayName": "Office Speaker", "volume": 40, "muted": false, "available": true, "enabled": false},
	})

	res := runCLI(t, "enable", "outputs/office-speaker", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if *lastBody != `{"enabled":true}` {
		t.Errorf("got request body %s, want enabled=true", *lastBody)
	}
}
