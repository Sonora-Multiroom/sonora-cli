package unit

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"sonora-cli/internal/cli/inputs"
)

func TestInputsRunDelete_Help(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := inputs.RunDelete([]string{"--help"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stdout: %s", code, stdout.String())
	}
	if !strings.Contains(stdout.String(), "Flags:") {
		t.Errorf("expected a Flags: section, got stdout:\n%s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "sonora delete inputs/<input-id>") {
		t.Errorf("expected usage line to name the delete grammar, got stdout:\n%s", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Errorf("expected help on stdout only, got stderr:\n%s", stderr.String())
	}
}

func TestInputsRunDelete_MissingIdentifier(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := inputs.RunDelete([]string{}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
}

func TestInputsRunDelete_TooManyArguments(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := inputs.RunDelete([]string{"a", "b"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
}

func TestInputsRunDelete_UnreachableHubURL(t *testing.T) {
	var stdout, stderr bytes.Buffer
	start := time.Now()
	code := inputs.RunDelete([]string{"spotify-1", "--hub-url", "http://127.0.0.1:1"}, &stdout, &stderr)
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

func TestInputsRunDelete_VerboseAppendsRawErrorDetail(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := inputs.RunDelete([]string{"spotify-1", "--hub-url", "http://127.0.0.1:1", "--verbose"}, &stdout, &stderr)

	if code != 4 {
		t.Fatalf("exit code = %d, want 4; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "detail:") {
		t.Errorf("expected --verbose to append raw error detail, got stderr:\n%s", stderr.String())
	}
}

func TestInputsRunDelete_NonVerboseOmitsRawErrorDetail(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := inputs.RunDelete([]string{"spotify-1", "--hub-url", "http://127.0.0.1:1"}, &stdout, &stderr)

	if code != 4 {
		t.Fatalf("exit code = %d, want 4; stderr: %s", code, stderr.String())
	}
	if strings.Contains(stderr.String(), "detail:") {
		t.Errorf("expected no raw error detail without --verbose, got stderr:\n%s", stderr.String())
	}
}
