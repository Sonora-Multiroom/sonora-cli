package unit

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/Sonora-Multiroom/sonora-cli/internal/cli/routes"
)

func TestRoutesRunSetVolume_UsageErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"missing identifier", []string{}, "error: missing required arguments: <route-id> volume <0-100>"},
		{"missing attribute and value", []string{"route-1"}, "error: missing required arguments: <route-id> volume <0-100>"},
		{"wrong attribute word", []string{"route-1", "loudness", "40"}, `error: unsupported attribute "loudness" for routes; only volume is supported`},
		{"non-numeric value", []string{"route-1", "volume", "loud"}, `error: volume must be an integer between 0 and 100, got "loud"`},
		{"too many arguments", []string{"route-1", "volume", "40", "extra"}, "error: unexpected argument(s): [extra]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := routes.RunSetVolume(tt.args, &stdout, &stderr)

			if code != 2 {
				t.Fatalf("exit code = %d, want 2; stderr: %s", code, stderr.String())
			}
			if first, _, _ := strings.Cut(stderr.String(), "\n"); first != "usage: sonora set routes/<route-id> volume <0-100> [flags]" {
				t.Errorf("expected the usage line first on stderr, got: %s", stderr.String())
			}
			if !strings.Contains(stderr.String(), tt.want) {
				t.Errorf("expected stderr to contain %q, got: %s", tt.want, stderr.String())
			}
			if stdout.Len() != 0 {
				t.Errorf("expected empty stdout, got: %s", stdout.String())
			}
		})
	}
}

func TestRoutesRunSetVolume_OutOfRangeValue(t *testing.T) {
	for _, v := range []string{"-1", "-5", "101", "99999999999999999999"} {
		var stdout, stderr bytes.Buffer
		code := routes.RunSetVolume([]string{"route-1", "volume", v, "--json"}, &stdout, &stderr)

		if code != 2 {
			t.Fatalf("value=%s: exit code = %d, want 2; stderr: %s", v, code, stderr.String())
		}
		want := fmt.Sprintf("error: volume must be an integer between 0 and 100, got %q", v)
		if !strings.Contains(stderr.String(), want) {
			t.Errorf("value=%s: expected stderr to contain %q, got:\n%s", v, want, stderr.String())
		}
	}
}

func TestRoutesRunSetVolume_UnreachableHubURL(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := routes.RunSetVolume([]string{"route-1", "volume", "40", "--hub-url", "http://127.0.0.1:1", "--verbose"}, &stdout, &stderr)

	if code != 4 {
		t.Fatalf("exit code = %d, want 4; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "detail:") {
		t.Errorf("expected --verbose to append raw error detail, got stderr:\n%s", stderr.String())
	}
}

func TestRoutesRunSetVolume_Help(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := routes.RunSetVolume([]string{"--help"}, &stdout, &stderr)

	if code != 0 {
		t.Fatalf("exit code = %d, want 0; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "usage: sonora set routes/<route-id> volume <0-100> [flags]") {
		t.Errorf("expected the usage line on stdout, got:\n%s", stdout.String())
	}
}
