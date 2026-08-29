package unit

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"sonora-cli/internal/cli/inputs"
)

func TestInputsRunCreate_Help(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := inputs.RunCreate([]string{"--help"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stdout: %s", code, stdout.String())
	}
	if !strings.Contains(stdout.String(), "Flags:") {
		t.Errorf("expected a Flags: section, got stdout:\n%s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "sonora create inputs/<input-id>") {
		t.Errorf("expected usage line to name the create grammar, got stdout:\n%s", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Errorf("expected help on stdout only, got stderr:\n%s", stderr.String())
	}
}

func TestInputsRunCreate_MissingBothArguments(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := inputs.RunCreate([]string{"--display-name", "Spotify"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "input-id") {
		t.Errorf("expected stderr to name the missing <input-id> argument, got: %s", stderr.String())
	}
}

func TestInputsRunCreate_MissingURI(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := inputs.RunCreate([]string{"spotify-1", "--display-name", "Spotify"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "uri") {
		t.Errorf("expected stderr to name the missing <uri> argument, got: %s", stderr.String())
	}
}

func TestInputsRunCreate_TooManyArguments(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := inputs.RunCreate([]string{"spotify-1", "u1", "extra", "--display-name", "Spotify"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
}

func TestInputsRunCreate_MissingDisplayName(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := inputs.RunCreate([]string{"spotify-1", "u1"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "display-name") {
		t.Errorf("expected stderr to name the missing --display-name flag, got: %s", stderr.String())
	}
}

func TestInputsRunCreate_UnknownFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := inputs.RunCreate([]string{"spotify-1", "u1", "--display-name", "Spotify", "--unknown-flag"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
}

func TestInputsRunCreate_UnreachableHubURL(t *testing.T) {
	var stdout, stderr bytes.Buffer
	start := time.Now()
	code := inputs.RunCreate([]string{"spotify-1", "u1", "--display-name", "Spotify", "--hub-url", "http://127.0.0.1:1"}, &stdout, &stderr)
	elapsed := time.Since(start)

	if code != 4 {
		t.Fatalf("exit code = %d, want 4; stderr: %s", code, stderr.String())
	}
	if elapsed >= 5*time.Second {
		t.Errorf("expected the failure to return well under 5s, took %v", elapsed)
	}
	if stdout.String() != "" {
		t.Errorf("expected empty stdout on failure, got:\n%s", stdout.String())
	}
}

func TestInputsRunCreate_VerboseAppendsRawErrorDetail(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := inputs.RunCreate([]string{"spotify-1", "u1", "--display-name", "Spotify", "--hub-url", "http://127.0.0.1:1", "--verbose"}, &stdout, &stderr)

	if code != 4 {
		t.Fatalf("exit code = %d, want 4; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "detail:") {
		t.Errorf("expected --verbose to append raw error detail, got stderr:\n%s", stderr.String())
	}
}
