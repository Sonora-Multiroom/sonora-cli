package unit

import (
	"bytes"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/Sonora-Multiroom/sonora-cli/hub"
	"github.com/Sonora-Multiroom/sonora-cli/internal/cli/radio"
)

func TestRadioReportError_NotOffered_DiagnosesWithOneRequest(t *testing.T) {
	srv, count := extensionsServer(t, 200, `{"extensions":[{"id":"radio","status":"DISABLED"}]}`)
	var stderr bytes.Buffer

	code := radio.ReportError(&stderr, &hub.RadioNotOfferedError{}, srv.URL, false)

	if code != 15 {
		t.Fatalf("exit code = %d, want 15; stderr: %s", code, stderr.String())
	}
	want := "error: radio is not available on this hub: the radio extension is installed but disabled in the hub's configuration (hub URL: " + srv.URL + ")\n"
	if stderr.String() != want {
		t.Errorf("stderr = %q, want %q", stderr.String(), want)
	}
	if got := atomic.LoadInt32(count); got != 1 {
		t.Errorf("inventory requests = %d, want 1", got)
	}
}

func TestRadioReportError_OtherErrors_NoInventoryRequest(t *testing.T) {
	srv, count := extensionsServer(t, 200, `{"extensions":[]}`)
	var stderr bytes.Buffer

	code := radio.ReportError(&stderr, &hub.APIError{StatusCode: 400, Type: "urn:multiroom:radio:error:invalid-parameter", Detail: "limit must be between 1 and 100"}, srv.URL, false)

	if code != 6 {
		t.Fatalf("exit code = %d, want 6", code)
	}
	want := "error: limit must be between 1 and 100 (hub URL: " + srv.URL + ")\n"
	if stderr.String() != want {
		t.Errorf("stderr = %q, want %q", stderr.String(), want)
	}
	if got := atomic.LoadInt32(count); got != 0 {
		t.Errorf("inventory requests = %d, want 0", got)
	}
}

func TestRadioReportError_HubAddress(t *testing.T) {
	srv, _ := extensionsServer(t, 404, "")
	var stderr bytes.Buffer

	code := radio.ReportError(&stderr, &hub.RadioNotOfferedError{}, srv.URL, false)

	if code != 4 {
		t.Fatalf("exit code = %d, want 4; stderr: %s", code, stderr.String())
	}
	want := "error: " + srv.URL + " is not serving the Multiroom Audio Hub API: the hub URL is wrong, or the hub's control API (REST) extension is not installed or not loaded; set the correct address with --hub-url, MULTIROOM_URL, or the config file\n"
	if stderr.String() != want {
		t.Errorf("stderr = %q, want %q", stderr.String(), want)
	}
	if strings.Contains(stderr.String(), "(hub URL:") {
		t.Error("hub-address message must not carry the (hub URL: …) suffix")
	}
}

func TestRadioReportError_Verbose(t *testing.T) {
	var stderr bytes.Buffer
	err := &hub.APIError{StatusCode: 503, Type: "urn:multiroom:radio:error:catalogue-unavailable", Detail: "timed out"}

	radio.ReportError(&stderr, err, "http://hub.invalid", true)

	if !strings.Contains(stderr.String(), "detail: timed out\n") {
		t.Errorf("expected a detail line, got %q", stderr.String())
	}
}

func TestRadioReportError_VerboseUsesDiagnosisCause(t *testing.T) {
	srv, _ := extensionsServer(t, 500, "")
	var stderr bytes.Buffer

	radio.ReportError(&stderr, &hub.RadioNotOfferedError{}, srv.URL, true)

	if !strings.Contains(stderr.String(), "detail: hub returned HTTP 500\n") {
		t.Errorf("expected the inventory failure as detail, got %q", stderr.String())
	}
}

func TestRadioReportError_HubDetailIsEscaped(t *testing.T) {
	var stderr bytes.Buffer
	err := &hub.APIError{StatusCode: 400, Type: "urn:multiroom:radio:error:invalid-parameter",
		Detail: "bad \x1b[2J value \u009b and \u202e end"}

	radio.ReportError(&stderr, err, "http://hub.invalid", true)

	for _, r := range stderr.String() {
		if r == '\x1b' || (r >= 0x80 && r <= 0x9f) || r == '\u202e' {
			t.Errorf("stderr holds a raw control rune %U: %q", r, stderr.String())
		}
	}
	if !strings.Contains(stderr.String(), `\x1b[2J`) || !strings.Contains(stderr.String(), `\u009b`) || !strings.Contains(stderr.String(), `\u202e`) {
		t.Errorf("expected escaped forms in %q", stderr.String())
	}
}
