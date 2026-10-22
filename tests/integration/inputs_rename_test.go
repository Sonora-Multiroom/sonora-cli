package integration

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// mockInputDisplayNameServer answers PUT /api/v2/inputs/{id}/display-name,
// applying the hub's rename rule: a null or blank name restores
// originalName.
func mockInputDisplayNameServer(t *testing.T, inputsByID map[string]map[string]any, originalName string) (*httptest.Server, *string) {
	t.Helper()
	var lastBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/v2/inputs/"), "/display-name")
		i, ok := inputsByID[id]
		if !ok || r.Method != http.MethodPut {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{"detail": "Input not found: " + id})
			return
		}
		b, _ := io.ReadAll(r.Body)
		lastBody = string(b)
		var body struct {
			DisplayName *string `json:"displayName"`
		}
		_ = json.Unmarshal(b, &body)
		if body.DisplayName == nil || strings.TrimSpace(*body.DisplayName) == "" {
			i["displayName"] = originalName
		} else {
			i["displayName"] = strings.TrimSpace(*body.DisplayName)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(i)
	}))
	t.Cleanup(srv.Close)
	return srv, &lastBody
}

func renameTestInputs() map[string]map[string]any {
	return map[string]map[string]any{
		"spotify-1": {"inputId": "spotify-1", "displayName": "Spotify Stream", "uri": "u1", "enabled": true, "autoRemove": false, "source": "STATIC", "createdAt": nil, "pauseable": true},
	}
}

func TestInputsRename_Success(t *testing.T) {
	srv, lastBody := mockInputDisplayNameServer(t, renameTestInputs(), "Spotify Stream")

	res := runCLI(t, "rename", "inputs/spotify-1", "Kiss FM", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if *lastBody != `{"displayName":"Kiss FM"}` {
		t.Errorf("got request body %s, want displayName=Kiss FM", *lastBody)
	}
	if !strings.Contains(res.stdout, "Kiss FM") {
		t.Errorf("expected the new name in stdout, got:\n%s", res.stdout)
	}
}

func TestInputsRename_FlagsBeforeName(t *testing.T) {
	srv, lastBody := mockInputDisplayNameServer(t, renameTestInputs(), "Spotify Stream")

	res := runCLI(t, "rename", "inputs/spotify-1", "--hub-url", srv.URL, "Kiss FM")

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if *lastBody != `{"displayName":"Kiss FM"}` {
		t.Errorf("got request body %s, want displayName=Kiss FM", *lastBody)
	}
}

func TestInputsRename_Reset(t *testing.T) {
	inputs := renameTestInputs()
	inputs["spotify-1"]["displayName"] = "Kiss FM"
	srv, lastBody := mockInputDisplayNameServer(t, inputs, "Spotify Stream")

	res := runCLI(t, "rename", "inputs/spotify-1", "--reset", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if *lastBody != `{"displayName":null}` {
		t.Errorf("got request body %s, want displayName=null", *lastBody)
	}
	if !strings.Contains(res.stdout, "Spotify Stream") {
		t.Errorf("expected the original name in stdout, got:\n%s", res.stdout)
	}
}

func TestInputsRename_JSONOutput(t *testing.T) {
	srv, _ := mockInputDisplayNameServer(t, renameTestInputs(), "Spotify Stream")

	res := runCLI(t, "rename", "in/spotify-1", "Kiss FM", "--hub-url", srv.URL, "--json")

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
	if decoded.InputID != "spotify-1" || decoded.DisplayName != "Kiss FM" {
		t.Errorf("unexpected decoded content: %+v", decoded)
	}
}

func TestInputsRename_NotFound(t *testing.T) {
	srv, _ := mockInputDisplayNameServer(t, map[string]map[string]any{}, "")

	res := runCLI(t, "rename", "inputs/missing-input", "Kiss FM", "--hub-url", srv.URL)

	if res.exitCode != 5 {
		t.Fatalf("exit code = %d, want 5; stderr: %s", res.exitCode, res.stderr)
	}
	if !strings.Contains(strings.ToLower(res.stderr), "not found") || !strings.Contains(res.stderr, "missing-input") {
		t.Errorf("expected a clear 'input not found' message naming the identifier, got:\n%s", res.stderr)
	}
	if res.stdout != "" {
		t.Errorf("expected empty stdout on failure, got:\n%s", res.stdout)
	}
}

func TestInputsRename_NameAndResetIsUsageError(t *testing.T) {
	res := runCLI(t, "rename", "inputs/spotify-1", "Kiss FM", "--reset", "--hub-url", "http://127.0.0.1:1")

	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
	if res.stdout != "" {
		t.Errorf("expected empty stdout on a usage error, got:\n%s", res.stdout)
	}
}

func TestInputsRename_NeitherNameNorResetIsUsageError(t *testing.T) {
	res := runCLI(t, "rename", "inputs/spotify-1", "--hub-url", "http://127.0.0.1:1")

	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
}

func TestInputsRename_ExtraArgumentIsUsageError(t *testing.T) {
	res := runCLI(t, "rename", "inputs/spotify-1", "Kiss", "FM", "--hub-url", "http://127.0.0.1:1")

	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
}

func TestInputsRename_MissingIdentifier(t *testing.T) {
	res := runCLI(t, "rename", "inputs/", "Kiss FM")

	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
}

func TestInputsRename_UnsupportedResourceIsUsageError(t *testing.T) {
	res := runCLI(t, "rename", "outputs/office-speaker", "Office")

	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
}
