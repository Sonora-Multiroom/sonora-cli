package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// mockWarmHub fakes the TTS warm and speak endpoints with a real cache keyed
// by text and voice options, and records every path it was asked for.
type mockWarmHub struct {
	mu sync.Mutex

	cache map[string]bool
	paths []string
	body  map[string]any

	warmStatus int
	warmBody   map[string]any
	warmDelay  time.Duration
}

func warmKey(b map[string]any) string {
	parts := []string{}
	for _, k := range []string{"text", "providerName", "voice", "language"} {
		v, _ := b[k].(string)
		parts = append(parts, k+"="+v)
	}
	return strings.Join(parts, "|")
}

func newMockWarmHub(t *testing.T) (*httptest.Server, *mockWarmHub) {
	t.Helper()
	m := &mockWarmHub{cache: map[string]bool{}}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		m.mu.Lock()
		m.paths = append(m.paths, r.URL.Path)
		m.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/api/tts/cache/warm":
			m.mu.Lock()
			m.body = body
			status, respBody, delay := m.warmStatus, m.warmBody, m.warmDelay
			key := warmKey(body)
			hit := m.cache[key]
			if status == 0 {
				m.cache[key] = true
			}
			m.mu.Unlock()
			if delay > 0 {
				time.Sleep(delay)
			}
			if status != 0 {
				w.WriteHeader(status)
				if respBody != nil {
					_ = json.NewEncoder(w).Encode(respBody)
				}
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"cacheHit": hit})

		case "/api/tts/speak":
			m.mu.Lock()
			key := warmKey(body)
			hit := m.cache[key]
			m.cache[key] = true
			m.mu.Unlock()
			w.WriteHeader(http.StatusAccepted)
			_ = json.NewEncoder(w).Encode(map[string]any{"announcementId": "a1", "cacheHit": hit, "queueDepth": 1})

		case "/api/v2/extensions":
			_ = json.NewEncoder(w).Encode(map[string]any{"extensions": []map[string]any{}})

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)
	return srv, m
}

func (m *mockWarmHub) count(path string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	n := 0
	for _, p := range m.paths {
		if p == path {
			n++
		}
	}
	return n
}

// playedNothing reports that no request reached anything but the warm
// endpoint (and, for a diagnosis, the extension inventory).
func (m *mockWarmHub) playedNothing() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, p := range m.paths {
		if p != "/api/tts/cache/warm" && p != "/api/v2/extensions" {
			return false
		}
	}
	return true
}

func TestWarm_ColdThenHit_PlaysNothing(t *testing.T) {
	srv, m := newMockWarmHub(t)

	res := runCLI(t, "warm", "tts-cache", "dinner", "--hub-url", srv.URL)
	if res.exitCode != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", res.exitCode, res.stderr)
	}
	if res.stdout != "cacheHit: false\n" {
		t.Errorf("stdout = %q", res.stdout)
	}

	res = runCLI(t, "warm", "tts-cache", "dinner", "--hub-url", srv.URL)
	if res.exitCode != 0 || res.stdout != "cacheHit: true\n" {
		t.Errorf("repeat: exit %d stdout %q stderr %q", res.exitCode, res.stdout, res.stderr)
	}

	if !m.playedNothing() {
		t.Errorf("warm must reach only the warm endpoint, requests: %v", m.paths)
	}
}

func TestWarm_JSON(t *testing.T) {
	srv, _ := newMockWarmHub(t)
	res := runCLI(t, "warm", "tts-cache", "dinner", "--hub-url", srv.URL, "--json")
	if res.exitCode != 0 || res.stdout != "{\"cacheHit\":false}\n" {
		t.Errorf("exit %d stdout %q stderr %q", res.exitCode, res.stdout, res.stderr)
	}
}

func TestWarm_OptionsReachHub_FlagsBeforeAndAfterText(t *testing.T) {
	srv, m := newMockWarmHub(t)

	res := runCLI(t, "warm", "tts-cache", "--provider", "piper-local", "dinner", "--voice", "v1", "--language", "uk-UA", "--hub-url", srv.URL)
	if res.exitCode != 0 {
		t.Fatalf("exit code = %d; stderr: %s", res.exitCode, res.stderr)
	}
	if m.body["text"] != "dinner" || m.body["providerName"] != "piper-local" || m.body["voice"] != "v1" || m.body["language"] != "uk-UA" {
		t.Errorf("body = %+v", m.body)
	}
	if len(m.body) != 4 {
		t.Errorf("only the given fields should be sent, body = %+v", m.body)
	}
}

func TestWarm_StdinTrimsTrailingLineBreaks(t *testing.T) {
	srv, m := newMockWarmHub(t)

	res := runCLIWithStdin(t, "Dinner is ready\r\n\n", "warm", "tts-cache", "-", "--hub-url", srv.URL)
	if res.exitCode != 0 {
		t.Fatalf("exit code = %d; stderr: %s", res.exitCode, res.stderr)
	}
	if m.body["text"] != "Dinner is ready" {
		t.Errorf("text = %q", m.body["text"])
	}
}

func TestWarm_UsageErrors_SendNothing(t *testing.T) {
	srv, m := newMockWarmHub(t)
	cases := map[string][]string{
		"no text":        {"warm", "tts-cache"},
		"empty text":     {"warm", "tts-cache", ""},
		"blank text":     {"warm", "tts-cache", "   "},
		"extra argument": {"warm", "tts-cache", "a", "b"},
		"empty provider": {"warm", "tts-cache", "a", "--provider", ""},
		"empty voice":    {"warm", "tts-cache", "a", "--voice", ""},
		"empty language": {"warm", "tts-cache", "a", "--language", ""},
		"bad timeout":    {"warm", "tts-cache", "a", "--timeout", "soon"},
		"zero timeout":   {"warm", "tts-cache", "a", "--timeout", "0s"},
	}
	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			res := runCLI(t, append(args, "--hub-url", srv.URL)...)
			if res.exitCode != 2 {
				t.Errorf("exit code = %d, want 2; stderr: %s", res.exitCode, res.stderr)
			}
			if res.stdout != "" {
				t.Errorf("stdout = %q", res.stdout)
			}
		})
	}
	res := runCLIWithStdin(t, "\n", "warm", "tts-cache", "-", "--hub-url", srv.URL)
	if res.exitCode != 2 || !strings.Contains(res.stderr, "text must not be empty") {
		t.Errorf("blank stdin: exit %d stderr %q", res.exitCode, res.stderr)
	}
	if m.count("/api/tts/cache/warm") != 0 {
		t.Errorf("usage errors must send no warm request, got %d", m.count("/api/tts/cache/warm"))
	}
}

func TestWarm_WrongResource_IsUsageError(t *testing.T) {
	for _, args := range [][]string{{"warm"}, {"warm", "outputs"}} {
		res := runCLI(t, args...)
		if res.exitCode != 2 || !strings.Contains(res.stderr, "usage: sonora warm tts-cache <text|-> [flags]") ||
			!strings.Contains(res.stderr, "error: warm supports only tts-cache") {
			t.Errorf("%v: exit %d stderr %q", args, res.exitCode, res.stderr)
		}
	}
	res := runCLI(t, "warm", "--help")
	if res.exitCode != 0 || !strings.Contains(res.stdout, "usage: sonora warm tts-cache") {
		t.Errorf("warm --help: exit %d stdout %q", res.exitCode, res.stdout)
	}
	res = runCLI(t, "warm", "tts-cache", "--help")
	if res.exitCode != 0 || !strings.Contains(res.stdout, "usage: sonora warm tts-cache") {
		t.Errorf("warm tts-cache --help: exit %d stdout %q", res.exitCode, res.stdout)
	}
}

func TestWarm_HubErrors(t *testing.T) {
	cases := []struct {
		name   string
		status int
		code   string
		exit   int
	}{
		{"invalid", 400, "INVALID_REQUEST", 6},
		{"unknown provider", 400, "PROVIDER_NOT_FOUND", 5},
		{"provider error", 503, "PROVIDER_ERROR", 10},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			srv, m := newMockWarmHub(t)
			m.warmStatus = c.status
			m.warmBody = map[string]any{"error": c.code, "message": "hub says no"}
			res := runCLI(t, "warm", "tts-cache", "x", "--hub-url", srv.URL)
			if res.exitCode != c.exit {
				t.Fatalf("exit code = %d, want %d; stderr: %s", res.exitCode, c.exit, res.stderr)
			}
			if !strings.Contains(res.stderr, "hub says no") {
				t.Errorf("stderr = %q", res.stderr)
			}
			wantNote := "; nothing was stored, repeating the request is safe"
			if got := strings.Contains(res.stderr, wantNote); got != (c.status == 503) {
				t.Errorf("note present = %v for status %d; stderr %q", got, c.status, res.stderr)
			}
		})
	}
}

func TestWarm_CacheWriteFailed_SaysNothingStored(t *testing.T) {
	srv, m := newMockWarmHub(t)
	m.warmStatus = 503
	m.warmBody = map[string]any{"error": "CACHE_WRITE_FAILED", "message": "disk full"}

	res := runCLI(t, "warm", "tts-cache", "x", "--hub-url", srv.URL)
	if res.exitCode != 10 {
		t.Fatalf("exit code = %d, want 10; stderr: %s", res.exitCode, res.stderr)
	}
	if !strings.Contains(res.stderr, "disk full; nothing was stored, repeating the request is safe") {
		t.Errorf("stderr = %q", res.stderr)
	}
}

func TestWarm_ExtensionAbsent_Diagnosed(t *testing.T) {
	srv, m := newMockWarmHub(t)
	m.warmStatus = 404

	res := runCLI(t, "warm", "tts-cache", "x", "--hub-url", srv.URL)
	if res.exitCode != 13 {
		t.Fatalf("exit code = %d, want 13; stderr: %s", res.exitCode, res.stderr)
	}
	if !strings.Contains(res.stderr, "text-to-speech is not available") {
		t.Errorf("stderr = %q", res.stderr)
	}
}

func TestWarm_UnreachableHub(t *testing.T) {
	res := runCLI(t, "warm", "tts-cache", "x", "--hub-url", "http://127.0.0.1:1")
	if res.exitCode != 4 || res.stdout != "" {
		t.Errorf("exit %d stdout %q stderr %q", res.exitCode, res.stdout, res.stderr)
	}
}

func TestWarm_TimeoutFlagBoundsTheRequest(t *testing.T) {
	srv, m := newMockWarmHub(t)
	m.warmDelay = 1500 * time.Millisecond

	start := time.Now()
	res := runCLI(t, "warm", "tts-cache", "x", "--hub-url", srv.URL, "--timeout", "200ms")
	if res.exitCode != 4 {
		t.Fatalf("exit code = %d, want 4; stderr: %s", res.exitCode, res.stderr)
	}
	if time.Since(start) >= 1500*time.Millisecond {
		t.Errorf("--timeout 200ms did not bound the request, took %v", time.Since(start))
	}

	// A generous bound lets the same slow reply through.
	res = runCLI(t, "warm", "tts-cache", "x", "--hub-url", srv.URL, "--timeout", "10s")
	if res.exitCode != 0 {
		t.Errorf("with --timeout 10s: exit %d stderr %q", res.exitCode, res.stderr)
	}
}

func TestWarm_ThenSpeak_SameOptionsHitsCache(t *testing.T) {
	srv, _ := newMockWarmHub(t)

	res := runCLI(t, "warm", "tts-cache", "dinner", "--provider", "piper-local", "--voice", "v1", "--language", "uk-UA", "--hub-url", srv.URL)
	if res.exitCode != 0 {
		t.Fatalf("warm: exit %d stderr %q", res.exitCode, res.stderr)
	}

	res = runCLI(t, "speak", "dinner", "outputs/kitchen", "--provider", "piper-local", "--voice", "v1", "--language", "uk-UA", "--hub-url", srv.URL, "--json")
	if res.exitCode != 0 {
		t.Fatalf("speak: exit %d stderr %q", res.exitCode, res.stderr)
	}
	var got struct {
		CacheHit bool `json:"cacheHit"`
	}
	if err := json.Unmarshal([]byte(res.stdout), &got); err != nil || !got.CacheHit {
		t.Errorf("speak with the same options should hit the cache: %q (%v)", res.stdout, err)
	}

	res = runCLI(t, "speak", "dinner", "outputs/kitchen", "--provider", "piper-local", "--voice", "other", "--language", "uk-UA", "--hub-url", srv.URL, "--json")
	if err := json.Unmarshal([]byte(res.stdout), &got); err != nil || got.CacheHit {
		t.Errorf("speak with a different voice must not hit the cache: %q (%v)", res.stdout, err)
	}
}
