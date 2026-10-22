package unit

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/Sonora-Multiroom/sonora-cli/internal/cli/inputs"
)

func TestInputsRunEnable_MissingIdentifier(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := inputs.RunEnable([]string{}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
}

func TestInputsRunEnable_TooManyArguments(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := inputs.RunEnable([]string{"a", "b"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
}

func TestInputsRunEnable_UnreachableHubURL(t *testing.T) {
	var stdout, stderr bytes.Buffer
	start := time.Now()
	code := inputs.RunEnable([]string{"spotify-1", "--hub-url", "http://127.0.0.1:1"}, &stdout, &stderr)
	elapsed := time.Since(start)

	if code != 4 {
		t.Fatalf("exit code = %d, want 4; stderr: %s", code, stderr.String())
	}
	if elapsed >= 5*time.Second {
		t.Errorf("expected the failure to return well under 5s, took %v", elapsed)
	}
}

func TestInputsRunEnable_VerboseAppendsRawErrorDetail(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := inputs.RunEnable([]string{"spotify-1", "--hub-url", "http://127.0.0.1:1", "--verbose"}, &stdout, &stderr)

	if code != 4 {
		t.Fatalf("exit code = %d, want 4; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "detail:") {
		t.Errorf("expected --verbose to append raw error detail, got stderr:\n%s", stderr.String())
	}
}

func TestInputsRunDisable_MissingIdentifier(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := inputs.RunDisable([]string{}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
}

func TestInputsRunDisable_TooManyArguments(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := inputs.RunDisable([]string{"a", "b"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
}

func TestInputsRunDisable_UnreachableHubURL(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := inputs.RunDisable([]string{"spotify-1", "--hub-url", "http://127.0.0.1:1"}, &stdout, &stderr)

	if code != 4 {
		t.Fatalf("exit code = %d, want 4; stderr: %s", code, stderr.String())
	}
}

func TestInputsRunEnable_HelpDescribesPlayback(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := inputs.RunEnable([]string{"--help"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", code, stderr.String())
	}
	lines := strings.Split(stdout.String(), "\n")
	if len(lines) < 3 {
		t.Fatalf("expected at least 3 lines of help, got:\n%s", stdout.String())
	}
	if want := "usage: sonora enable inputs/<input-id> [flags]"; lines[0] != want {
		t.Errorf("line 1 = %q, want %q", lines[0], want)
	}
	if lines[1] != "" {
		t.Errorf("line 2 = %q, want empty", lines[1])
	}
	if want := "Enabling lets new routes use the input again; it starts nothing."; lines[2] != want {
		t.Errorf("line 3 = %q, want %q", lines[2], want)
	}
}

func TestInputsRunDisable_HelpDescribesPlayback(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := inputs.RunDisable([]string{"--help"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", code, stderr.String())
	}
	lines := strings.Split(stdout.String(), "\n")
	if len(lines) < 3 {
		t.Fatalf("expected at least 3 lines of help, got:\n%s", stdout.String())
	}
	if want := "usage: sonora disable inputs/<input-id> [flags]"; lines[0] != want {
		t.Errorf("line 1 = %q, want %q", lines[0], want)
	}
	if lines[1] != "" {
		t.Errorf("line 2 = %q, want empty", lines[1])
	}
	if want := "Disabling stops every route playing the input, including paused routes and routes sharing its hardware input, and removes an input registered with auto-remove."; lines[2] != want {
		t.Errorf("line 3 = %q, want %q", lines[2], want)
	}
}

func TestInputsRunDisable_UsageErrorOmitsDescription(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := inputs.RunDisable([]string{}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "usage: sonora disable inputs/<input-id> [flags]") {
		t.Errorf("expected the usage line on stderr, got:\n%s", stderr.String())
	}
	if strings.Contains(stderr.String(), "Disabling stops every route playing the input, including paused routes and routes sharing its hardware input, and removes an input registered with auto-remove.") {
		t.Errorf("expected no description on a usage error, got stderr:\n%s", stderr.String())
	}
}
