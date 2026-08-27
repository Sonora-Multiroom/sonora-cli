package unit

import (
	"bytes"
	"strings"
	"testing"

	"sonora-cli/internal/cli/outputs"
)

func TestOutputsRunList_Help(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := outputs.RunList([]string{"--help"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "Flags:") {
		t.Errorf("expected a Flags: section, got stderr:\n%s", stderr.String())
	}
}

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

func TestOutputsRunList_HelpUsesNewGrammar(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := outputs.RunList([]string{"--help"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "sonora get|list outputs") {
		t.Errorf("expected usage line to name the new get/list grammar, got stderr:\n%s", stderr.String())
	}
	if strings.Contains(stderr.String(), "sonora outputs list") {
		t.Errorf("expected the removed old-grammar usage line to be gone, got stderr:\n%s", stderr.String())
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
