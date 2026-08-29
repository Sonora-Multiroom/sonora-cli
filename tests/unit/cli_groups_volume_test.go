package unit

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"sonora-cli/internal/cli/groups"
)

func TestGroupsRunSetVolume_MissingIdentifier(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := groups.RunSetVolume([]string{}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
}

func TestGroupsRunSetVolume_MissingAttributeAndValue(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := groups.RunSetVolume([]string{"downstairs"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
}

func TestGroupsRunSetVolume_WrongAttributeWord(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := groups.RunSetVolume([]string{"downstairs", "loudness", "75"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
}

func TestGroupsRunSetVolume_NonNumericValue(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := groups.RunSetVolume([]string{"downstairs", "volume", "loud"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
}

func TestGroupsRunSetVolume_OutOfRangeValue(t *testing.T) {
	// A negative value must reach the range check rather than flag parsing,
	// which would otherwise report "flag provided but not defined: -1" —
	// also exit 2, so the message is what distinguishes the two paths.
	for _, v := range []string{"-1", "-5", "101", "99999999999999999999"} {
		var stdout, stderr bytes.Buffer
		code := groups.RunSetVolume([]string{"downstairs", "volume", v}, &stdout, &stderr)

		if code != 2 {
			t.Fatalf("value=%s: exit code = %d, want 2; stderr: %s", v, code, stderr.String())
		}
		want := fmt.Sprintf("error: volume must be an integer between 0 and 100, got %q", v)
		if !strings.Contains(stderr.String(), want) {
			t.Errorf("value=%s: expected stderr to contain %q, got:\n%s", v, want, stderr.String())
		}
	}
}

func TestGroupsRunSetVolume_NegativeValueWithFlags(t *testing.T) {
	// Flags on either side of the positionals must not change how the
	// negative value is read.
	argSets := [][]string{
		{"downstairs", "volume", "-5", "--json"},
		{"downstairs", "volume", "-5", "--hub-url", "http://127.0.0.1:1"},
	}
	for _, args := range argSets {
		var stdout, stderr bytes.Buffer
		code := groups.RunSetVolume(args, &stdout, &stderr)

		if code != 2 {
			t.Fatalf("args=%v: exit code = %d, want 2; stderr: %s", args, code, stderr.String())
		}
		if !strings.Contains(stderr.String(), "volume must be an integer between 0 and 100") {
			t.Errorf("args=%v: expected the range error, got stderr:\n%s", args, stderr.String())
		}
	}
}

func TestGroupsRunSetVolume_TooManyArguments(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := groups.RunSetVolume([]string{"downstairs", "volume", "75", "extra"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
}

func TestGroupsRunSetVolume_UnreachableHubURL(t *testing.T) {
	var stdout, stderr bytes.Buffer
	start := time.Now()
	code := groups.RunSetVolume([]string{"downstairs", "volume", "75", "--hub-url", "http://127.0.0.1:1"}, &stdout, &stderr)
	elapsed := time.Since(start)

	if code != 4 {
		t.Fatalf("exit code = %d, want 4; stderr: %s", code, stderr.String())
	}
	if elapsed >= 5*time.Second {
		t.Errorf("expected the failure to return well under 5s, took %v", elapsed)
	}
}

func TestGroupsRunSetVolume_VerboseAppendsRawErrorDetail(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := groups.RunSetVolume([]string{"downstairs", "volume", "75", "--hub-url", "http://127.0.0.1:1", "--verbose"}, &stdout, &stderr)

	if code != 4 {
		t.Fatalf("exit code = %d, want 4; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "detail:") {
		t.Errorf("expected --verbose to append raw error detail, got stderr:\n%s", stderr.String())
	}
}
