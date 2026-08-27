package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func mockGroupsServer(t *testing.T, body string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv
}

// mockGroupsServerFilteringDisabled mirrors the hub's documented listGroups
// behavior (api/openapi.json: "By default only enabled groups are
// returned"): it drops disabled groups unless includeDisabled=true is
// present on the request.
func mockGroupsServerFilteringDisabled(t *testing.T, allGroups []map[string]any) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		includeDisabled := r.URL.Query().Get("includeDisabled") == "true"
		var filtered []map[string]any
		for _, g := range allGroups {
			if includeDisabled || g["enabled"] == true {
				filtered = append(filtered, g)
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

var oneEnabledOneDisabledGroup = []map[string]any{
	{"groupId": "living-room", "displayName": "Living Room Speakers", "outputIds": []string{"office-speaker", "bedroom-speaker"}, "muted": false, "enabled": true},
	{"groupId": "unused-group", "displayName": "Unused Group", "outputIds": []string{}, "muted": true, "enabled": false},
}

func TestGroupsList_DefaultEnabledOnlyYAML(t *testing.T) {
	srv := mockGroupsServerFilteringDisabled(t, oneEnabledOneDisabledGroup)

	// Warm up the freshly built binary once, untimed (see
	// TestOutputsGet_SuccessYAML's comment for why).
	runCLI(t, "get", "groups", "--hub-url", srv.URL)

	start := time.Now()
	res := runCLI(t, "get", "groups", "--hub-url", srv.URL)
	elapsed := time.Since(start)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if elapsed > time.Second {
		t.Errorf("expected completion under 1s (SC-001), took %v", elapsed)
	}
	if !strings.Contains(res.stdout, "living-room") {
		t.Errorf("expected enabled group in stdout, got:\n%s", res.stdout)
	}
	if strings.Contains(res.stdout, "unused-group") {
		t.Errorf("disabled group should not appear by default, got:\n%s", res.stdout)
	}
	for _, field := range []string{"groupId", "displayName", "outputIds", "muted", "enabled"} {
		if !strings.Contains(res.stdout, field) {
			t.Errorf("expected field %q in stdout, got:\n%s", field, res.stdout)
		}
	}
}

func TestGroupsList_ZeroGroupsIsUnambiguousSuccess(t *testing.T) {
	srv := mockGroupsServer(t, `[]`)

	res := runCLI(t, "get", "groups", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	lower := strings.ToLower(res.stdout)
	if !strings.Contains(lower, "no groups") {
		t.Errorf("expected an unambiguous 'no groups' indication, got:\n%s", res.stdout)
	}
}

func TestGroupsList_UnreachableHub(t *testing.T) {
	// Port 1 is a reserved/unassigned port; connections are refused
	// immediately rather than hanging.
	res := runCLI(t, "get", "groups", "--hub-url", "http://127.0.0.1:1")

	if res.exitCode != 4 {
		t.Fatalf("exit code = %d, want 4; stderr: %s", res.exitCode, res.stderr)
	}
	if res.stdout != "" {
		t.Errorf("expected empty stdout on failure, got:\n%s", res.stdout)
	}
}

func TestGroupsList_HubNon2xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	res := runCLI(t, "get", "groups", "--hub-url", srv.URL)

	if res.exitCode != 3 {
		t.Fatalf("exit code = %d, want 3; stderr: %s", res.exitCode, res.stderr)
	}
	if res.stdout != "" {
		t.Errorf("expected empty stdout on failure, got:\n%s", res.stdout)
	}
}

func TestGroupsList_UnknownFlag(t *testing.T) {
	res := runCLI(t, "get", "groups", "--unknown-flag")

	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
	if res.stderr == "" {
		t.Error("expected a usage message on stderr")
	}
}

func TestGroupsList_IncludeDisabledShowsBothStates(t *testing.T) {
	srv := mockGroupsServerFilteringDisabled(t, oneEnabledOneDisabledGroup)

	res := runCLI(t, "get", "groups", "--hub-url", srv.URL, "--include-disabled")

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if !strings.Contains(res.stdout, "living-room") || !strings.Contains(res.stdout, "unused-group") {
		t.Errorf("expected both groups in stdout, got:\n%s", res.stdout)
	}
	if !strings.Contains(res.stdout, "enabled: true") || !strings.Contains(res.stdout, "enabled: false") {
		t.Errorf("expected both enabled states visible, got:\n%s", res.stdout)
	}
}

func TestGroupsList_JSONOutput(t *testing.T) {
	srv := mockGroupsServer(t, `[{"groupId":"living-room","displayName":"Living Room Speakers","outputIds":["office-speaker","bedroom-speaker"],"muted":false,"enabled":true}]`)

	res := runCLI(t, "get", "groups", "--hub-url", srv.URL, "--json")

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	var decoded struct {
		Groups []struct {
			GroupID     string   `json:"groupId"`
			DisplayName string   `json:"displayName"`
			OutputIDs   []string `json:"outputIds"`
			Muted       bool     `json:"muted"`
			Enabled     bool     `json:"enabled"`
		} `json:"groups"`
	}
	if err := json.Unmarshal([]byte(res.stdout), &decoded); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\ngot: %s", err, res.stdout)
	}
	if len(decoded.Groups) != 1 || decoded.Groups[0].GroupID != "living-room" || len(decoded.Groups[0].OutputIDs) != 2 {
		t.Errorf("unexpected decoded content: %+v", decoded)
	}
}

func TestGroupsList_JSONZeroGroups(t *testing.T) {
	srv := mockGroupsServer(t, `[]`)

	res := runCLI(t, "get", "groups", "--hub-url", srv.URL, "--json")

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	var decoded struct {
		Groups []json.RawMessage `json:"groups"`
	}
	if err := json.Unmarshal([]byte(res.stdout), &decoded); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\ngot: %s", err, res.stdout)
	}
	if decoded.Groups == nil || len(decoded.Groups) != 0 {
		t.Errorf("expected an explicit empty groups array, got: %s", res.stdout)
	}
}
