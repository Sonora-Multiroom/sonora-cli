package unit

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"sonora-cli/internal/cli/play"
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

func TestPlayRun_MissingTargetID(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := play.Run([]string{"https://stream.example.com/live.mp3"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "target-id") {
		t.Errorf("expected stderr to name the missing <target-id> argument, got: %s", stderr.String())
	}
}

func TestPlayRun_TooManyPositionalArguments(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := play.Run([]string{"uri", "target", "extra"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
}

func TestPlayRun_UnknownFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := play.Run([]string{"uri", "target", "--unknown-flag"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
}

func TestPlayRun_UnreachableHubURL(t *testing.T) {
	var stdout, stderr bytes.Buffer
	start := time.Now()
	code := play.Run([]string{"https://stream.example.com/live.mp3", "office-speaker", "--hub-url", "http://127.0.0.1:1"}, &stdout, &stderr)
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
	code := play.Run([]string{"https://stream.example.com/live.mp3", "office-speaker", "--hub-url", "http://127.0.0.1:1", "--verbose"}, &stdout, &stderr)

	if code != 4 {
		t.Fatalf("exit code = %d, want 4; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "detail:") {
		t.Errorf("expected --verbose to append raw error detail, got stderr:\n%s", stderr.String())
	}
}

func TestPlayRun_NonVerboseOmitsRawErrorDetail(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := play.Run([]string{"https://stream.example.com/live.mp3", "office-speaker", "--hub-url", "http://127.0.0.1:1"}, &stdout, &stderr)

	if code != 4 {
		t.Fatalf("exit code = %d, want 4; stderr: %s", code, stderr.String())
	}
	if strings.Contains(stderr.String(), "detail:") {
		t.Errorf("expected no raw error detail without --verbose, got stderr:\n%s", stderr.String())
	}
}

func TestPlayRun_GroupAndOutputTogether_RejectedBeforeAnyRequest(t *testing.T) {
	srv, count := countingHub(t)
	var stdout, stderr bytes.Buffer
	code := play.Run([]string{"uri", "target", "--group", "--output", "--hub-url", srv.URL}, &stdout, &stderr)

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
			code := play.Run([]string{"uri", "target", "--volume", volume, "--hub-url", srv.URL}, &stdout, &stderr)

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
			code := play.Run([]string{"uri", "target", "--volume", volume, "--hub-url", srv.URL}, &stdout, &stderr)

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
