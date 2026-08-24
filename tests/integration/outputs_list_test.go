package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// binPath is the freshly built sonora binary, produced once in TestMain and
// exercised by every test in this package as a full process invocation.
var binPath string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "sonora-integration-bin")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer os.RemoveAll(dir)

	name := "sonora"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	binPath = filepath.Join(dir, name)

	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	repoRoot := filepath.Join(wd, "..", "..")

	cmd := exec.Command("go", "build", "-o", binPath, "./cmd/sonora")
	cmd.Dir = repoRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "building sonora binary failed: %v\n%s", err, out)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

type cliResult struct {
	stdout   string
	stderr   string
	exitCode int
}

// runCLI executes the built sonora binary with args, isolated from any real
// user config file via an empty HOME/USERPROFILE.
func runCLI(t *testing.T, args ...string) cliResult {
	t.Helper()
	emptyHome := t.TempDir()

	cmd := exec.Command(binPath, args...)
	cmd.Env = append(os.Environ(), "HOME="+emptyHome, "USERPROFILE="+emptyHome, "MULTIROOM_URL=")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("failed to run sonora binary: %v", err)
		}
	}
	return cliResult{stdout: stdout.String(), stderr: stderr.String(), exitCode: exitCode}
}

func mockOutputsServer(t *testing.T, body string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv
}

// mockOutputsServerFilteringDisabled mirrors the real hub's documented
// behavior (api/openapi.json: "By default only enabled outputs are
// returned"): it drops disabled outputs unless includeDisabled=true is
// present on the request, so the CLI is exercised against a server that
// actually performs the filtering server-side.
func mockOutputsServerFilteringDisabled(t *testing.T, allOutputs []map[string]any) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		includeDisabled := r.URL.Query().Get("includeDisabled") == "true"
		var filtered []map[string]any
		for _, o := range allOutputs {
			if includeDisabled || o["enabled"] == true {
				filtered = append(filtered, o)
			}
		}
		if filtered == nil {
			filtered = []map[string]any{}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(filtered)
	}))
	t.Cleanup(srv.Close)
	return srv
}

var twoEnabledOneDisabled = []map[string]any{
	{"outputId": "office-speaker", "displayName": "Office Speaker", "volume": 75, "muted": false, "available": true, "enabled": true},
	{"outputId": "kitchen-speaker", "displayName": "Kitchen Speaker", "volume": 50, "muted": true, "available": true, "enabled": true},
	{"outputId": "garage-speaker", "displayName": "Garage Speaker", "volume": 0, "muted": false, "available": false, "enabled": false},
}

func TestOutputsList_DefaultEnabledOnlyYAML(t *testing.T) {
	srv := mockOutputsServerFilteringDisabled(t, twoEnabledOneDisabled)

	res := runCLI(t, "outputs", "list", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if !strings.Contains(res.stdout, "office-speaker") || !strings.Contains(res.stdout, "kitchen-speaker") {
		t.Errorf("expected enabled outputs in stdout, got:\n%s", res.stdout)
	}
	if strings.Contains(res.stdout, "garage-speaker") {
		t.Errorf("disabled output should not appear by default, got:\n%s", res.stdout)
	}
	for _, field := range []string{"outputId", "displayName", "volume", "muted", "available", "enabled"} {
		if !strings.Contains(res.stdout, field) {
			t.Errorf("expected field %q in stdout, got:\n%s", field, res.stdout)
		}
	}
}

func TestOutputsList_ZeroOutputsIsUnambiguousSuccess(t *testing.T) {
	srv := mockOutputsServer(t, `[]`)

	res := runCLI(t, "outputs", "list", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	lower := strings.ToLower(res.stdout)
	if !strings.Contains(lower, "no outputs") {
		t.Errorf("expected an unambiguous 'no outputs' indication, got:\n%s", res.stdout)
	}
}

func TestOutputsList_UnreachableHub(t *testing.T) {
	// Port 1 is a reserved/unassigned port; connections are refused
	// immediately rather than hanging.
	res := runCLI(t, "outputs", "list", "--hub-url", "http://127.0.0.1:1")

	if res.exitCode != 4 {
		t.Fatalf("exit code = %d, want 4; stderr: %s", res.exitCode, res.stderr)
	}
	if res.stdout != "" {
		t.Errorf("expected empty stdout on failure, got:\n%s", res.stdout)
	}
	if !strings.Contains(res.stderr, "127.0.0.1:1") {
		t.Errorf("expected stderr to mention the hub URL, got:\n%s", res.stderr)
	}
}

func TestOutputsList_HubNon2xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	res := runCLI(t, "outputs", "list", "--hub-url", srv.URL)

	if res.exitCode != 3 {
		t.Fatalf("exit code = %d, want 3; stderr: %s", res.exitCode, res.stderr)
	}
	if res.stdout != "" {
		t.Errorf("expected empty stdout on failure, got:\n%s", res.stdout)
	}
}

func TestOutputsList_MalformedResponse(t *testing.T) {
	srv := mockOutputsServer(t, `{"not": "a list"}`)

	res := runCLI(t, "outputs", "list", "--hub-url", srv.URL)

	if res.exitCode != 3 {
		t.Fatalf("exit code = %d, want 3; stderr: %s", res.exitCode, res.stderr)
	}
	if res.stdout != "" {
		t.Errorf("expected empty stdout on failure, got:\n%s", res.stdout)
	}
}

func TestOutputsList_UnknownFlag(t *testing.T) {
	res := runCLI(t, "outputs", "list", "--unknown-flag")

	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
	if res.stderr == "" {
		t.Error("expected a usage message on stderr")
	}
}

func TestOutputsList_IncludeDisabledShowsBothStates(t *testing.T) {
	srv := mockOutputsServerFilteringDisabled(t, []map[string]any{
		{"outputId": "office-speaker", "displayName": "Office Speaker", "volume": 75, "muted": false, "available": true, "enabled": true},
		{"outputId": "garage-speaker", "displayName": "Garage Speaker", "volume": 0, "muted": false, "available": false, "enabled": false},
	})

	res := runCLI(t, "outputs", "list", "--hub-url", srv.URL, "--include-disabled")

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if !strings.Contains(res.stdout, "office-speaker") || !strings.Contains(res.stdout, "garage-speaker") {
		t.Errorf("expected both outputs in stdout, got:\n%s", res.stdout)
	}
	if !strings.Contains(res.stdout, "enabled: true") || !strings.Contains(res.stdout, "enabled: false") {
		t.Errorf("expected both enabled states visible, got:\n%s", res.stdout)
	}
}

func TestOutputsList_JSONOutput(t *testing.T) {
	srv := mockOutputsServer(t, `[{"outputId":"office-speaker","displayName":"Office Speaker","volume":75,"muted":false,"available":true,"enabled":true}]`)

	res := runCLI(t, "outputs", "list", "--hub-url", srv.URL, "--json")

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	var decoded struct {
		Outputs []struct {
			OutputID    string `json:"outputId"`
			DisplayName string `json:"displayName"`
			Volume      int    `json:"volume"`
			Muted       bool   `json:"muted"`
			Available   bool   `json:"available"`
			Enabled     bool   `json:"enabled"`
		} `json:"outputs"`
	}
	if err := json.Unmarshal([]byte(res.stdout), &decoded); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\ngot: %s", err, res.stdout)
	}
	if len(decoded.Outputs) != 1 || decoded.Outputs[0].OutputID != "office-speaker" {
		t.Errorf("unexpected decoded content: %+v", decoded)
	}
}

// TestOutputsList_AllFlagCombinations covers all 8 combinations of the
// three independent boolean flags from contracts/cli-outputs-list.md's
// example table (--include-disabled, --json, --verbose): every combination
// must succeed (exit 0) against a healthy hub, show disabled outputs iff
// --include-disabled is set, and use the right output format.
func TestOutputsList_AllFlagCombinations(t *testing.T) {
	srv := mockOutputsServerFilteringDisabled(t, twoEnabledOneDisabled)

	for _, includeDisabled := range []bool{false, true} {
		for _, jsonOut := range []bool{false, true} {
			for _, verbose := range []bool{false, true} {
				name := fmt.Sprintf("include-disabled=%v,json=%v,verbose=%v", includeDisabled, jsonOut, verbose)
				t.Run(name, func(t *testing.T) {
					args := []string{"outputs", "list", "--hub-url", srv.URL}
					if includeDisabled {
						args = append(args, "--include-disabled")
					}
					if jsonOut {
						args = append(args, "--json")
					}
					if verbose {
						args = append(args, "--verbose")
					}
					res := runCLI(t, args...)

					if res.exitCode != 0 {
						t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
					}

					hasDisabled := strings.Contains(res.stdout, "garage-speaker")
					if hasDisabled != includeDisabled {
						t.Errorf("garage-speaker (disabled) present=%v, want %v", hasDisabled, includeDisabled)
					}

					if jsonOut {
						var decoded struct {
							Outputs []json.RawMessage `json:"outputs"`
						}
						if err := json.Unmarshal([]byte(res.stdout), &decoded); err != nil {
							t.Errorf("expected valid JSON with --json, got error %v; stdout: %s", err, res.stdout)
						}
					} else if !strings.HasPrefix(res.stdout, "outputs:") && !strings.HasPrefix(res.stdout, "# no outputs found") {
						t.Errorf("expected YAML output without --json, got: %s", res.stdout)
					}
				})
			}
		}
	}
}

func TestOutputsList_JSONZeroOutputs(t *testing.T) {
	srv := mockOutputsServer(t, `[]`)

	res := runCLI(t, "outputs", "list", "--hub-url", srv.URL, "--json")

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	var decoded struct {
		Outputs []json.RawMessage `json:"outputs"`
	}
	if err := json.Unmarshal([]byte(res.stdout), &decoded); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\ngot: %s", err, res.stdout)
	}
	if decoded.Outputs == nil || len(decoded.Outputs) != 0 {
		t.Errorf("expected an explicit empty outputs array, got: %s", res.stdout)
	}
}
