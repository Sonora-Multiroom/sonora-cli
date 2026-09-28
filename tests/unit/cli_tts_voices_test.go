package unit

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/Sonora-Multiroom/sonora-cli/internal/cli/tts"
)

// voicesHub answers every request with status and body, recording the last
// request's escaped path and query.
func voicesHub(t *testing.T, status int, body map[string]any) (*httptest.Server, *string, *url.Values) {
	t.Helper()
	var gotPath string
	var gotQuery url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotQuery = r.URL.EscapedPath(), r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(body)
	}))
	t.Cleanup(srv.Close)
	return srv, &gotPath, &gotQuery
}

func TestRunListVoices_Help(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := tts.RunListVoices([]string{"--help"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", code, stderr.String())
	}
	for _, want := range []string{"--provider", "--language", "--engine", "fullName"} {
		if !strings.Contains(stdout.String(), want) {
			t.Errorf("expected help to contain %q, got:\n%s", want, stdout.String())
		}
	}
}

func TestRunListVoices_Success_YAML(t *testing.T) {
	srv, gotPath, gotQuery := voicesHub(t, http.StatusOK, map[string]any{
		"providerName": "gemini",
		"voices":       []map[string]any{{"shortName": "Achernar", "fullName": "Achernar", "engine": "gemini-2.5-flash-tts"}},
	})

	var stdout, stderr bytes.Buffer
	code := tts.RunListVoices([]string{"--provider", "gemini", "--hub-url", srv.URL}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", code, stderr.String())
	}
	if *gotPath != "/api/tts/providers/gemini/voices" {
		t.Errorf("got path %q", *gotPath)
	}
	if len(*gotQuery) != 0 {
		t.Errorf("expected no query params, got %v", *gotQuery)
	}
	for _, want := range []string{`providerName: "gemini"`, `shortName: "Achernar"`, "language: null", "gender: null"} {
		if !strings.Contains(stdout.String(), want) {
			t.Errorf("expected stdout to contain %q, got:\n%s", want, stdout.String())
		}
	}
}

func TestRunListVoices_Filters_JSON(t *testing.T) {
	srv, _, gotQuery := voicesHub(t, http.StatusOK, map[string]any{"providerName": "google", "voices": []any{}})

	var stdout, stderr bytes.Buffer
	code := tts.RunListVoices([]string{"--provider", "google", "--language", "uk-UA", "--engine", "chirp3-hd", "--json", "--hub-url", srv.URL}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", code, stderr.String())
	}
	if gotQuery.Get("language") != "uk-UA" || gotQuery.Get("engine") != "chirp3-hd" {
		t.Errorf("unexpected query: %v", *gotQuery)
	}
	want := `{"providerName":"google","voices":[]}` + "\n"
	if stdout.String() != want {
		t.Errorf("stdout = %q, want %q", stdout.String(), want)
	}
}

func TestRunListVoices_UsageErrors_NoRequest(t *testing.T) {
	cases := map[string]struct {
		args []string
		want string
	}{
		"missing provider": {[]string{}, "--provider is required"},
		"empty provider":   {[]string{"--provider", ""}, "--provider must not be empty"},
		"blank provider":   {[]string{"--provider", "  "}, "--provider must not be empty"},
		"blank language":   {[]string{"--provider", "google", "--language", " "}, "--language must not be empty"},
		"blank engine":     {[]string{"--provider", "google", "--engine", ""}, "--engine must not be empty"},
		"positional":       {[]string{"--provider", "google", "extra"}, "unexpected argument(s)"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			srv, count := countingSpeakHub(t)
			var stdout, stderr bytes.Buffer
			code := tts.RunListVoices(append(tc.args, "--hub-url", srv.URL), &stdout, &stderr)

			if code != 2 {
				t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
			}
			if !strings.Contains(stderr.String(), tc.want) {
				t.Errorf("expected stderr to contain %q, got: %s", tc.want, stderr.String())
			}
			if stdout.Len() != 0 {
				t.Errorf("expected empty stdout, got: %s", stdout.String())
			}
			if got := atomic.LoadInt32(count); got != 0 {
				t.Errorf("expected zero requests, got %d", got)
			}
		})
	}
}

func TestRunListVoices_HubErrors(t *testing.T) {
	cases := map[string]struct {
		status   int
		code     string
		wantExit int
	}{
		"invalid request":    {http.StatusBadRequest, "INVALID_REQUEST", 6},
		"provider not found": {http.StatusBadRequest, "PROVIDER_NOT_FOUND", 5},
		"catalogue down":     {http.StatusServiceUnavailable, "VOICE_CATALOGUE_UNAVAILABLE", 10},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			srv, _, _ := voicesHub(t, tc.status, map[string]any{"error": tc.code, "message": "the hub said no"})
			var stdout, stderr bytes.Buffer
			code := tts.RunListVoices([]string{"--provider", "google", "--hub-url", srv.URL}, &stdout, &stderr)

			if code != tc.wantExit {
				t.Fatalf("exit code = %d, want %d; stderr: %s", code, tc.wantExit, stderr.String())
			}
			if !strings.Contains(stderr.String(), tc.code+": the hub said no") {
				t.Errorf("expected stderr to carry the hub's code and message, got: %s", stderr.String())
			}
			if stdout.Len() != 0 {
				t.Errorf("expected empty stdout, got: %s", stdout.String())
			}
		})
	}
}

func TestRunListVoices_404_ActiveExtension_VersionMismatch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v2/extensions" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"extensions": []map[string]any{{"id": "tts", "status": "ACTIVE"}}})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	var stdout, stderr bytes.Buffer
	code := tts.RunListVoices([]string{"--provider", "google", "--hub-url", srv.URL}, &stdout, &stderr)

	if code != 13 {
		t.Fatalf("exit code = %d, want 13; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "does not match this CLI version") {
		t.Errorf("expected the version-mismatch diagnosis, got: %s", stderr.String())
	}
}
