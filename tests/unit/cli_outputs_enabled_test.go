package unit

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/Sonora-Multiroom/sonora-cli/internal/cli/outputs"
)

func TestOutputsRunEnable_MissingIdentifier(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := outputs.RunEnable([]string{}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
}

func TestOutputsRunEnable_TooManyArguments(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := outputs.RunEnable([]string{"a", "b"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
}

func TestOutputsRunEnable_UnreachableHubURL(t *testing.T) {
	var stdout, stderr bytes.Buffer
	start := time.Now()
	code := outputs.RunEnable([]string{"office-speaker", "--hub-url", "http://127.0.0.1:1"}, &stdout, &stderr)
	elapsed := time.Since(start)

	if code != 4 {
		t.Fatalf("exit code = %d, want 4; stderr: %s", code, stderr.String())
	}
	if elapsed >= 5*time.Second {
		t.Errorf("expected the failure to return well under 5s, took %v", elapsed)
	}
}

func TestOutputsRunEnable_VerboseAppendsRawErrorDetail(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := outputs.RunEnable([]string{"office-speaker", "--hub-url", "http://127.0.0.1:1", "--verbose"}, &stdout, &stderr)

	if code != 4 {
		t.Fatalf("exit code = %d, want 4; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "detail:") {
		t.Errorf("expected --verbose to append raw error detail, got stderr:\n%s", stderr.String())
	}
}

func TestOutputsRunDisable_MissingIdentifier(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := outputs.RunDisable([]string{}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
}

func TestOutputsRunDisable_TooManyArguments(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := outputs.RunDisable([]string{"a", "b"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
}

func TestOutputsRunDisable_UnreachableHubURL(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := outputs.RunDisable([]string{"office-speaker", "--hub-url", "http://127.0.0.1:1"}, &stdout, &stderr)

	if code != 4 {
		t.Fatalf("exit code = %d, want 4; stderr: %s", code, stderr.String())
	}
}

func TestOutputsRunEnable_HelpDescribesPlayback(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := outputs.RunEnable([]string{"--help"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", code, stderr.String())
	}
	lines := strings.Split(stdout.String(), "\n")
	if len(lines) < 3 {
		t.Fatalf("expected at least 3 lines of help, got:\n%s", stdout.String())
	}
	if want := "usage: sonora enable outputs/<output-id> [flags]"; lines[0] != want {
		t.Errorf("line 1 = %q, want %q", lines[0], want)
	}
	if lines[1] != "" {
		t.Errorf("line 2 = %q, want empty", lines[1])
	}
	if want := "Enabling puts the output back into the group routes still playing on its groups; a route stopped by the disable does not come back."; lines[2] != want {
		t.Errorf("line 3 = %q, want %q", lines[2], want)
	}
}

func TestOutputsRunDisable_HelpDescribesPlayback(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := outputs.RunDisable([]string{"--help"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", code, stderr.String())
	}
	lines := strings.Split(stdout.String(), "\n")
	if len(lines) < 3 {
		t.Fatalf("expected at least 3 lines of help, got:\n%s", stdout.String())
	}
	if want := "usage: sonora disable outputs/<output-id> [flags]"; lines[0] != want {
		t.Errorf("line 1 = %q, want %q", lines[0], want)
	}
	if lines[1] != "" {
		t.Errorf("line 2 = %q, want empty", lines[1])
	}
	if want := "Disabling takes the output out of every route playing on it; a route addressed only to it stops."; lines[2] != want {
		t.Errorf("line 3 = %q, want %q", lines[2], want)
	}
}

func TestOutputsRunDisable_UsageErrorOmitsDescription(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := outputs.RunDisable([]string{}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "usage: sonora disable outputs/<output-id> [flags]") {
		t.Errorf("expected the usage line on stderr, got:\n%s", stderr.String())
	}
	if strings.Contains(stderr.String(), "Disabling takes the output out of every route playing on it; a route addressed only to it stops.") {
		t.Errorf("expected no description on a usage error, got stderr:\n%s", stderr.String())
	}
}
