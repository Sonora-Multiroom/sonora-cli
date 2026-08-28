package unit

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"sonora-cli/internal/cli/outputs"
)

func TestOutputsRunSetVolume_MissingIdentifier(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := outputs.RunSetVolume([]string{}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
}

func TestOutputsRunSetVolume_MissingAttributeAndValue(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := outputs.RunSetVolume([]string{"office-speaker"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
}

func TestOutputsRunSetVolume_WrongAttributeWord(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := outputs.RunSetVolume([]string{"office-speaker", "loudness", "75"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
}

func TestOutputsRunSetVolume_NonNumericValue(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := outputs.RunSetVolume([]string{"office-speaker", "volume", "loud"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
}

func TestOutputsRunSetVolume_OutOfRangeValue(t *testing.T) {
	for _, v := range []string{"-1", "101"} {
		var stdout, stderr bytes.Buffer
		code := outputs.RunSetVolume([]string{"office-speaker", "volume", v}, &stdout, &stderr)

		if code != 2 {
			t.Fatalf("value=%s: exit code = %d, want 2; stderr: %s", v, code, stderr.String())
		}
	}
}

func TestOutputsRunSetVolume_TooManyArguments(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := outputs.RunSetVolume([]string{"office-speaker", "volume", "75", "extra"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
}

func TestOutputsRunSetVolume_UnreachableHubURL(t *testing.T) {
	var stdout, stderr bytes.Buffer
	start := time.Now()
	code := outputs.RunSetVolume([]string{"office-speaker", "volume", "75", "--hub-url", "http://127.0.0.1:1"}, &stdout, &stderr)
	elapsed := time.Since(start)

	if code != 4 {
		t.Fatalf("exit code = %d, want 4; stderr: %s", code, stderr.String())
	}
	if elapsed >= 5*time.Second {
		t.Errorf("expected the failure to return well under 5s, took %v", elapsed)
	}
}

func TestOutputsRunSetVolume_VerboseAppendsRawErrorDetail(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := outputs.RunSetVolume([]string{"office-speaker", "volume", "75", "--hub-url", "http://127.0.0.1:1", "--verbose"}, &stdout, &stderr)

	if code != 4 {
		t.Fatalf("exit code = %d, want 4; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "detail:") {
		t.Errorf("expected --verbose to append raw error detail, got stderr:\n%s", stderr.String())
	}
}
