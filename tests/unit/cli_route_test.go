package unit

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"sonora-cli/internal/cli/route"
)

// countingRouteHub is a fake hub that counts every request it receives, so
// tests can assert a client-side validation short-circuit never reaches the
// network (FR-002a, FR-002b).
func countingRouteHub(t *testing.T) (*httptest.Server, *int32) {
	t.Helper()
	var count int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&count, 1)
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)
	return srv, &count
}

func TestRouteRun_Help(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := route.Run([]string{"--help"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "Flags:") {
		t.Errorf("expected a Flags: section, got stderr:\n%s", stderr.String())
	}
}

func TestRouteRun_MissingBothArguments(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := route.Run([]string{}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "input") {
		t.Errorf("expected stderr to name the missing <input-path> argument, got: %s", stderr.String())
	}
}

func TestRouteRun_MissingTargetPath(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := route.Run([]string{"inputs/spotify-1"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "target") {
		t.Errorf("expected stderr to name the missing target-path argument, got: %s", stderr.String())
	}
}

func TestRouteRun_TooManyPositionalArguments(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := route.Run([]string{"inputs/spotify-1", "outputs/target", "extra"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
}

func TestRouteRun_UnknownFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := route.Run([]string{"inputs/spotify-1", "outputs/target", "--unknown-flag"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
}

func TestRouteRun_InputPathInvalidPrefix(t *testing.T) {
	srv, count := countingRouteHub(t)
	var stdout, stderr bytes.Buffer
	code := route.Run([]string{"outputs/x", "outputs/office-speaker", "--hub-url", srv.URL}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
	if got := atomic.LoadInt32(count); got != 0 {
		t.Errorf("expected zero requests to the hub, got %d", got)
	}
}

func TestRouteRun_TargetPathInvalidPrefix(t *testing.T) {
	srv, count := countingRouteHub(t)
	var stdout, stderr bytes.Buffer
	code := route.Run([]string{"inputs/spotify-1", "inputs/x", "--hub-url", srv.URL}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
	if got := atomic.LoadInt32(count); got != 0 {
		t.Errorf("expected zero requests to the hub, got %d", got)
	}
}

func TestRouteRun_BareInputNoID(t *testing.T) {
	srv, count := countingRouteHub(t)
	var stdout, stderr bytes.Buffer
	code := route.Run([]string{"inputs", "outputs/office-speaker", "--hub-url", srv.URL}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
	if got := atomic.LoadInt32(count); got != 0 {
		t.Errorf("expected zero requests to the hub, got %d", got)
	}
}

func TestRouteRun_BareTargetNoID(t *testing.T) {
	srv, count := countingRouteHub(t)
	var stdout, stderr bytes.Buffer
	code := route.Run([]string{"inputs/spotify-1", "outputs", "--hub-url", srv.URL}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
	if got := atomic.LoadInt32(count); got != 0 {
		t.Errorf("expected zero requests to the hub, got %d", got)
	}
}

func TestRouteRun_UnreachableHubURL(t *testing.T) {
	var stdout, stderr bytes.Buffer
	start := time.Now()
	code := route.Run([]string{"inputs/spotify-1", "outputs/office-speaker", "--hub-url", "http://127.0.0.1:1"}, &stdout, &stderr)
	elapsed := time.Since(start)

	if code != 4 {
		t.Fatalf("exit code = %d, want 4; stderr: %s", code, stderr.String())
	}
	if elapsed >= 5*time.Second {
		t.Errorf("expected the failure to return well under 5s, took %v", elapsed)
	}
}

func TestRouteRun_VerboseAppendsRawErrorDetail(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := route.Run([]string{"inputs/spotify-1", "outputs/office-speaker", "--hub-url", "http://127.0.0.1:1", "--verbose"}, &stdout, &stderr)

	if code != 4 {
		t.Fatalf("exit code = %d, want 4; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "detail:") {
		t.Errorf("expected --verbose to append raw error detail, got stderr:\n%s", stderr.String())
	}
}

func TestRouteRun_NonVerboseOmitsRawErrorDetail(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := route.Run([]string{"inputs/spotify-1", "outputs/office-speaker", "--hub-url", "http://127.0.0.1:1"}, &stdout, &stderr)

	if code != 4 {
		t.Fatalf("exit code = %d, want 4; stderr: %s", code, stderr.String())
	}
	if strings.Contains(stderr.String(), "detail:") {
		t.Errorf("expected no raw error detail without --verbose, got stderr:\n%s", stderr.String())
	}
}

// TestRouteRun_TargetPathAliases_MapToExpectedTargetType covers T014: both
// groups/gr aliases map to targetType OUTPUT_GROUP, both outputs/out map to
// SINGLE_OUTPUT, and both inputs/in aliases are accepted for the input path.
func TestRouteRun_TargetPathAliases_MapToExpectedTargetType(t *testing.T) {
	for _, inputArg := range []string{"inputs/spotify-1", "in/spotify-1"} {
		for _, targetArg := range []string{"outputs/x", "out/x", "groups/x", "gr/x"} {
			t.Run(inputArg+"_"+targetArg, func(t *testing.T) {
				srv, count := countingRouteHub(t)
				var stdout, stderr bytes.Buffer
				code := route.Run([]string{inputArg, targetArg, "--hub-url", srv.URL}, &stdout, &stderr)

				// countingRouteHub returns 404 for everything, so a
				// well-formed path resolves to a not-found error (11) —
				// the point is that parsing succeeded and a request was
				// attempted, not a usage error (2).
				if code != 11 {
					t.Fatalf("exit code = %d, want 11; stderr: %s", code, stderr.String())
				}
				if got := atomic.LoadInt32(count); got == 0 {
					t.Error("expected at least one request to the hub for a well-formed path")
				}
			})
		}
	}
}

func TestRouteRunTransfer_Help(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := route.RunTransfer([]string{"--help"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "Flags:") {
		t.Errorf("expected a Flags: section, got stderr:\n%s", stderr.String())
	}
}

func TestRouteRunTransfer_MissingBothArguments(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := route.RunTransfer([]string{}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "route") {
		t.Errorf("expected stderr to name the missing <route-path> argument, got: %s", stderr.String())
	}
}

func TestRouteRunTransfer_MissingTargetPath(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := route.RunTransfer([]string{"routes/route_abc123"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "target") {
		t.Errorf("expected stderr to name the missing target-path argument, got: %s", stderr.String())
	}
}

func TestRouteRunTransfer_TooManyPositionalArguments(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := route.RunTransfer([]string{"routes/route_abc123", "outputs/target", "extra"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
}

func TestRouteRunTransfer_UnknownFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := route.RunTransfer([]string{"routes/route_abc123", "outputs/target", "--unknown-flag"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
}

func TestRouteRunTransfer_RoutePathInvalidPrefix(t *testing.T) {
	srv, count := countingRouteHub(t)
	var stdout, stderr bytes.Buffer
	code := route.RunTransfer([]string{"outputs/x", "outputs/office-speaker", "--hub-url", srv.URL}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
	if got := atomic.LoadInt32(count); got != 0 {
		t.Errorf("expected zero requests to the hub, got %d", got)
	}
}

func TestRouteRunTransfer_TargetPathInvalidPrefix(t *testing.T) {
	srv, count := countingRouteHub(t)
	var stdout, stderr bytes.Buffer
	code := route.RunTransfer([]string{"routes/route_abc123", "inputs/x", "--hub-url", srv.URL}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
	if got := atomic.LoadInt32(count); got != 0 {
		t.Errorf("expected zero requests to the hub, got %d", got)
	}
}

func TestRouteRunTransfer_BareRouteNoID(t *testing.T) {
	srv, count := countingRouteHub(t)
	var stdout, stderr bytes.Buffer
	code := route.RunTransfer([]string{"routes", "outputs/office-speaker", "--hub-url", srv.URL}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
	if got := atomic.LoadInt32(count); got != 0 {
		t.Errorf("expected zero requests to the hub, got %d", got)
	}
}

func TestRouteRunTransfer_BareTargetNoID(t *testing.T) {
	srv, count := countingRouteHub(t)
	var stdout, stderr bytes.Buffer
	code := route.RunTransfer([]string{"routes/route_abc123", "outputs", "--hub-url", srv.URL}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
	if got := atomic.LoadInt32(count); got != 0 {
		t.Errorf("expected zero requests to the hub, got %d", got)
	}
}

func TestRouteRunTransfer_UnreachableHubURL(t *testing.T) {
	var stdout, stderr bytes.Buffer
	start := time.Now()
	code := route.RunTransfer([]string{"routes/route_abc123", "outputs/office-speaker", "--hub-url", "http://127.0.0.1:1"}, &stdout, &stderr)
	elapsed := time.Since(start)

	if code != 4 {
		t.Fatalf("exit code = %d, want 4; stderr: %s", code, stderr.String())
	}
	if elapsed >= 5*time.Second {
		t.Errorf("expected the failure to return well under 5s, took %v", elapsed)
	}
}

// TestRouteRunTransfer_TargetPathAliases_MapToExpectedTargetType exercises
// both the routes/rt aliases for the route path and the outputs/out,
// groups/gr aliases for the target path.
func TestRouteRunTransfer_TargetPathAliases_MapToExpectedTargetType(t *testing.T) {
	for _, routeArg := range []string{"routes/route_abc123", "rt/route_abc123"} {
		for _, targetArg := range []string{"outputs/x", "out/x", "groups/x", "gr/x"} {
			t.Run(routeArg+"_"+targetArg, func(t *testing.T) {
				srv, count := countingRouteHub(t)
				var stdout, stderr bytes.Buffer
				code := route.RunTransfer([]string{routeArg, targetArg, "--hub-url", srv.URL}, &stdout, &stderr)

				// countingRouteHub returns 404 for everything, so a
				// well-formed path resolves to a target-not-found error
				// (12) from the ResolveTarget pre-check — the point is
				// that parsing succeeded and a request was attempted, not
				// a usage error (2).
				if code != 12 {
					t.Fatalf("exit code = %d, want 12; stderr: %s", code, stderr.String())
				}
				if got := atomic.LoadInt32(count); got == 0 {
					t.Error("expected at least one request to the hub for a well-formed path")
				}
			})
		}
	}
}
