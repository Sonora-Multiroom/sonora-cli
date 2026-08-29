package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func mockInputCreateServer(t *testing.T, existingIDs map[string]bool) (*httptest.Server, *map[string]any) {
	t.Helper()
	var lastBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		lastBody = body
		id, _ := body["inputId"].(string)
		if existingIDs[id] {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]any{"title": "Conflict", "detail": "input id already exists"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"inputId": id, "displayName": body["displayName"], "uri": body["uri"],
			"enabled": body["enabled"], "autoRemove": body["autoRemove"], "source": "EPHEMERAL",
			"createdAt": "2026-01-01T00:00:00Z", "pauseable": true,
		})
	}))
	t.Cleanup(srv.Close)
	return srv, &lastBody
}

func TestInputsCreate_Success_YAML(t *testing.T) {
	srv, lastBody := mockInputCreateServer(t, nil)

	res := runCLI(t, "create", "inputs/spotify-1", "https://stream.example.com/live.mp3", "--display-name", "Spotify Stream", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	for _, field := range []string{"inputId", "displayName", "uri", "source", "enabled", "autoRemove"} {
		if !strings.Contains(res.stdout, field) {
			t.Errorf("expected field %q in stdout, got:\n%s", field, res.stdout)
		}
	}
	if (*lastBody)["enabled"] != true || (*lastBody)["autoRemove"] != false {
		t.Errorf("expected default enabled=true/autoRemove=false, got body: %+v", *lastBody)
	}
}

func TestInputsCreate_AutoRemoveAndDisabledFlags(t *testing.T) {
	srv, lastBody := mockInputCreateServer(t, nil)

	res := runCLI(t, "create", "inputs/spotify-1", "u1", "--display-name", "Spotify", "--auto-remove", "--disabled", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if (*lastBody)["enabled"] != false || (*lastBody)["autoRemove"] != true {
		t.Errorf("expected enabled=false/autoRemove=true, got body: %+v", *lastBody)
	}
}

func TestInputsCreate_JSONOutput(t *testing.T) {
	srv, _ := mockInputCreateServer(t, nil)

	res := runCLI(t, "create", "inputs/spotify-1", "u1", "--display-name", "Spotify", "--hub-url", srv.URL, "--json")

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	var decoded struct {
		InputID     string `json:"inputId"`
		DisplayName string `json:"displayName"`
	}
	if err := json.Unmarshal([]byte(res.stdout), &decoded); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\ngot: %s", err, res.stdout)
	}
	if decoded.InputID != "spotify-1" || decoded.DisplayName != "Spotify" {
		t.Errorf("unexpected decoded content: %+v", decoded)
	}
}

func TestInputsCreate_DuplicateID_409(t *testing.T) {
	srv, _ := mockInputCreateServer(t, map[string]bool{"spotify-1": true})

	res := runCLI(t, "create", "inputs/spotify-1", "u1", "--display-name", "Spotify", "--hub-url", srv.URL)

	if res.exitCode != 3 {
		t.Fatalf("exit code = %d, want 3; stderr: %s", res.exitCode, res.stderr)
	}
	if res.stdout != "" {
		t.Errorf("expected empty stdout on failure, got:\n%s", res.stdout)
	}
}

func TestInputsCreate_MissingDisplayName(t *testing.T) {
	res := runCLI(t, "create", "inputs/spotify-1", "u1")

	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
}

func TestInputsCreate_UnsupportedResourceIsUsageError(t *testing.T) {
	res := runCLI(t, "create", "outputs/office-speaker")

	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
}

func TestInputsCreate_MissingArguments(t *testing.T) {
	res := runCLI(t, "create")

	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
}
