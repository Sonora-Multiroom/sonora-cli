package unit

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"sonora-cli/internal/cli/route"
)

func TestRouteRunPause_MissingIdentifier(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := route.RunPause([]string{}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
}

func TestRouteRunPause_TooManyArguments(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := route.RunPause([]string{"a", "b"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
}

func TestRouteRunPause_UnreachableHubURL(t *testing.T) {
	var stdout, stderr bytes.Buffer
	start := time.Now()
	code := route.RunPause([]string{"route_abc123", "--hub-url", "http://127.0.0.1:1"}, &stdout, &stderr)
	elapsed := time.Since(start)

	if code != 4 {
		t.Fatalf("exit code = %d, want 4; stderr: %s", code, stderr.String())
	}
	if elapsed >= 5*time.Second {
		t.Errorf("expected the failure to return well under 5s, took %v", elapsed)
	}
}

func TestRouteRunPause_VerboseAppendsRawErrorDetail(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := route.RunPause([]string{"route_abc123", "--hub-url", "http://127.0.0.1:1", "--verbose"}, &stdout, &stderr)

	if code != 4 {
		t.Fatalf("exit code = %d, want 4; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "detail:") {
		t.Errorf("expected --verbose to append raw error detail, got stderr:\n%s", stderr.String())
	}
}

func TestRouteRunResume_MissingIdentifier(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := route.RunResume([]string{}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
}

func TestRouteRunResume_TooManyArguments(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := route.RunResume([]string{"a", "b"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
}

func TestRouteRunResume_UnreachableHubURL(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := route.RunResume([]string{"route_abc123", "--hub-url", "http://127.0.0.1:1"}, &stdout, &stderr)

	if code != 4 {
		t.Fatalf("exit code = %d, want 4; stderr: %s", code, stderr.String())
	}
}
