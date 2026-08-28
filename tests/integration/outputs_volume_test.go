package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func mockOutputVolumeServer(t *testing.T, outputsByID map[string]map[string]any) (*httptest.Server, *string) {
	t.Helper()
	var lastBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/v2/outputs/"), "/volume")
		o, ok := outputsByID[id]
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "not found"})
			return
		}
		var body struct {
			Volume int `json:"volume"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		lastBody = strings.TrimSpace(func() string {
			b, _ := json.Marshal(body)
			return string(b)
		}())
		o["volume"] = body.Volume
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(o)
	}))
	t.Cleanup(srv.Close)
	return srv, &lastBody
}

func TestOutputsSetVolume_Success(t *testing.T) {
	srv, lastBody := mockOutputVolumeServer(t, map[string]map[string]any{
		"office-speaker": {"outputId": "office-speaker", "volume": 50, "updatedAt": "2026-06-22T14:30:00Z"},
	})

	res := runCLI(t, "set", "outputs/office-speaker", "volume", "75", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if *lastBody != `{"volume":75}` {
		t.Errorf("got request body %s, want volume=75", *lastBody)
	}
	if !strings.Contains(res.stdout, "office-speaker") {
		t.Errorf("expected outputId in stdout, got:\n%s", res.stdout)
	}
}

func TestOutputsSetVolume_JSONOutput(t *testing.T) {
	srv, _ := mockOutputVolumeServer(t, map[string]map[string]any{
		"office-speaker": {"outputId": "office-speaker", "volume": 50, "updatedAt": "2026-06-22T14:30:00Z"},
	})

	res := runCLI(t, "set", "outputs/office-speaker", "volume", "30", "--hub-url", srv.URL, "--json")

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	var decoded struct {
		OutputID string `json:"outputId"`
		Volume   int    `json:"volume"`
	}
	if err := json.Unmarshal([]byte(res.stdout), &decoded); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\ngot: %s", err, res.stdout)
	}
	if decoded.OutputID != "office-speaker" || decoded.Volume != 30 {
		t.Errorf("unexpected decoded content: %+v", decoded)
	}
}

func TestOutputsSetVolume_NotFound(t *testing.T) {
	srv, _ := mockOutputVolumeServer(t, map[string]map[string]any{})

	res := runCLI(t, "set", "outputs/missing-output", "volume", "50", "--hub-url", srv.URL)

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

func TestOutputsSetVolume_OutOfRangeIsUsageError(t *testing.T) {
	res := runCLI(t, "set", "outputs/office-speaker", "volume", "150")

	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
}

func TestOutputsSetVolume_NonOutputsResourceIsUsageError(t *testing.T) {
	res := runCLI(t, "set", "groups/main-floor", "volume", "50")

	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
}

func TestOutputsSetVolume_MissingIdentifier(t *testing.T) {
	res := runCLI(t, "set", "outputs/")

	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
}
