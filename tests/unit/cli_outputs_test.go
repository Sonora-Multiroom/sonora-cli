package unit

import (
	"bytes"
	"strings"
	"testing"

	"sonora-cli/internal/cli/outputs"
)

func TestOutputsRun_VerboseAppendsRawErrorDetail(t *testing.T) {
	var stdout, stderr bytes.Buffer
	// Nothing listens on port 1: a fast, deterministic connection failure.
	code := outputs.RunList([]string{"--hub-url", "http://127.0.0.1:1", "--verbose"}, &stdout, &stderr)

	if code != 4 {
		t.Fatalf("exit code = %d, want 4; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "detail:") {
		t.Errorf("expected --verbose to append raw error detail, got stderr:\n%s", stderr.String())
	}
}

func TestOutputsRun_NonVerboseOmitsRawErrorDetail(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := outputs.RunList([]string{"--hub-url", "http://127.0.0.1:1"}, &stdout, &stderr)

	if code != 4 {
		t.Fatalf("exit code = %d, want 4; stderr: %s", code, stderr.String())
	}
	if strings.Contains(stderr.String(), "detail:") {
		t.Errorf("expected no raw error detail without --verbose, got stderr:\n%s", stderr.String())
	}
}
