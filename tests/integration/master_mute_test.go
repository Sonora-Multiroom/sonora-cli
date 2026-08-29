package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func mockMasterMuteServer(t *testing.T, initialMuted bool) (*httptest.Server, *string) {
	t.Helper()
	muted := initialMuted
	var lastMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastMethod = r.Method
		if r.Method == http.MethodPut {
			var body struct {
				Muted bool `json:"muted"`
			}
			_ = json.NewDecoder(r.Body).Decode(&body)
			muted = body.Muted
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"muted": muted})
	}))
	t.Cleanup(srv.Close)
	return srv, &lastMethod
}

func TestMasterMute_Get_Success(t *testing.T) {
	srv, lastMethod := mockMasterMuteServer(t, true)

	res := runCLI(t, "get", "master-mute", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if *lastMethod != http.MethodGet {
		t.Errorf("got method %q, want GET", *lastMethod)
	}
	if !strings.Contains(res.stdout, "muted: true") {
		t.Errorf("expected muted: true in stdout, got:\n%s", res.stdout)
	}
}

func TestMasterMute_Get_JSONOutput(t *testing.T) {
	srv, _ := mockMasterMuteServer(t, false)

	res := runCLI(t, "get", "master-mute", "--hub-url", srv.URL, "--json")

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	var decoded struct {
		Muted bool `json:"muted"`
	}
	if err := json.Unmarshal([]byte(res.stdout), &decoded); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\ngot: %s", err, res.stdout)
	}
	if decoded.Muted {
		t.Errorf("expected muted=false, got: %+v", decoded)
	}
}

func TestMasterMute_MuteAll_Success(t *testing.T) {
	srv, lastMethod := mockMasterMuteServer(t, false)

	res := runCLI(t, "mute", "all", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if *lastMethod != http.MethodPut {
		t.Errorf("got method %q, want PUT", *lastMethod)
	}
	if !strings.Contains(res.stdout, "muted: true") {
		t.Errorf("expected muted: true in stdout, got:\n%s", res.stdout)
	}
}

func TestMasterMute_UnmuteAll_Success(t *testing.T) {
	srv, lastMethod := mockMasterMuteServer(t, true)

	res := runCLI(t, "unmute", "all", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if *lastMethod != http.MethodPut {
		t.Errorf("got method %q, want PUT", *lastMethod)
	}
	if !strings.Contains(res.stdout, "muted: false") {
		t.Errorf("expected muted: false in stdout, got:\n%s", res.stdout)
	}
}

func TestMasterMute_GetWithId_Rejected(t *testing.T) {
	res := runCLI(t, "get", "master-mute/foo")

	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
}

func TestMasterMute_ListRejected(t *testing.T) {
	res := runCLI(t, "list", "master-mute")

	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
}

func TestMasterMute_MuteAllSlashForm_Rejected(t *testing.T) {
	res := runCLI(t, "mute", "all/foo")

	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
}
