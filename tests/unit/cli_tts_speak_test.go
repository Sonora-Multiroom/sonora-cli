package unit

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"sonora-cli/internal/cli/tts"
)

// countingSpeakHub is a fake hub that counts every request and returns 404
// for everything (so a well-formed request resolves to "not offered"),
// letting tests assert a client-side validation short-circuit never reaches
// the network.
func countingSpeakHub(t *testing.T) (*httptest.Server, *int32) {
	t.Helper()
	var count int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&count, 1)
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)
	return srv, &count
}

// speakSuccessHub answers POST /api/tts/speak with a fixed 202 body and
// records the decoded request body and per-path request counts, so tests
// can assert exactly what was sent.
type speakSuccessHub struct {
	lastBody map[string]any
	count    int32
}

func newSpeakSuccessHub(t *testing.T) (*httptest.Server, *speakSuccessHub) {
	t.Helper()
	h := &speakSuccessHub{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&h.count, 1)
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		h.lastBody = body
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"announcementId": "a1b2c3d4-e5f6-7890-abcd-ef1234567890", "cacheHit": true, "queueDepth": 1,
		})
	}))
	t.Cleanup(srv.Close)
	return srv, h
}

// newSlowSpeakSuccessHub answers POST /api/tts/speak with a fixed 202 body,
// like newSpeakSuccessHub, but sleeps for delay before responding, so tests
// can exercise the --timeout override (009-tts-commands US6, T050).
func newSlowSpeakSuccessHub(t *testing.T, delay time.Duration) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(delay)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"announcementId": "a1b2c3d4-e5f6-7890-abcd-ef1234567890", "cacheHit": false, "queueDepth": 1,
		})
	}))
	t.Cleanup(srv.Close)
	return srv
}

// TestRunSpeak_TimeoutOverride_ShortensBelowDefault covers 009-tts-commands
// US6 (T050): a supplied --timeout actually overrides the 15s default,
// rather than being accepted and ignored, proven by a fake hub slower than
// the override but well within the default.
func TestRunSpeak_TimeoutOverride_ShortensBelowDefault(t *testing.T) {
	srv := newSlowSpeakSuccessHub(t, 300*time.Millisecond)
	var stdout, stderr bytes.Buffer
	start := time.Now()
	code := tts.RunSpeak([]string{"Hi", "outputs/kitchen", "--hub-url", srv.URL, "--timeout", "50ms"}, strings.NewReader(""), &stdout, &stderr)
	elapsed := time.Since(start)

	if code != 4 {
		t.Fatalf("exit code = %d, want 4; stderr: %s", code, stderr.String())
	}
	if elapsed >= time.Second {
		t.Errorf("expected --timeout to shorten the bound well under 1s, took %v", elapsed)
	}
}

// TestRunSpeak_NoTimeoutFlag_DefaultStillSucceeds covers 009-tts-commands
// US6 (T050): omitting --timeout keeps the 15s default, long enough for an
// ordinary slow response.
func TestRunSpeak_NoTimeoutFlag_DefaultStillSucceeds(t *testing.T) {
	srv := newSlowSpeakSuccessHub(t, 300*time.Millisecond)
	var stdout, stderr bytes.Buffer
	code := tts.RunSpeak([]string{"Hi", "outputs/kitchen", "--hub-url", srv.URL}, strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", code, stderr.String())
	}
}

// TestRunSpeak_TimeoutFlag_InvalidValues covers 009-tts-commands US6 (T050,
// FR-013a): a --timeout value that time.ParseDuration rejects, or a
// non-positive parsed value, is a usage error with zero requests sent.
func TestRunSpeak_TimeoutFlag_InvalidValues(t *testing.T) {
	for _, val := range []string{"0", "-5s", "notaduration"} {
		t.Run(val, func(t *testing.T) {
			srv, count := countingSpeakHub(t)
			var stdout, stderr bytes.Buffer
			code := tts.RunSpeak([]string{"Hi", "outputs/kitchen", "--hub-url", srv.URL, "--timeout", val}, strings.NewReader(""), &stdout, &stderr)

			if code != 2 {
				t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
			}
			if !strings.Contains(stderr.String(), "--timeout must be a positive duration") {
				t.Errorf("expected '--timeout must be a positive duration', got: %s", stderr.String())
			}
			if got := atomic.LoadInt32(count); got != 0 {
				t.Errorf("expected zero requests, got %d", got)
			}
		})
	}
}

func TestRunSpeak_Help(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := tts.RunSpeak([]string{"--help"}, strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Flags:") {
		t.Errorf("expected a Flags: section, got stdout:\n%s", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Errorf("expected help on stdout only, got stderr:\n%s", stderr.String())
	}
}

func TestRunSpeak_Success_YAML(t *testing.T) {
	srv, h := newSpeakSuccessHub(t)
	var stdout, stderr bytes.Buffer
	code := tts.RunSpeak([]string{"Dinner is ready", "outputs/kitchen", "--hub-url", srv.URL}, strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", code, stderr.String())
	}
	want := "announcementId: \"a1b2c3d4-e5f6-7890-abcd-ef1234567890\"\ncacheHit: true\nqueueDepth: 1\n"
	if stdout.String() != want {
		t.Errorf("stdout = %q, want %q", stdout.String(), want)
	}
	if h.lastBody["targetType"] != "SINGLE_OUTPUT" || h.lastBody["targetName"] != "kitchen" || h.lastBody["text"] != "Dinner is ready" {
		t.Errorf("unexpected request body: %+v", h.lastBody)
	}
	if got := atomic.LoadInt32(&h.count); got != 1 {
		t.Errorf("expected exactly 1 request, got %d", got)
	}
}

func TestRunSpeak_Success_JSON(t *testing.T) {
	srv, _ := newSpeakSuccessHub(t)
	var stdout, stderr bytes.Buffer
	code := tts.RunSpeak([]string{"Dinner is ready", "outputs/kitchen", "--hub-url", srv.URL, "--json"}, strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", code, stderr.String())
	}
	var decoded map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &decoded); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\ngot: %s", err, stdout.String())
	}
	if len(decoded) != 3 {
		t.Errorf("expected exactly 3 keys, got: %+v", decoded)
	}
}

func TestRunSpeak_FlagsBeforeBetweenAfterPositionals(t *testing.T) {
	srv, _ := newSpeakSuccessHub(t)

	cases := [][]string{
		{"--hub-url", srv.URL, "Hi", "outputs/kitchen"},
		{"Hi", "--hub-url", srv.URL, "outputs/kitchen"},
		{"Hi", "outputs/kitchen", "--hub-url", srv.URL},
	}
	for _, args := range cases {
		var stdout, stderr bytes.Buffer
		code := tts.RunSpeak(args, strings.NewReader(""), &stdout, &stderr)
		if code != 0 {
			t.Fatalf("args %v: exit code = %d, want 0; stderr: %s", args, code, stderr.String())
		}
	}
}

func TestRunSpeak_StdinText_TrimsTrailingCRLF(t *testing.T) {
	srv, h := newSpeakSuccessHub(t)
	var stdout, stderr bytes.Buffer
	code := tts.RunSpeak([]string{"-", "outputs/kitchen", "--hub-url", srv.URL}, strings.NewReader("Piped hello\r\n"), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", code, stderr.String())
	}
	if h.lastBody["text"] != "Piped hello" {
		t.Errorf("expected trimmed stdin text, got: %+v", h.lastBody["text"])
	}
}

func TestRunSpeak_StdinText_KeepsInnerNewlines(t *testing.T) {
	srv, h := newSpeakSuccessHub(t)
	var stdout, stderr bytes.Buffer
	code := tts.RunSpeak([]string{"-", "outputs/kitchen", "--hub-url", srv.URL}, strings.NewReader("Line one\nLine two\n"), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", code, stderr.String())
	}
	if h.lastBody["text"] != "Line one\nLine two" {
		t.Errorf("expected inner newline preserved, got: %q", h.lastBody["text"])
	}
}

func TestRunSpeak_StdinText_EmptyOrWhitespace(t *testing.T) {
	for _, in := range []string{"", "   \n", "\r\n"} {
		t.Run(in, func(t *testing.T) {
			srv, count := countingSpeakHub(t)
			var stdout, stderr bytes.Buffer
			code := tts.RunSpeak([]string{"-", "outputs/kitchen", "--hub-url", srv.URL}, strings.NewReader(in), &stdout, &stderr)

			if code != 2 {
				t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
			}
			if got := atomic.LoadInt32(count); got != 0 {
				t.Errorf("expected zero requests, got %d", got)
			}
		})
	}
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("read failed") }

func TestRunSpeak_StdinReadError(t *testing.T) {
	srv, count := countingSpeakHub(t)
	var stdout, stderr bytes.Buffer
	code := tts.RunSpeak([]string{"-", "outputs/kitchen", "--hub-url", srv.URL}, failingReader{}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "could not read text from standard input") {
		t.Errorf("expected a stdin-read error message, got: %s", stderr.String())
	}
	if got := atomic.LoadInt32(count); got != 0 {
		t.Errorf("expected zero requests, got %d", got)
	}
}

func TestRunSpeak_LiteralEmptyOrWhitespaceText(t *testing.T) {
	for _, text := range []string{"", "   "} {
		t.Run(text, func(t *testing.T) {
			srv, count := countingSpeakHub(t)
			var stdout, stderr bytes.Buffer
			code := tts.RunSpeak([]string{text, "outputs/kitchen", "--hub-url", srv.URL}, strings.NewReader(""), &stdout, &stderr)

			if code != 2 {
				t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
			}
			if !strings.Contains(stderr.String(), "text must not be empty") {
				t.Errorf("expected 'text must not be empty', got: %s", stderr.String())
			}
			if got := atomic.LoadInt32(count); got != 0 {
				t.Errorf("expected zero requests, got %d", got)
			}
		})
	}
}

// TestRunSpeak_DashTerminator_EverythingAfterIsPositional proves "--" ends
// flag parsing entirely: every argument after it, including one shaped like
// a flag, becomes positional — here that leaves 4 positionals ("-5 degrees
// outside", "outputs/kitchen", "--hub-url", srv.URL), a usage error.
func TestRunSpeak_DashTerminator_EverythingAfterIsPositional(t *testing.T) {
	srv, _ := newSpeakSuccessHub(t)
	var stdout, stderr bytes.Buffer
	code := tts.RunSpeak([]string{"--", "-5 degrees outside", "outputs/kitchen", "--hub-url", srv.URL}, strings.NewReader(""), &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
}

func TestRunSpeak_DashTerminator_TextSpokenAsIs(t *testing.T) {
	srv, h := newSpeakSuccessHub(t)
	var stdout, stderr bytes.Buffer
	code := tts.RunSpeak([]string{"--hub-url", srv.URL, "--", "-5 degrees outside", "outputs/kitchen"}, strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", code, stderr.String())
	}
	if h.lastBody["text"] != "-5 degrees outside" {
		t.Errorf("expected literal dash-prefixed text, got: %+v", h.lastBody["text"])
	}
}

func TestRunSpeak_DashTerminator_TrailingFlagLikeArgTreatedAsPositional(t *testing.T) {
	srv, count := countingSpeakHub(t)
	var stdout, stderr bytes.Buffer
	code := tts.RunSpeak([]string{"--hub-url", srv.URL, "--", "-5 degrees", "outputs/kitchen", "--json"}, strings.NewReader(""), &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "unexpected argument") {
		t.Errorf("expected an unexpected-argument usage error, got: %s", stderr.String())
	}
	if got := atomic.LoadInt32(count); got != 0 {
		t.Errorf("expected zero requests, got %d", got)
	}
}

func TestRunSpeak_MissingText(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := tts.RunSpeak([]string{}, strings.NewReader(""), &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "<text>") {
		t.Errorf("expected stderr to name the missing <text> argument, got: %s", stderr.String())
	}
}

func TestRunSpeak_MissingTarget(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := tts.RunSpeak([]string{"Hi"}, strings.NewReader(""), &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "target") {
		t.Errorf("expected stderr to name the missing target argument, got: %s", stderr.String())
	}
}

func TestRunSpeak_ExtraPositional(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := tts.RunSpeak([]string{"Hi", "outputs/kitchen", "extra"}, strings.NewReader(""), &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "unexpected argument") {
		t.Errorf("expected an unexpected-argument usage error, got: %s", stderr.String())
	}
}

func TestRunSpeak_PositionalsInWrongOrder(t *testing.T) {
	srv, count := countingSpeakHub(t)
	var stdout, stderr bytes.Buffer
	code := tts.RunSpeak([]string{"outputs/kitchen", "Hi", "--hub-url", srv.URL}, strings.NewReader(""), &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "Hi") {
		t.Errorf("expected the respath error to name %q, got: %s", "Hi", stderr.String())
	}
	if got := atomic.LoadInt32(count); got != 0 {
		t.Errorf("expected zero requests, got %d", got)
	}
}

func TestRunSpeak_TargetTable(t *testing.T) {
	cases := []struct {
		name       string
		target     string
		wantExit   int
		wantTarget string // targetType expected, when wantExit == 0
		wantName   string
		wantMsg    string // substring expected in stderr, when wantExit != 0
	}{
		{name: "out alias", target: "out/kitchen", wantExit: 0, wantTarget: "SINGLE_OUTPUT", wantName: "kitchen"},
		{name: "routes wrong kind", target: "routes/x", wantExit: 2, wantMsg: "speak target must be outputs/<id> or groups/<id>"},
		{name: "inputs wrong kind", target: "inputs/x", wantExit: 2, wantMsg: "speak target must be outputs/<id> or groups/<id>"},
		{name: "outputs missing id", target: "outputs", wantExit: 2, wantMsg: "must include an id"},
		{name: "gr missing id", target: "gr", wantExit: 2, wantMsg: "must include an id"},
		{name: "unrecognized resource", target: "bogus/x", wantExit: 2, wantMsg: "unrecognized resource"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			srv, h := newSpeakSuccessHub(t)
			var stdout, stderr bytes.Buffer
			code := tts.RunSpeak([]string{"Hi", c.target, "--hub-url", srv.URL}, strings.NewReader(""), &stdout, &stderr)

			if code != c.wantExit {
				t.Fatalf("exit code = %d, want %d; stderr: %s", code, c.wantExit, stderr.String())
			}
			if c.wantExit == 0 {
				if h.lastBody["targetType"] != c.wantTarget || h.lastBody["targetName"] != c.wantName {
					t.Errorf("unexpected request body: %+v", h.lastBody)
				}
			} else {
				if !strings.Contains(stderr.String(), c.wantMsg) {
					t.Errorf("expected stderr to contain %q, got: %s", c.wantMsg, stderr.String())
				}
				if h.count != 0 {
					t.Errorf("expected zero requests, got %d", h.count)
				}
			}
		})
	}
}

// TestRunSpeak_GroupTargets covers 009-tts-commands US2 (T018): groups/<id>
// and its alias gr/<id> both map to OUTPUT_GROUP.
func TestRunSpeak_GroupTargets(t *testing.T) {
	for _, target := range []string{"groups/all-rooms", "gr/all-rooms"} {
		t.Run(target, func(t *testing.T) {
			srv, h := newSpeakSuccessHub(t)
			var stdout, stderr bytes.Buffer
			code := tts.RunSpeak([]string{"Motion detected", target, "--hub-url", srv.URL}, strings.NewReader(""), &stdout, &stderr)

			if code != 0 {
				t.Fatalf("exit code = %d, want 0; stderr: %s", code, stderr.String())
			}
			if h.lastBody["targetType"] != "OUTPUT_GROUP" || h.lastBody["targetName"] != "all-rooms" {
				t.Errorf("unexpected request body: %+v", h.lastBody)
			}
		})
	}
}

// TestRunSpeak_ProviderVoiceLanguage_ForwardedIndependently covers
// 009-tts-commands US3 (T022): each of --provider/--voice/--language,
// given alone, is forwarded, and the other optional keys stay absent.
func TestRunSpeak_ProviderVoiceLanguage_ForwardedIndependently(t *testing.T) {
	cases := []struct {
		flag  string
		value string
		key   string
	}{
		{"--provider", "piper-local", "providerName"},
		{"--voice", "en_US-ryan-medium", "voice"},
		{"--language", "en-US", "language"},
	}
	for _, c := range cases {
		t.Run(c.flag, func(t *testing.T) {
			srv, h := newSpeakSuccessHub(t)
			var stdout, stderr bytes.Buffer
			code := tts.RunSpeak([]string{"Hi", "outputs/kitchen", c.flag, c.value, "--hub-url", srv.URL}, strings.NewReader(""), &stdout, &stderr)

			if code != 0 {
				t.Fatalf("exit code = %d, want 0; stderr: %s", code, stderr.String())
			}
			if h.lastBody[c.key] != c.value {
				t.Errorf("expected %s=%q in request body, got: %+v", c.key, c.value, h.lastBody)
			}
			for _, other := range []string{"providerName", "voice", "language"} {
				if other == c.key {
					continue
				}
				if _, ok := h.lastBody[other]; ok {
					t.Errorf("expected %q omitted, got body: %+v", other, h.lastBody)
				}
			}
		})
	}
}

// TestRunSpeak_ProviderVoiceLanguage_AllThree covers the full-override
// example from contracts/cli-tts.md.
func TestRunSpeak_ProviderVoiceLanguage_AllThree(t *testing.T) {
	srv, h := newSpeakSuccessHub(t)
	var stdout, stderr bytes.Buffer
	code := tts.RunSpeak([]string{
		"Hi", "out/kitchen",
		"--provider", "piper-local", "--voice", "en_US-ryan-medium", "--language", "en-US",
		"--hub-url", srv.URL,
	}, strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", code, stderr.String())
	}
	if h.lastBody["providerName"] != "piper-local" || h.lastBody["voice"] != "en_US-ryan-medium" || h.lastBody["language"] != "en-US" {
		t.Errorf("unexpected request body: %+v", h.lastBody)
	}
}

// TestRunSpeak_ProviderVoiceLanguage_EmptyValue covers 009-tts-commands
// US3 (T022, FR-004): a supplied but empty/whitespace-only value is a usage
// error, and no request is sent.
func TestRunSpeak_ProviderVoiceLanguage_EmptyValue(t *testing.T) {
	cases := []struct {
		flag, value, wantMsg string
	}{
		{"--provider", "", "--provider must not be empty"},
		{"--voice", "", "--voice must not be empty"},
		{"--language", "  ", "--language must not be empty"},
	}
	for _, c := range cases {
		t.Run(c.flag, func(t *testing.T) {
			srv, count := countingSpeakHub(t)
			var stdout, stderr bytes.Buffer
			code := tts.RunSpeak([]string{"Hi", "outputs/kitchen", c.flag, c.value, "--hub-url", srv.URL}, strings.NewReader(""), &stdout, &stderr)

			if code != 2 {
				t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
			}
			if !strings.Contains(stderr.String(), c.wantMsg) {
				t.Errorf("expected stderr to contain %q, got: %s", c.wantMsg, stderr.String())
			}
			if got := atomic.LoadInt32(count); got != 0 {
				t.Errorf("expected zero requests, got %d", got)
			}
		})
	}
}

// TestRunSpeak_Help_ListsAllSevenFlags covers 009-tts-commands US3 (T022)
// and US6 (T050), which extends it with --timeout.
func TestRunSpeak_Help_ListsAllSevenFlags(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := tts.RunSpeak([]string{"--help"}, strings.NewReader(""), &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", code, stderr.String())
	}
	for _, flag := range []string{"--hub-url", "--json", "--language", "--provider", "--timeout", "--verbose", "--voice"} {
		if !strings.Contains(stdout.String(), flag) {
			t.Errorf("expected help to list %q, got:\n%s", flag, stdout.String())
		}
	}
}

func TestRunSpeak_404_UsesReportError(t *testing.T) {
	var extCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/tts/speak":
			w.WriteHeader(http.StatusNotFound)
		case "/api/v2/extensions":
			atomic.AddInt32(&extCount, 1)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"extensions": []map[string]any{}})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	var stdout, stderr bytes.Buffer
	code := tts.RunSpeak([]string{"Hi", "outputs/kitchen", "--hub-url", srv.URL}, strings.NewReader(""), &stdout, &stderr)

	if code != 13 {
		t.Fatalf("exit code = %d, want 13; stderr: %s", code, stderr.String())
	}
	if atomic.LoadInt32(&extCount) != 1 {
		t.Errorf("expected exactly 1 extensions inventory request, got %d", extCount)
	}
}
