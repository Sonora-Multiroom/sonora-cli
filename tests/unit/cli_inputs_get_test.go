package unit

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"sonora-cli/internal/cli/inputs"
)

func TestInputsRunGet_Help(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := inputs.RunGet([]string{"--help"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "Flags:") {
		t.Errorf("expected a Flags: section, got stderr:\n%s", stderr.String())
	}
}

func TestInputsRunGet_MissingIdentifier(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := inputs.RunGet([]string{}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
}

func TestInputsRunGet_TooManyArguments(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := inputs.RunGet([]string{"a", "b"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
}

func TestInputsRunGet_UnreachableHubURL(t *testing.T) {
	var stdout, stderr bytes.Buffer
	start := time.Now()
	code := inputs.RunGet([]string{"spotify-1", "--hub-url", "http://127.0.0.1:1"}, &stdout, &stderr)
	elapsed := time.Since(start)

	if code != 4 {
		t.Fatalf("exit code = %d, want 4; stderr: %s", code, stderr.String())
	}
	if elapsed >= 5*time.Second {
		t.Errorf("expected the failure to return well under 5s (SC-004), took %v", elapsed)
	}
}

func TestInputsRunGet_VerboseAppendsRawErrorDetail(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := inputs.RunGet([]string{"spotify-1", "--hub-url", "http://127.0.0.1:1", "--verbose"}, &stdout, &stderr)

	if code != 4 {
		t.Fatalf("exit code = %d, want 4; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "detail:") {
		t.Errorf("expected --verbose to append raw error detail, got stderr:\n%s", stderr.String())
	}
}

func TestInputsRunGet_NonVerboseOmitsRawErrorDetail(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := inputs.RunGet([]string{"spotify-1", "--hub-url", "http://127.0.0.1:1"}, &stdout, &stderr)

	if code != 4 {
		t.Fatalf("exit code = %d, want 4; stderr: %s", code, stderr.String())
	}
	if strings.Contains(stderr.String(), "detail:") {
		t.Errorf("expected no raw error detail without --verbose, got stderr:\n%s", stderr.String())
	}
}
