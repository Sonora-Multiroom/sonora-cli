package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"
)

// mockTTSHub fakes /api/tts/speak, /api/tts/cache/stats, /api/tts/cache, and
// /api/v2/extensions, recording per-path request counts and the last
// request bodies/queries so tests can assert exactly what the CLI sent
// (SC-005a: one request per successful command, one extra on a 404).
type mockTTSHub struct {
	mu sync.Mutex

	speakStatus int // 0 -> default 202 success
	speakBody   map[string]any
	speakDelay  time.Duration

	cacheStatsStatus int // 0 -> default 200 empty stats
	cacheStatsBody   map[string]any

	clearStatus int // 0 -> default 204
	clearBody   map[string]any

	extStatus int // 0 -> default 200 empty inventory
	extBody   map[string]any

	lastSpeakBody  map[string]any
	lastClearQuery url.Values
	counts         map[string]int
}

func newMockTTSHub(t *testing.T) (*httptest.Server, *mockTTSHub) {
	t.Helper()
	m := &mockTTSHub{counts: map[string]int{}}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.mu.Lock()
		m.counts[r.URL.Path]++
		m.mu.Unlock()

		switch {
		case r.URL.Path == "/api/tts/speak":
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			m.mu.Lock()
			m.lastSpeakBody = body
			delay, status, respBody := m.speakDelay, m.speakStatus, m.speakBody
			m.mu.Unlock()
			if delay > 0 {
				time.Sleep(delay)
			}
			w.Header().Set("Content-Type", "application/json")
			if status != 0 {
				w.WriteHeader(status)
				if respBody != nil {
					_ = json.NewEncoder(w).Encode(respBody)
				}
				return
			}
			w.WriteHeader(http.StatusAccepted)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"announcementId": "a1b2c3d4-e5f6-7890-abcd-ef1234567890", "cacheHit": false, "queueDepth": 1,
			})

		case r.URL.Path == "/api/tts/cache/stats":
			m.mu.Lock()
			status, respBody := m.cacheStatsStatus, m.cacheStatsBody
			m.mu.Unlock()
			w.Header().Set("Content-Type", "application/json")
			if status != 0 {
				w.WriteHeader(status)
				if respBody != nil {
					_ = json.NewEncoder(w).Encode(respBody)
				}
				return
			}
			w.WriteHeader(http.StatusOK)
			if respBody != nil {
				_ = json.NewEncoder(w).Encode(respBody)
			} else {
				_ = json.NewEncoder(w).Encode(map[string]any{"totalEntries": 0, "totalSizeBytes": 0, "maxSizeBytes": 0})
			}

		case r.URL.Path == "/api/tts/cache" && r.Method == http.MethodDelete:
			m.mu.Lock()
			m.lastClearQuery = r.URL.Query()
			status, respBody := m.clearStatus, m.clearBody
			m.mu.Unlock()
			if status != 0 {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(status)
				if respBody != nil {
					_ = json.NewEncoder(w).Encode(respBody)
				}
				return
			}
			w.WriteHeader(http.StatusNoContent)

		case r.URL.Path == "/api/v2/extensions":
			m.mu.Lock()
			status, respBody := m.extStatus, m.extBody
			m.mu.Unlock()
			w.Header().Set("Content-Type", "application/json")
			if status != 0 {
				w.WriteHeader(status)
				if respBody != nil {
					_ = json.NewEncoder(w).Encode(respBody)
				}
				return
			}
			w.WriteHeader(http.StatusOK)
			if respBody != nil {
				_ = json.NewEncoder(w).Encode(respBody)
			} else {
				_ = json.NewEncoder(w).Encode(map[string]any{"extensions": []map[string]any{}})
			}

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)
	return srv, m
}

func (m *mockTTSHub) requestCount(path string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.counts[path]
}

// runCLIWithStdin runs the built sonora binary like runCLI, but pipes stdin
// into the process, for exercising `sonora speak -`.
func runCLIWithStdin(t *testing.T, stdin string, args ...string) cliResult {
	t.Helper()
	emptyHome := t.TempDir()

	cmd := exec.Command(binPath, args...)
	cmd.Env = append(os.Environ(), "HOME="+emptyHome, "USERPROFILE="+emptyHome, "MULTIROOM_URL=")
	cmd.Stdin = strings.NewReader(stdin)

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

func TestSpeak_Success_YAML_OneRequestNoInventoryLookup(t *testing.T) {
	srv, m := newMockTTSHub(t)

	res := runCLI(t, "speak", "Dinner is ready", "outputs/kitchen", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	for _, field := range []string{"announcementId", "cacheHit", "queueDepth"} {
		if !strings.Contains(res.stdout, field) {
			t.Errorf("expected field %q in stdout, got:\n%s", field, res.stdout)
		}
	}
	if got := m.requestCount("/api/tts/speak"); got != 1 {
		t.Errorf("expected exactly 1 speak request, got %d", got)
	}
	if got := m.requestCount("/api/v2/extensions"); got != 0 {
		t.Errorf("expected 0 inventory requests on success, got %d", got)
	}
}

func TestSpeak_Success_JSON(t *testing.T) {
	srv, _ := newMockTTSHub(t)

	res := runCLI(t, "speak", "Dinner is ready", "outputs/kitchen", "--hub-url", srv.URL, "--json")

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	var decoded map[string]any
	if err := json.Unmarshal([]byte(res.stdout), &decoded); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\ngot: %s", err, res.stdout)
	}
}

func TestSpeak_PipedStdin(t *testing.T) {
	srv, m := newMockTTSHub(t)

	res := runCLIWithStdin(t, "Piped hello\n", "speak", "-", "outputs/kitchen", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if m.lastSpeakBody["text"] != "Piped hello" {
		t.Errorf("expected hub to receive 'Piped hello', got: %+v", m.lastSpeakBody)
	}
}

// TestSpeak_GroupTarget covers 009-tts-commands US2 (T019): a gr/<id>
// target succeeds with the same YAML fields as US1, and the recorded body
// carries targetType OUTPUT_GROUP.
func TestSpeak_GroupTarget(t *testing.T) {
	srv, m := newMockTTSHub(t)

	res := runCLI(t, "speak", "Motion detected in garden", "gr/all-rooms", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	for _, field := range []string{"announcementId", "cacheHit", "queueDepth"} {
		if !strings.Contains(res.stdout, field) {
			t.Errorf("expected field %q in stdout, got:\n%s", field, res.stdout)
		}
	}
	if m.lastSpeakBody["targetType"] != "OUTPUT_GROUP" || m.lastSpeakBody["targetName"] != "all-rooms" {
		t.Errorf("unexpected request body: %+v", m.lastSpeakBody)
	}
}

// TestSpeak_ProviderNotFound_DistinctFromTargetNotFound covers
// 009-tts-commands US3 (T023): --provider no-such against a hub answering
// PROVIDER_NOT_FOUND exits 5, distinct from TARGET_NOT_FOUND's exit 12.
func TestSpeak_ProviderNotFound_DistinctFromTargetNotFound(t *testing.T) {
	srv, m := newMockTTSHub(t)
	m.speakStatus = 400
	m.speakBody = map[string]any{"error": "PROVIDER_NOT_FOUND", "message": "no such provider: no-such"}

	res := runCLI(t, "speak", "Hi", "outputs/kitchen", "--provider", "no-such", "--hub-url", srv.URL)

	if res.exitCode != 5 {
		t.Fatalf("exit code = %d, want 5; stderr: %s", res.exitCode, res.stderr)
	}
	if !strings.Contains(res.stderr, "PROVIDER_NOT_FOUND:") {
		t.Errorf("expected stderr to contain PROVIDER_NOT_FOUND:, got: %s", res.stderr)
	}

	m.speakBody = map[string]any{"error": "TARGET_NOT_FOUND", "message": "no such output"}
	res2 := runCLI(t, "speak", "Hi", "outputs/kitchen", "--hub-url", srv.URL)
	if res2.exitCode != 12 {
		t.Fatalf("exit code = %d, want 12; stderr: %s", res2.exitCode, res2.stderr)
	}
}

func TestSpeak_TargetNotFound(t *testing.T) {
	srv, m := newMockTTSHub(t)
	m.speakStatus = 400
	m.speakBody = map[string]any{"error": "TARGET_NOT_FOUND", "message": "no such output: kitchen"}

	res := runCLI(t, "speak", "Hi", "outputs/kitchen", "--hub-url", srv.URL)

	if res.exitCode != 12 {
		t.Fatalf("exit code = %d, want 12; stderr: %s", res.exitCode, res.stderr)
	}
	if !strings.Contains(res.stderr, "TARGET_NOT_FOUND:") || !strings.Contains(res.stderr, "no such output: kitchen") {
		t.Errorf("expected stderr to contain the code and hub message, got: %s", res.stderr)
	}
	if got := m.requestCount("/api/tts/speak"); got != 1 {
		t.Errorf("expected exactly 1 speak request (no retry), got %d", got)
	}
}

func TestSpeak_InvalidRequest_TextTooLong(t *testing.T) {
	srv, m := newMockTTSHub(t)
	m.speakStatus = 400
	m.speakBody = map[string]any{"error": "INVALID_REQUEST", "message": "text exceeds maximum length"}

	res := runCLI(t, "speak", "Hi", "outputs/kitchen", "--hub-url", srv.URL)

	if res.exitCode != 6 {
		t.Fatalf("exit code = %d, want 6; stderr: %s", res.exitCode, res.stderr)
	}
	if got := m.requestCount("/api/tts/speak"); got != 1 {
		t.Errorf("expected exactly 1 speak request (no retry), got %d", got)
	}
}

func TestSpeak_ProviderTimeout(t *testing.T) {
	srv, m := newMockTTSHub(t)
	m.speakStatus = 503
	m.speakBody = map[string]any{"error": "PROVIDER_TIMEOUT", "message": "provider did not respond in time"}

	res := runCLI(t, "speak", "Hi", "outputs/kitchen", "--hub-url", srv.URL)

	if res.exitCode != 10 {
		t.Fatalf("exit code = %d, want 10; stderr: %s", res.exitCode, res.stderr)
	}
	if got := m.requestCount("/api/tts/speak"); got != 1 {
		t.Errorf("expected exactly 1 speak request (no retry), got %d", got)
	}
}

func TestSpeak_WrongTargetKind_NoRequests(t *testing.T) {
	srv, m := newMockTTSHub(t)

	res := runCLI(t, "speak", "Hi", "routes/x", "--hub-url", srv.URL)

	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
	if got := m.requestCount("/api/tts/speak"); got != 0 {
		t.Errorf("expected zero requests, got %d", got)
	}
}

func TestSpeak_404_InventoryShowsDisabled(t *testing.T) {
	srv, m := newMockTTSHub(t)
	m.speakStatus = 404
	m.extBody = map[string]any{"extensions": []map[string]any{{"id": "tts", "status": "DISABLED"}}}

	res := runCLI(t, "speak", "Hi", "outputs/kitchen", "--hub-url", srv.URL)

	if res.exitCode != 13 {
		t.Fatalf("exit code = %d, want 13; stderr: %s", res.exitCode, res.stderr)
	}
	if !strings.Contains(res.stderr, "disabled") {
		t.Errorf("expected stderr to mention the disabled diagnosis, got: %s", res.stderr)
	}
	if got := m.requestCount("/api/v2/extensions"); got != 1 {
		t.Errorf("expected exactly 1 inventory request, got %d", got)
	}
}

func TestSpeak_404_InventoryAlso404_HubAddress(t *testing.T) {
	srv, m := newMockTTSHub(t)
	m.speakStatus = 404
	m.extStatus = 404

	res := runCLI(t, "speak", "Hi", "outputs/kitchen", "--hub-url", srv.URL)

	if res.exitCode != 4 {
		t.Fatalf("exit code = %d, want 4; stderr: %s", res.exitCode, res.stderr)
	}
	if !strings.Contains(res.stderr, srv.URL) {
		t.Errorf("expected stderr to name the hub URL, got: %s", res.stderr)
	}
}

func TestSpeak_UnreachableHub(t *testing.T) {
	res := runCLI(t, "speak", "Hi", "outputs/kitchen", "--hub-url", "http://127.0.0.1:1")

	if res.exitCode != 4 {
		t.Fatalf("exit code = %d, want 4; stderr: %s", res.exitCode, res.stderr)
	}
	if res.stdout != "" {
		t.Errorf("expected empty stdout on failure, got:\n%s", res.stdout)
	}
}

func TestSpeak_MalformedSuccessBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tts/speak" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte(`{"cacheHit":true,"queueDepth":1}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	res := runCLI(t, "speak", "Hi", "outputs/kitchen", "--hub-url", srv.URL)

	if res.exitCode != 3 {
		t.Fatalf("exit code = %d, want 3; stderr: %s", res.exitCode, res.stderr)
	}
	if res.stdout != "" {
		t.Errorf("expected empty stdout on failure, got:\n%s", res.stdout)
	}
}

// get tts-cache integration tests (009-tts-commands US4, T028).

func TestGetTTSCache_Success(t *testing.T) {
	srv, m := newMockTTSHub(t)
	m.cacheStatsBody = map[string]any{
		"totalEntries": 42, "totalSizeBytes": 15728640, "maxSizeBytes": 524288000,
		"entriesByProvider": map[string]any{"openai": 35, "piper-local": 7},
	}

	res := runCLI(t, "get", "tts-cache", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	for _, want := range []string{"totalEntries: 42", "openai", "piper-local"} {
		if !strings.Contains(res.stdout, want) {
			t.Errorf("expected stdout to contain %q, got:\n%s", want, res.stdout)
		}
	}
	if got := m.requestCount("/api/tts/cache/stats"); got != 1 {
		t.Errorf("expected exactly 1 request, got %d", got)
	}
}

func TestGetTTSCache_JSON(t *testing.T) {
	srv, m := newMockTTSHub(t)
	m.cacheStatsBody = map[string]any{"totalEntries": 1, "totalSizeBytes": 10, "maxSizeBytes": 100}

	res := runCLI(t, "get", "tts-cache", "--hub-url", srv.URL, "--json")

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	var decoded map[string]any
	if err := json.Unmarshal([]byte(res.stdout), &decoded); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\ngot: %s", err, res.stdout)
	}
}

func TestGetTTSCache_EmptyCache(t *testing.T) {
	srv, _ := newMockTTSHub(t)

	res := runCLI(t, "get", "tts-cache", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if !strings.Contains(res.stdout, "totalEntries: 0") || !strings.Contains(res.stdout, "entriesByProvider: {}") {
		t.Errorf("expected an empty-cache YAML rendering, got:\n%s", res.stdout)
	}
}

func TestGetTTSCache_404_InventoryShowsRejectedWithReason(t *testing.T) {
	srv, m := newMockTTSHub(t)
	m.cacheStatsStatus = 404
	m.extBody = map[string]any{"extensions": []map[string]any{{"id": "tts", "status": "REJECTED", "rejectionReason": "missing dependency"}}}

	res := runCLI(t, "get", "tts-cache", "--hub-url", srv.URL)

	if res.exitCode != 13 {
		t.Fatalf("exit code = %d, want 13; stderr: %s", res.exitCode, res.stderr)
	}
	if !strings.Contains(res.stderr, "missing dependency") {
		t.Errorf("expected stderr to contain the rejection reason, got: %s", res.stderr)
	}
}

func TestListTTSCache_UsageError(t *testing.T) {
	res := runCLI(t, "list", "tts-cache")

	if res.exitCode != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
	}
	if !strings.Contains(res.stderr, "use 'sonora get tts-cache'") {
		t.Errorf("expected stderr to point at 'sonora get tts-cache', got: %s", res.stderr)
	}
}

// clear tts-cache integration tests (009-tts-commands US5, T037).

func TestClearTTSCache_All(t *testing.T) {
	srv, m := newMockTTSHub(t)

	res := runCLI(t, "clear", "tts-cache", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if !strings.Contains(res.stdout, "cleared: all") {
		t.Errorf("expected 'cleared: all', got:\n%s", res.stdout)
	}
	if m.lastClearQuery.Get("providerName") != "" {
		t.Errorf("expected no providerName, got: %v", m.lastClearQuery)
	}
}

func TestClearTTSCache_WithProvider(t *testing.T) {
	srv, m := newMockTTSHub(t)

	res := runCLI(t, "clear", "tts-cache", "--provider", "openai", "--hub-url", srv.URL)

	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if !strings.Contains(res.stdout, "cleared: provider") || !strings.Contains(res.stdout, "openai") {
		t.Errorf("expected 'cleared: provider' and the provider name, got:\n%s", res.stdout)
	}
	if m.lastClearQuery.Get("providerName") != "openai" {
		t.Errorf("expected providerName=openai, got: %v", m.lastClearQuery)
	}
}

func TestClearTTSCache_TwiceBothSucceed(t *testing.T) {
	srv, _ := newMockTTSHub(t)

	res1 := runCLI(t, "clear", "tts-cache", "--hub-url", srv.URL)
	res2 := runCLI(t, "clear", "tts-cache", "--hub-url", srv.URL)

	if res1.exitCode != 0 || res2.exitCode != 0 {
		t.Fatalf("exit codes = %d, %d, want 0, 0", res1.exitCode, res2.exitCode)
	}
}

func TestClearTTSCache_UnknownProvider(t *testing.T) {
	srv, m := newMockTTSHub(t)
	m.clearStatus = 400
	m.clearBody = map[string]any{"error": "PROVIDER_NOT_FOUND", "message": "no such provider"}

	res := runCLI(t, "clear", "tts-cache", "--provider", "no-such", "--hub-url", srv.URL)

	if res.exitCode != 5 {
		t.Fatalf("exit code = %d, want 5; stderr: %s", res.exitCode, res.stderr)
	}
}

func TestClearTTSCache_404_InventoryShowsInert(t *testing.T) {
	srv, m := newMockTTSHub(t)
	m.clearStatus = 404
	m.extBody = map[string]any{"extensions": []map[string]any{{"id": "tts", "status": "INERT"}}}

	res := runCLI(t, "clear", "tts-cache", "--hub-url", srv.URL)

	if res.exitCode != 13 {
		t.Fatalf("exit code = %d, want 13; stderr: %s", res.exitCode, res.stderr)
	}
}

func TestClearCommand_UsageErrors(t *testing.T) {
	cases := [][]string{
		{"clear"},
		{"clear", "outputs/x"},
		{"clear", "routes"},
	}
	for _, args := range cases {
		t.Run(strings.Join(args, "_"), func(t *testing.T) {
			res := runCLI(t, args...)
			if res.exitCode != 2 {
				t.Fatalf("args %v: exit code = %d, want 2; stderr: %s", args, res.exitCode, res.stderr)
			}
			if !strings.Contains(res.stderr, "clear supports only tts-cache") {
				t.Errorf("expected 'clear supports only tts-cache', got: %s", res.stderr)
			}
		})
	}
}

// TestSpeak_SlowProviderStillReceived proves the 15s speak timeout (SC-003):
// the hub's own ~10s provider timeout answers before the CLI gives up, so
// the exit code is the hub's PROVIDER_TIMEOUT (10), not a CLI-side network
// timeout (4). Skipped under -short since it always takes ~11s.
func TestSpeak_SlowProviderStillReceived(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow-provider test in -short mode")
	}
	srv, m := newMockTTSHub(t)
	m.speakDelay = 11 * time.Second
	m.speakStatus = 503
	m.speakBody = map[string]any{"error": "PROVIDER_TIMEOUT", "message": "provider did not respond in time"}

	start := time.Now()
	res := runCLI(t, "speak", "Hi", "outputs/kitchen", "--hub-url", srv.URL)
	elapsed := time.Since(start)

	if res.exitCode != 10 {
		t.Fatalf("exit code = %d, want 10; stderr: %s", res.exitCode, res.stderr)
	}
	if elapsed >= 15*time.Second {
		t.Errorf("expected the hub's own timeout to answer before the 15s client bound, took %v", elapsed)
	}
}
