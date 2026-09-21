package unit

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"sonora-cli/internal/cli/tts"
	"sonora-cli/internal/hub"
)

// extensionsServer fakes GET /api/v2/extensions, returning status/body and
// counting how many requests it received (research.md §5: exactly one
// inventory lookup per 404).
func extensionsServer(t *testing.T, status int, body string) (*httptest.Server, *int32) {
	t.Helper()
	var count int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/extensions" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		atomic.AddInt32(&count, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if body != "" {
			_, _ = w.Write([]byte(body))
		}
	}))
	t.Cleanup(srv.Close)
	return srv, &count
}

const ttsUnavailableHead = "text-to-speech is not available on this hub"

func TestReportError_TTSNotOffered_DiagnosisMessages(t *testing.T) {
	cases := []struct {
		name      string
		status    int
		body      string
		wantExit  int
		wantTail  string // appended to ttsUnavailableHead; "" for the generic-only message
		wantExact string // if set, the full message text (bypasses wantTail/head composition)
	}{
		{
			name:     "no tts entry, loadingEnabled true",
			status:   200,
			body:     `{"loadingEnabled":true,"extensions":[]}`,
			wantExit: 13,
			wantTail: ": the TTS extension is not installed on this hub",
		},
		{
			name:     "no tts entry, loadingEnabled absent",
			status:   200,
			body:     `{"extensions":[]}`,
			wantExit: 13,
			wantTail: ": the TTS extension is not installed on this hub",
		},
		{
			name:     "no tts entry, loadingEnabled false",
			status:   200,
			body:     `{"loadingEnabled":false,"extensions":[]}`,
			wantExit: 13,
			wantTail: ": the TTS extension is not installed on this hub (extension loading is switched off in the hub's configuration)",
		},
		{
			name:     "DISABLED",
			status:   200,
			body:     `{"extensions":[{"id":"tts","status":"DISABLED"}]}`,
			wantExit: 13,
			wantTail: ": the TTS extension is installed but disabled in the hub's configuration",
		},
		{
			name:     "REJECTED with reason",
			status:   200,
			body:     `{"extensions":[{"id":"tts","status":"REJECTED","rejectionReason":"bad jar"}]}`,
			wantExit: 13,
			wantTail: ": the TTS extension failed to load: bad jar",
		},
		{
			name:     "REJECTED with multi-line reason",
			status:   200,
			body:     `{"extensions":[{"id":"tts","status":"REJECTED","rejectionReason":"failed to init:\nstack trace line 2"}]}`,
			wantExit: 13,
			wantTail: ": the TTS extension failed to load: failed to init: stack trace line 2",
		},
		{
			name:     "REJECTED with null reason",
			status:   200,
			body:     `{"extensions":[{"id":"tts","status":"REJECTED","rejectionReason":null}]}`,
			wantExit: 13,
			wantTail: ": the TTS extension failed to load",
		},
		{
			name:     "INERT",
			status:   200,
			body:     `{"extensions":[{"id":"tts","status":"INERT"}]}`,
			wantExit: 13,
			wantTail: ": the TTS extension is loaded but inactive",
		},
		{
			name:     "ACTIVE",
			status:   200,
			body:     `{"extensions":[{"id":"tts","status":"ACTIVE"}]}`,
			wantExit: 13,
			wantTail: ": the hub's TTS API does not match this CLI version",
		},
		{
			name:     "unknown status",
			status:   200,
			body:     `{"extensions":[{"id":"tts","status":"SOMETHING_ELSE"}]}`,
			wantExit: 13,
			wantTail: "",
		},
		{
			name:     "match by id only, name is a decoy",
			status:   200,
			body:     `{"extensions":[{"id":"other","name":"tts","status":"ACTIVE"}]}`,
			wantExit: 13,
			wantTail: ": the TTS extension is not installed on this hub",
		},
		{
			name:     "inventory 500",
			status:   500,
			body:     ``,
			wantExit: 13,
			wantTail: "",
		},
		{
			name:     "malformed inventory body",
			status:   200,
			body:     `not json`,
			wantExit: 13,
			wantTail: "",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			srv, count := extensionsServer(t, c.status, c.body)
			var stderr bytes.Buffer
			code := tts.ReportError(&stderr, &hub.TTSNotOfferedError{}, srv.URL, false)

			if code != c.wantExit {
				t.Fatalf("exit code = %d, want %d; stderr: %s", code, c.wantExit, stderr.String())
			}
			wantMsg := c.wantExact
			if wantMsg == "" {
				wantMsg = ttsUnavailableHead + c.wantTail
			}
			wantLine := "error: " + wantMsg + " (hub URL: " + srv.URL + ")\n"
			if stderr.String() != wantLine {
				t.Errorf("stderr = %q, want %q", stderr.String(), wantLine)
			}
			if got := atomic.LoadInt32(count); got != 1 {
				t.Errorf("expected exactly 1 inventory request, got %d", got)
			}
		})
	}
}

func TestReportError_TTSNotOffered_InventoryAlso404_HubAddress(t *testing.T) {
	srv, count := extensionsServer(t, 404, "")
	var stderr bytes.Buffer
	code := tts.ReportError(&stderr, &hub.TTSNotOfferedError{}, srv.URL, false)

	if code != 4 {
		t.Fatalf("exit code = %d, want 4; stderr: %s", code, stderr.String())
	}
	out := stderr.String()
	if !strings.Contains(out, srv.URL) {
		t.Errorf("expected message to name the hub URL, got: %s", out)
	}
	for _, want := range []string{"--hub-url", "MULTIROOM_URL", "config file"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected message to mention %q, got: %s", want, out)
		}
	}
	if strings.Contains(out, "(hub URL:") {
		t.Errorf("expected no '(hub URL: ...)' suffix for the hub-address message, got: %s", out)
	}
	if got := atomic.LoadInt32(count); got != 1 {
		t.Errorf("expected exactly 1 inventory request, got %d", got)
	}
}

func TestReportError_TTSNotOffered_InventoryUnreachable(t *testing.T) {
	srv, count := extensionsServer(t, 200, `{"extensions":[]}`)
	srv.Close() // now unreachable

	var stderr bytes.Buffer
	code := tts.ReportError(&stderr, &hub.TTSNotOfferedError{}, srv.URL, false)

	if code != 13 {
		t.Fatalf("exit code = %d, want 13; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), ttsUnavailableHead) {
		t.Errorf("expected the generic unavailable message, got: %s", stderr.String())
	}
	if got := atomic.LoadInt32(count); got != 0 {
		t.Errorf("expected 0 inventory requests reaching an unreachable server, got %d", got)
	}
}

func TestReportError_NonNotOfferedError_NoInventoryLookup(t *testing.T) {
	srv, count := extensionsServer(t, 200, `{"extensions":[]}`)

	var stderr bytes.Buffer
	code := tts.ReportError(&stderr, &hub.TTSError{StatusCode: 404, Code: "TARGET_NOT_FOUND", Message: "no such output"}, srv.URL, false)

	if code != 12 {
		t.Fatalf("exit code = %d, want 12; stderr: %s", code, stderr.String())
	}
	want := "error: TARGET_NOT_FOUND: no such output (hub URL: " + srv.URL + ")\n"
	if stderr.String() != want {
		t.Errorf("stderr = %q, want %q", stderr.String(), want)
	}
	if got := atomic.LoadInt32(count); got != 0 {
		t.Errorf("expected zero inventory requests for a non-TTSNotOfferedError, got %d", got)
	}
}

func TestReportError_Verbose_AddsDetailLine(t *testing.T) {
	srv, _ := extensionsServer(t, 500, "")
	var stderr bytes.Buffer
	code := tts.ReportError(&stderr, &hub.TTSNotOfferedError{}, srv.URL, true)

	if code != 13 {
		t.Fatalf("exit code = %d, want 13; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "detail:") {
		t.Errorf("expected a detail: line with --verbose, got: %s", stderr.String())
	}
}

func TestReportError_NonVerbose_OmitsLookupError(t *testing.T) {
	srv, _ := extensionsServer(t, 500, "")
	var stderr bytes.Buffer
	code := tts.ReportError(&stderr, &hub.TTSNotOfferedError{}, srv.URL, false)

	if code != 13 {
		t.Fatalf("exit code = %d, want 13; stderr: %s", code, stderr.String())
	}
	if strings.Contains(stderr.String(), "detail:") {
		t.Errorf("expected no detail: line without --verbose, got: %s", stderr.String())
	}
}
