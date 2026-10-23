package unit

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Sonora-Multiroom/sonora-cli/internal/cli/play"
)

// countingHub is a fake hub that counts every request it receives, so tests
// can assert a client-side validation short-circuit never reaches the
// network (FR-002a, FR-004).
func countingHub(t *testing.T) (*httptest.Server, *int32) {
	t.Helper()
	var count int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&count, 1)
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)
	return srv, &count
}

func TestPlayRun_Help(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := play.Run([]string{"--help"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stdout: %s", code, stdout.String())
	}
	if !strings.Contains(stdout.String(), "Flags:") {
		t.Errorf("expected a Flags: section, got stdout:\n%s", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Errorf("expected help on stdout only, got stderr:\n%s", stderr.String())
	}
}

func TestPlayRun_MissingBothArguments(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := play.Run([]string{}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "uri") {
		t.Errorf("expected stderr to name the missing <uri> argument, got: %s", stderr.String())
	}
}

func TestPlayRun_MissingTargetPath(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := play.Run([]string{"https://stream.example.com/live.mp3"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "target") {
		t.Errorf("expected stderr to name the missing target-path argument, got: %s", stderr.String())
	}
}

// TestPlayRun_TargetTrailingSlash_ReportsMissingID guards against
// respath.Parse's "malformed identifier" message (meant for an invalid id)
// masking play's own clearer "must include an id" message when the id
// segment is simply empty, e.g. a trailing slash left on "outputs/".
func TestPlayRun_TargetTrailingSlash_ReportsMissingID(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := play.Run([]string{"uri", "outputs/"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "must include an id") {
		t.Errorf("expected the missing-id message, got: %s", stderr.String())
	}
	if strings.Contains(stderr.String(), "malformed identifier") {
		t.Errorf("expected respath's malformed-identifier message not to leak through, got: %s", stderr.String())
	}
}

func TestPlayRun_TooManyPositionalArguments(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := play.Run([]string{"uri", "outputs/target", "extra"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
}

func TestPlayRun_UnknownFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := play.Run([]string{"uri", "outputs/target", "--unknown-flag"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
}

func TestPlayRun_UnreachableHubURL(t *testing.T) {
	var stdout, stderr bytes.Buffer
	start := time.Now()
	code := play.Run([]string{"https://stream.example.com/live.mp3", "outputs/office-speaker", "--hub-url", "http://127.0.0.1:1"}, &stdout, &stderr)
	elapsed := time.Since(start)

	if code != 4 {
		t.Fatalf("exit code = %d, want 4; stderr: %s", code, stderr.String())
	}
	if elapsed >= 5*time.Second {
		t.Errorf("expected the failure to return well under 5s (SC-004), took %v", elapsed)
	}
}

func TestPlayRun_VerboseAppendsRawErrorDetail(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := play.Run([]string{"https://stream.example.com/live.mp3", "outputs/office-speaker", "--hub-url", "http://127.0.0.1:1", "--verbose"}, &stdout, &stderr)

	if code != 4 {
		t.Fatalf("exit code = %d, want 4; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "detail:") {
		t.Errorf("expected --verbose to append raw error detail, got stderr:\n%s", stderr.String())
	}
}

func TestPlayRun_NonVerboseOmitsRawErrorDetail(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := play.Run([]string{"https://stream.example.com/live.mp3", "outputs/office-speaker", "--hub-url", "http://127.0.0.1:1"}, &stdout, &stderr)

	if code != 4 {
		t.Fatalf("exit code = %d, want 4; stderr: %s", code, stderr.String())
	}
	if strings.Contains(stderr.String(), "detail:") {
		t.Errorf("expected no raw error detail without --verbose, got stderr:\n%s", stderr.String())
	}
}

func TestPlayRun_OldGroupFlag_UnrecognizedFlag(t *testing.T) {
	srv, count := countingHub(t)
	var stdout, stderr bytes.Buffer
	code := play.Run([]string{"uri", "office-speaker", "--group", "--hub-url", srv.URL}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
	if got := atomic.LoadInt32(count); got != 0 {
		t.Errorf("expected zero requests to the hub, got %d", got)
	}
}

func TestPlayRun_OldOutputFlag_UnrecognizedFlag(t *testing.T) {
	srv, count := countingHub(t)
	var stdout, stderr bytes.Buffer
	code := play.Run([]string{"uri", "office-speaker", "--output", "--hub-url", srv.URL}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
	if got := atomic.LoadInt32(count); got != 0 {
		t.Errorf("expected zero requests to the hub, got %d", got)
	}
}

func TestPlayRun_OldNameFlag_UnrecognizedFlag(t *testing.T) {
	srv, count := countingHub(t)
	var stdout, stderr bytes.Buffer
	code := play.Run([]string{"uri", "outputs/office-speaker", "--name", "Radio", "--hub-url", srv.URL}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
	if got := atomic.LoadInt32(count); got != 0 {
		t.Errorf("expected zero requests to the hub, got %d", got)
	}
}

func TestPlayRun_TargetKindRestrictedToOutputsGroups(t *testing.T) {
	for _, target := range []string{"inputs/some-id", "routes/some-id"} {
		t.Run(target, func(t *testing.T) {
			srv, count := countingHub(t)
			var stdout, stderr bytes.Buffer
			code := play.Run([]string{"uri", target, "--hub-url", srv.URL}, &stdout, &stderr)

			if code != 2 {
				t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
			}
			if got := atomic.LoadInt32(count); got != 0 {
				t.Errorf("expected zero requests to the hub, got %d", got)
			}
		})
	}
}

func TestPlayRun_MalformedTargetPath_RejectedBeforeAnyRequest(t *testing.T) {
	srv, count := countingHub(t)
	var stdout, stderr bytes.Buffer
	code := play.Run([]string{"uri", "out/foo/bar", "--hub-url", srv.URL}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
	if got := atomic.LoadInt32(count); got != 0 {
		t.Errorf("expected zero requests to the hub, got %d", got)
	}
}

func TestPlayRun_VolumeOutOfRange_RejectedBeforeAnyRequest(t *testing.T) {
	for _, volume := range []string{"-1", "150"} {
		t.Run(volume, func(t *testing.T) {
			srv, count := countingHub(t)
			var stdout, stderr bytes.Buffer
			code := play.Run([]string{"uri", "outputs/target", "--volume", volume, "--hub-url", srv.URL}, &stdout, &stderr)

			if code != 6 {
				t.Fatalf("exit code = %d, want 6; stderr: %s", code, stderr.String())
			}
			if got := atomic.LoadInt32(count); got != 0 {
				t.Errorf("expected zero requests to the hub, got %d", got)
			}
			if !strings.Contains(strings.ToLower(stderr.String()), "volume") {
				t.Errorf("expected a volume range error message, got: %s", stderr.String())
			}
		})
	}
}

func TestPlayRun_VolumeBoundaryValues_ProceedToCallHub(t *testing.T) {
	for _, volume := range []string{"0", "100"} {
		t.Run(volume, func(t *testing.T) {
			srv, count := countingHub(t)
			var stdout, stderr bytes.Buffer
			code := play.Run([]string{"uri", "outputs/target", "--volume", volume, "--hub-url", srv.URL}, &stdout, &stderr)

			// countingHub returns 404 for everything, so this resolves to a
			// not-found target — the point is that a request was attempted.
			if code != 5 {
				t.Fatalf("exit code = %d, want 5; stderr: %s", code, stderr.String())
			}
			if got := atomic.LoadInt32(count); got == 0 {
				t.Error("expected at least one request to the hub for a boundary-valid volume")
			}
		})
	}
}

func TestPlayRun_HelpDescribesDisplayName(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := play.Run([]string{"--help"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", code, stderr.String())
	}
	if want := "without it the hub names a SoundCloud or YouTube link after its track"; !strings.Contains(stdout.String(), want) {
		t.Errorf("expected --display-name help to contain %q, got:\n%s", want, stdout.String())
	}
}

func TestPlayRun_HelpDescribesJoinMode(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := play.Run([]string{"--help"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", code, stderr.String())
	}
	for _, want := range []string{
		"-join-mode MODE",
		"how the playback joins what already plays on the target: MODE is REPLACE, MIX or DUCK_OTHERS (any case); without it, REPLACE",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Errorf("expected --help to contain %q, got:\n%s", want, stdout.String())
		}
	}
}

func TestPlayRun_InvalidJoinMode_UsageErrorNoRequest(t *testing.T) {
	for _, mode := range []string{"fade", "", "duck-others"} {
		t.Run(mode, func(t *testing.T) {
			srv, count := countingHub(t)
			var stdout, stderr bytes.Buffer
			code := play.Run([]string{"uri", "outputs/target", "--join-mode", mode, "--hub-url", srv.URL}, &stdout, &stderr)

			if code != 2 {
				t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
			}
			if first, _, _ := strings.Cut(stderr.String(), "\n"); first != "usage: sonora play <uri> <outputs|groups>/<id> [flags]" {
				t.Errorf("expected the usage line first on stderr, got: %s", stderr.String())
			}
			if want := `error: unknown join mode "` + mode + `"; valid modes: REPLACE, MIX, DUCK_OTHERS`; !strings.Contains(stderr.String(), want) {
				t.Errorf("expected stderr to contain %q, got: %s", want, stderr.String())
			}
			if got := atomic.LoadInt32(count); got != 0 {
				t.Errorf("expected zero requests to the hub, got %d", got)
			}
		})
	}
}

func TestPlayRun_InvalidJoinMode_CheckedAfterArguments(t *testing.T) {
	srv, count := countingHub(t)
	var stdout, stderr bytes.Buffer
	code := play.Run([]string{"uri", "--join-mode", "fade", "--hub-url", srv.URL}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "missing required argument: <target-path>") {
		t.Errorf("expected the missing-argument message, got: %s", stderr.String())
	}
	if strings.Contains(stderr.String(), "join mode") {
		t.Errorf("expected the join mode not to be checked yet, got: %s", stderr.String())
	}
	if got := atomic.LoadInt32(count); got != 0 {
		t.Errorf("expected zero requests to the hub, got %d", got)
	}
}
