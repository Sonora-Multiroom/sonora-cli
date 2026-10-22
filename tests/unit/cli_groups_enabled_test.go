package unit

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/Sonora-Multiroom/sonora-cli/internal/cli/groups"
)

func TestGroupsRunEnable_MissingIdentifier(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := groups.RunEnable([]string{}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
}

func TestGroupsRunEnable_TooManyArguments(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := groups.RunEnable([]string{"a", "b"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
}

func TestGroupsRunEnable_UnreachableHubURL(t *testing.T) {
	var stdout, stderr bytes.Buffer
	start := time.Now()
	code := groups.RunEnable([]string{"living-room", "--hub-url", "http://127.0.0.1:1"}, &stdout, &stderr)
	elapsed := time.Since(start)

	if code != 4 {
		t.Fatalf("exit code = %d, want 4; stderr: %s", code, stderr.String())
	}
	if elapsed >= 5*time.Second {
		t.Errorf("expected the failure to return well under 5s, took %v", elapsed)
	}
}

func TestGroupsRunEnable_VerboseAppendsRawErrorDetail(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := groups.RunEnable([]string{"living-room", "--hub-url", "http://127.0.0.1:1", "--verbose"}, &stdout, &stderr)

	if code != 4 {
		t.Fatalf("exit code = %d, want 4; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "detail:") {
		t.Errorf("expected --verbose to append raw error detail, got stderr:\n%s", stderr.String())
	}
}

func TestGroupsRunDisable_MissingIdentifier(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := groups.RunDisable([]string{}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
}

func TestGroupsRunDisable_TooManyArguments(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := groups.RunDisable([]string{"a", "b"}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
}

func TestGroupsRunDisable_UnreachableHubURL(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := groups.RunDisable([]string{"living-room", "--hub-url", "http://127.0.0.1:1"}, &stdout, &stderr)

	if code != 4 {
		t.Fatalf("exit code = %d, want 4; stderr: %s", code, stderr.String())
	}
}

func TestGroupsRunEnable_HelpDescribesPlayback(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := groups.RunEnable([]string{"--help"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", code, stderr.String())
	}
	lines := strings.Split(stdout.String(), "\n")
	if len(lines) < 3 {
		t.Fatalf("expected at least 3 lines of help, got:\n%s", stdout.String())
	}
	if want := "usage: sonora enable groups/<group-id> [flags]"; lines[0] != want {
		t.Errorf("line 1 = %q, want %q", lines[0], want)
	}
	if lines[1] != "" {
		t.Errorf("line 2 = %q, want empty", lines[1])
	}
	if want := "Enabling lets new routes use the group again; it starts nothing."; lines[2] != want {
		t.Errorf("line 3 = %q, want %q", lines[2], want)
	}
}

func TestGroupsRunDisable_HelpDescribesPlayback(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := groups.RunDisable([]string{"--help"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", code, stderr.String())
	}
	lines := strings.Split(stdout.String(), "\n")
	if len(lines) < 3 {
		t.Fatalf("expected at least 3 lines of help, got:\n%s", stdout.String())
	}
	if want := "usage: sonora disable groups/<group-id> [flags]"; lines[0] != want {
		t.Errorf("line 1 = %q, want %q", lines[0], want)
	}
	if lines[1] != "" {
		t.Errorf("line 2 = %q, want empty", lines[1])
	}
	if want := "Disabling stops every route addressed to the group; routes addressed to a member output keep playing."; lines[2] != want {
		t.Errorf("line 3 = %q, want %q", lines[2], want)
	}
}

func TestGroupsRunDisable_UsageErrorOmitsDescription(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := groups.RunDisable([]string{}, &stdout, &stderr)

	if code != 2 {
		t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "usage: sonora disable groups/<group-id> [flags]") {
		t.Errorf("expected the usage line on stderr, got:\n%s", stderr.String())
	}
	if strings.Contains(stderr.String(), "Disabling stops every route addressed to the group; routes addressed to a member output keep playing.") {
		t.Errorf("expected no description on a usage error, got stderr:\n%s", stderr.String())
	}
}
